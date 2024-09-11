package panasonic

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const networkTimeout = 3 * time.Second

// Camera represent a remote camera to be controlled via the AW protocol
type Camera struct {
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
	httpOnce sync.Once
}

// httpInit sets good defaults of the Http client for the quirks in AW protocol
func (c *Camera) httpInit() {
	if c.Http.CheckRedirect == nil {
		c.Http.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	if c.Http.Timeout == 0 {
		c.Http.Timeout = networkTimeout
	}
}

// httpGet does an http.Get to the camera with the quirks of the AW protocol
func (c *Camera) httpGet(path string, query string) (*http.Response, error) {
	c.httpOnce.Do(c.httpInit)
	// The AW-RP50 just makes a one-liner HTTP/1.0 request, then proceeds to
	// provide a Host header anyway filled with an incorrectly zero-padded IP.
	// 	 GET /cgi-bin/aw_ptz?cmd=#R00&res=1 HTTP/1.0
	//   Host:198.051.100.008
	// We use a proper HTTP client instead.
	var host string
	if c.Remote.Port() == 0 {
		host = c.Remote.Addr().String()
	} else {
		host = c.Remote.String()
	}
	return c.Http.Do(&http.Request{
		Method: "GET",
		URL: &url.URL{
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

// strCommand sends a command string to the camera over the http transport
func (c *Camera) strCommand(cmd string) (string, error) {
	var path string

	// "guess" the endpoint based on the first character of the command
	if cmd[0] == '#' {
		path = "/cgi-bin/aw_ptz"
	} else {
		path = "/cgi-bin/aw_cam"
	}

	// Panasonic panels do NOT urlencode the command even though it contains #
	// Since the specification permits encoding, we do it for http compliance.
	res, err := c.httpGet(path, "cmd="+url.QueryEscape(cmd)+"&res=1")
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
func (c *Camera) AWCommand(req AWRequest) (AWResponse, error) {
	cmd := req.packRequest()

	ret, err := c.strCommand(cmd)
	if err != nil {
		return nil, &SystemError{err}
	}

	res := req.Response()
	if sig := res.responseSignature(); !match(sig, ret) {
		res = newResponse(ret)
	}

	res.unpackResponse(ret)

	if err, ok := res.(*AWError); ok {
		return nil, err
	}
	return res, nil
}

// AWBatch returns the command responses available at the camdata.html page.
func (c *Camera) AWBatch() ([]AWResponse, error) {
	data, err := c.httpGet("/live/camdata.html", "")
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
		res = append(res, newResponse(scan.Text()))
	}
	if err := scan.Err(); err != nil {
		return nil, &SystemError{err}
	}
	return res, nil
}

// Listener returns a listener for the AW notification protocol
//
// The returned listener already has an open a TCP listening port and ready
// to accept notifications.
func (c *Camera) Listener() (*NotifyListener, error) {
	listener, err := net.ListenTCP("tcp4", &net.TCPAddr{})
	if err != nil {
		return nil, &SystemError{err}
	}
	return &NotifyListener{
		lis: listener,
		cam: c,
	}, nil
}

var _ AWHandler = (*Camera)(nil)

type AWHandler interface {
	AWCommand(AWRequest) (AWResponse, error)
	AWBatch() ([]AWResponse, error)
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
	c.wrapAW(true, w, r)
}

// serveCam is the /cgi-bin/aw_cam endpoint handler
func (c *CameraServer) serveCam(w http.ResponseWriter, r *http.Request) {
	c.wrapAW(false, w, r)
}

// wrapAW does the http dance around AW commands
func (c *CameraServer) wrapAW(hash bool, w http.ResponseWriter, r *http.Request) {
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
	// Generate a "Bad Request" for confused endpoints
	if (strcmd[0] == '#') != hash {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	awcmd := newRequest(strcmd)
	awres, err := c.AWHandler.AWCommand(awcmd)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
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
	b, err := c.AWHandler.AWBatch()
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
	}
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	for _, r := range b {
		w.Write([]byte(r.packResponse()))
		w.Write([]byte("\r\n"))
	}
}

var _ http.Handler = (*CameraServer)(nil)
