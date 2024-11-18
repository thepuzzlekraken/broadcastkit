package panasonic

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const networkTimeout = 3 * time.Second

// CameraClient represent a remote camera to be controlled via the AW protocol
type CameraClient struct {
	// Remote is the IP address and port of the remote camera
	// If the port is 0, the default port 80 is used.
	Remote netip.AddrPort
	// Http is the HTTP client to use for requests
	//
	// Some variables are adjusted for Panasonic protocol unless explicitly set
	// before the first request is sent.
	// CheckRedirects is set func(...) error { return http.ErrUseLastResponse }
	// Timeout is set to 3 seconds
	Http     http.Client
	httpOnce sync.Once     // track http initialization
	dummyCtr atomic.Uint64 // source for dummy cache-disabling numbers
}

// httpInit sets good defaults of the Http client for the quirks in AW protocol
func (c *CameraClient) httpInit() {
	if c.Http.CheckRedirect == nil {
		c.Http.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	if c.Http.Timeout == 0 {
		c.Http.Timeout = networkTimeout
	}
	c.dummyCtr.Store(rand.Uint64())
}

// httpGet does an http.Get to the camera with the quirks of the AW protocol
func (c *CameraClient) httpGet(path string, query string, user *url.Userinfo) (*http.Response, error) {
	c.httpOnce.Do(c.httpInit)
	// The AW-RP50 just makes a one-liner HTTP/1.0 request, then proceeds to
	// provide a Host header anyway filled with an incorrectly zero-padded IP.
	// 	 GET /cgi-bin/aw_ptz?cmd=#R00&res=1 HTTP/1.0
	//   Host:198.051.100.008
	// We use a proper HTTP client instead.
	var host string
	if port := c.Remote.Port(); port == 0 || port == 80 {
		host = c.Remote.Addr().String()
	} else {
		host = c.Remote.String()
	}
	return c.Http.Do(&http.Request{
		Method: "GET",
		URL: &url.URL{
			User:     user,
			Scheme:   "http",
			Host:     host,
			Path:     path,
			RawQuery: query,
		},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Body:       nil,
		Host:       host,
	})
}

func guessQuirks(cmd string) quirkMode {
	// "guess" the endpoint based on the first character of the command
	if cmd[0] == '#' {
		return quirkPtz
	} else {
		return quirkCamera
	}
}

// strCommand sends a command string to the camera over the http transport
func (c *CameraClient) strCommand(cmd string) (string, error) {
	var path string

	// "guess" the endpoint based on the first character of the command
	switch guessQuirks(cmd) {
	case quirkPtz:
		path = "/cgi-bin/aw_ptz"
	case quirkCamera:
		path = "/cgi-bin/aw_cam"
	default:
		panic("unknown command quirks")
	}

	// Panasonic panels do NOT urlencode the command even though it contains #
	// Since the specification permits encoding, we do it for http compliance.
	res, err := c.httpGet(path, "cmd="+url.QueryEscape(cmd)+"&res=1", nil)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http status code: %d (expected %d)", res.StatusCode, http.StatusOK)
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(trim(b)), nil
}

// AWCommand sends the passed AWRequest to the camera
//
// AW protocol error responses are returned as errors, not AWResponse objects.
func (c *CameraClient) AWCommand(req AWRequest) (AWResponse, error) {
	cmd := req.packRequest()

	ret, err := c.strCommand(cmd)
	if err != nil {
		return nil, &SystemError{err}
	}

	res := req.Response()
	if sig := res.responseSignature(); match(sig, ret) {
		res.unpackResponse(ret)
	} else {
		res = newResponse(ret, guessQuirks(cmd))
	}

	if err, ok := res.(AWError); ok {
		return nil, err
	}
	return res, nil
}

// AWBatch returns the command responses available at the camdata.html page.
func (c *CameraClient) AWBatch() ([]AWResponse, error) {
	data, err := c.httpGet("/live/camdata.html", "", nil)
	if err != nil {
		return nil, &SystemError{err}
	}
	defer data.Body.Close()
	if data.StatusCode != http.StatusOK {
		return nil, &SystemError{fmt.Errorf("http status code: %d (expected %d)", data.StatusCode, http.StatusOK)}
	}
	scan := bufio.NewScanner(data.Body)
	res := make([]AWResponse, 0)
	for scan.Scan() {
		res = append(res, newResponse(scan.Text(), quirkBatch))
	}
	if err := scan.Err(); err != nil {
		return nil, &SystemError{err}
	}
	return res, nil
}

// Screenshot returns a JPEG-encoded still image from the camera
//
// You can specify an image width in pixels as resolution, which may be honored
// on a best-effort basis. The returned image size may be different.
func (c *CameraClient) Screenshot(resolution int) ([]byte, error) {
	// The page value is defined by the documentation to defeat caches. Probably not necessary in practice.
	// httpInit is called here to ensure dummyCtr is initialized to a pseudo-random value.
	c.httpOnce.Do(c.httpInit)
	query := "resolution=" + strconv.Itoa(resolution) + "&page=" + strconv.FormatUint(c.dummyCtr.Add(1), 10)

	data, err := c.httpGet("/cgi-bin/camera", query, nil)
	if err != nil {
		return nil, &SystemError{err}
	}
	defer data.Body.Close()

	if data.StatusCode != http.StatusOK {
		return nil, &SystemError{fmt.Errorf("http status code: %d (expected %d)", data.StatusCode, http.StatusOK)}
	}

	if ct := data.Header.Get("Content-Type"); ct != "image/jpeg" {
		return nil, &SystemError{fmt.Errorf("unexpected content type: %s", ct)}
	}

	img, err := io.ReadAll(data.Body)
	if err != nil {
		return nil, &SystemError{err}
	}

	return img, nil
}

func (c *CameraClient) GetTitle() (string, error) {
	// Using the AWBatch endpoint because it does not require authentication.
	batch, err := c.AWBatch()
	if err != nil {
		return "", err
	}
	for _, res := range batch {
		if t, ok := res.(AWTitle); ok {
			return t.Title, nil
		}
	}
	return "", fmt.Errorf("camera did not report title")
}

var defaultUserPassword = url.UserPassword("admin", "12345")

func (c *CameraClient) SetTitle(title string, user *url.Userinfo) error {
	if user == nil {
		user = defaultUserPassword
	}
	res, err := c.httpGet("/cgi-bin/set_basic", "cam_title="+url.QueryEscape(title), user)
	if err != nil {
		return &SystemError{err}
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return &SystemError{fmt.Errorf("http status code: %d (expected %d)", res.StatusCode, http.StatusOK)}
	}
	compBuf := make([]byte, 10+len(title)+2)
	copy(compBuf, "cam_title=")
	copy(compBuf[10:], title)
	copy(compBuf[10+len(title):], "\r\n")
	checkBuf := make([]byte, len(compBuf))
	_, err = res.Body.Read(checkBuf)
	if err != nil {
		return &SystemError{err}
	}
	if !bytes.Equal(compBuf, checkBuf) {
		return fmt.Errorf("unexpected response from camera: %s", string(checkBuf))
	}
	return nil
}

// Listener returns a listener for the AW notification protocol
//
// The returned listener already has an open a TCP listening port and ready
// to accept notifications.
func (c *CameraClient) Listener() (*NotifyListener, error) {
	listener, err := net.ListenTCP("tcp4", &net.TCPAddr{})
	if err != nil {
		return nil, &SystemError{err}
	}
	return &NotifyListener{
		lis: listener,
		cam: c,
	}, nil
}

var _ AWHandler = (*CameraClient)(nil)

type AWHandler interface {
	AWCommand(AWRequest) (AWResponse, error)
	AWBatch() ([]AWResponse, error)
}
type AWHandlerCtx interface {
	AWCommandCtx(context.Context, AWRequest) (AWResponse, error)
	AWBatchCtx(context.Context) ([]AWResponse, error)
}

// CameraServer is an http.Handler that implements an endpoint for AW protocol.
//
// This can be used to receive AW protocol commands from a remote panel acting
// as a camera. This handler should be registered for the routes of:
// - /cgi-bin/aw_ptz
// - /cgi-bin/aw_cam
// - /cgi-bin/event
// - /cgi-bin/man_session
// - /live/camdata.html
//
// The user should provide the AWHandler which will be called to handle received
// requests. Camera implements AWHandler.
//
// Notification subscribers are maintained automatically, but it is the task of
// the user to send out notifications. See the NotifyForward function to forward
// notifications from a Camera.
type CameraServer struct {
	once      sync.Once
	mux       http.ServeMux
	AWHandler AWHandler
	Notify    NotifyServer
}

// setup initializes the CameraServer
func (c *CameraServer) setup() {
	c.mux = http.ServeMux{}
	c.mux.HandleFunc("/cgi-bin/aw_ptz", c.servePtz)
	c.mux.HandleFunc("/cgi-bin/aw_cam", c.serveCam)
	c.mux.HandleFunc("/cgi-bin/event", c.serveEvent)
	c.mux.HandleFunc("/cgi-bin/man_session", c.serveManSession)
	c.mux.HandleFunc("/live/camdata.html", c.serveCamData)
}

// ServeHTTP implements the http.Handler interface
func (c *CameraServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// There's no documentation about the error conditions that happen when
	// the AW Protocol is violated. We mostly follow the http spec instead,
	// returning 400/500 for anything wrong with the request and the server
	// respectively.
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("panic: %v\n", r)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}()
	// There are some undocumented APIs behind other HTTP methods in real
	// devices. We opt to not support them and return a proper 405 instead.
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// The spec appendix insists on all responses having no-cache directives.
	w.Header().Add("Cache-Control", "no-cache")
	c.once.Do(c.setup)
	c.mux.ServeHTTP(w, r)
}

// servePtz is the /cgi-bin/aw_ptz endpoint handler
func (c *CameraServer) servePtz(w http.ResponseWriter, r *http.Request) {
	c.wrapAW(quirkPtz, w, r)
}

// serveCam is the /cgi-bin/aw_cam endpoint handler
func (c *CameraServer) serveCam(w http.ResponseWriter, r *http.Request) {
	c.wrapAW(quirkCamera, w, r)
}

// wrapAW does the http dance around AW commands
func (c *CameraServer) wrapAW(mode quirkMode, w http.ResponseWriter, r *http.Request) {
	// Generate a "Bad Request" for missing parameters
	qry := r.URL.Query()
	if qry.Get("res") != "1" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	strcmd := qry.Get("cmd")
	if len(strcmd) < 1 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	// Generate a "Bad Request" for confused endpoints (only quirkPtz has #)
	if (strcmd[0] == '#') != (mode == quirkPtz) {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	awcmd := newRequest(strcmd)
	var awres AWResponse
	var err error
	if ctxhandler, ok := c.AWHandler.(AWHandlerCtx); ok {
		awres, err = ctxhandler.AWCommandCtx(r.Context(), awcmd)
	} else {
		awres, err = c.AWHandler.AWCommand(awcmd)
	}
	if errres, ok := err.(AWError); ok {
		awres = errres
		err = nil
	}
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	if q, ok := awres.(awQuirkedPacking); ok {
		awres = q.packingQuirk(mode)
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(awres.packResponse()))
}

// serveEvent is the /cgi-bin/event endpoint handler
func (c *CameraServer) serveEvent(w http.ResponseWriter, r *http.Request) {
	// If the connect or my_port values are bad, we will respond with 400
	// This does not match the Panasonic HW behavior, but we hate silent errors.
	qry := r.URL.Query()
	connect := qry.Get("connect")
	port := qry.Get("my_port")
	if (connect != "start" && connect != "stop") || port == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	portNo, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	// There's also a uid parameter:
	// Documentation states uid=0 constant. AW-RP50 sets it UID=50 constant.
	// AW-UE70 camera seems to ignore it. We also ignore it.

	ip := netip.MustParseAddrPort(r.RemoteAddr).Addr()
	client := netip.AddrPortFrom(ip, uint16(portNo))

	switch connect {
	case "start":
		c.Notify.Add(client)
	case "stop":
		c.Notify.Remove(client)
	default:
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// serveManSession is the /cgi-bin/man_session endpoint handler
func (c *CameraServer) serveManSession(w http.ResponseWriter, r *http.Request) {
	// The command=get parameter is required. We opt to provide a 400 instead of
	// the random mix of 204/403 codes observed in real devices.
	if r.URL.Query().Get("command") != "get" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	// Quick workaround to make AW-RP50 think it is connected to us.
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Event session:"))
	w.Write([]byte(strconv.Itoa(c.Notify.Len())))
}

// serveCamData is the /live/camdata.html endpoint handler
func (c *CameraServer) serveCamData(w http.ResponseWriter, r *http.Request) {
	var b []AWResponse
	var err error
	if ctxhandler, ok := c.AWHandler.(AWHandlerCtx); ok {
		b, err = ctxhandler.AWBatchCtx(r.Context())
	} else {
		b, err = c.AWHandler.AWBatch()
	}
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	for _, res := range b {
		if q, ok := res.(awQuirkedPacking); ok {
			res = q.packingQuirk(quirkBatch)
		}
		w.Write([]byte(res.packResponse()))
		w.Write([]byte("\r\n"))
	}
}

var _ http.Handler = (*CameraServer)(nil)
