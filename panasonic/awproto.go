package panasonic

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// awHint is a bitmask marking where this command is expected to appear.
//
// Because the Panasonic commands textual representation is not unique, this
// is used to determine look-up tables and endpoints.
type awHint int

const (
	awPtz       awHint = 1 << iota // expected over aw_ptz interface
	awCam                          // expected over aw_cam interface
	awNty                          // expected over notifications interface
	awHintCount awHint = 3
)

const networkTimeout = 3 * time.Second

// AWResponse is the interface implemented by all responses sent from a camera.
//
// To confirm success of an operation, the Ok() method can be used. For any
// further information, the application has to type-assert the response to the
// specific implementation.
type AWResponse interface {
	// Ok returns whether this response represents an error condition.
	//
	// The return value follows the Panasonic behavior, which is rarely useful
	// beyond the basic check for command acceptance.
	Ok() bool
	// responseSignature returns hint and pattern of Panasonic string literals.
	//
	// Pattern is in a custom format of this package. The match() function
	// should be used to test the pattern on any string literals before passing
	// them to unpackResponse().
	responseSignature() (awHint, string)
	// unpackResponse parses response values from Panasonic string literal.
	//
	// It is the responsibility of the caller to ensure pattern match before
	// a call. Behavior is undefined if the passed string does not match the
	// signature possibly resulting in panic or invalid command data!
	unpackResponse(string)
	// packResponse returns the Panasonic string representation of the response.
	//
	// The returned string is guaranteed to match the responseSignature pattern.
	packResponse() string
}

// AWRequest is the interface implemented by all commands sent to a camera
// To process an AWRequest other than proxying, the application should
// type-assert it to the specific implementation.
type AWRequest interface {
	// Acceptable returns whether the values of this command are within their
	// acceptable range.
	//
	// This function aims to follow the Panasonic behavior closely, returning
	// false exactly when an AWErrUnacceptable would be returned by a device.
	Acceptable() bool
	// Response returns an AWResponse object that the requeset is expected to be
	// replied with.
	//
	// This function often returns the receiver object itself, request should be
	// processed before any changes are made to the returned response.
	//
	// The actual response returned by a device may be of a different type.
	Response() AWResponse
	// requestSignature returns the hint and pattern of Panasonic string.
	//
	// Limitations are the same as AWResponse.responseSignature()
	requestSignature() (awHint, string)
	// unpackRequest parses request values from Panasonic string.
	//
	// The passed string must match the requestSignature() pattern. Limitations
	// are the same as AWResponse.unpackResponse().
	unpackRequest(string)
	// packRequest returns the Panasonic string representation of the request.
	//
	// The returned string is guaranteed to match the resquestSignature()
	// pattern, but it may still be invalid semantically.
	packRequest() string
}

// AWUknownResponse is a placeholder implementation for AWResponse.
//
// Used when a non-error response is not recognized by this library. It is not
// possible to understand the meaning of such replies. They are intended for
// proxying only.
type AWUnknownResponse struct {
	hint awHint
	text string
}

func (a *AWUnknownResponse) Ok() bool {
	return false
}
func (a *AWUnknownResponse) responseSignature() (awHint, string) {
	return a.hint, a.text
}
func (a *AWUnknownResponse) unpackResponse(_ string) {}
func (a *AWUnknownResponse) packResponse() string {
	return a.text
}

// AWUknownRequest is a placeholder implementation for AWRequest.
//
// Used when a request is not recognized by this library. It is not possible to
// understand the meaning of such requests. Their intended use is proxying or to
// be replied with AWErrUnsupported.
type AWUnknownRequest struct {
	hint awHint
	text string
}

func (a *AWUnknownRequest) Acceptable() bool {
	return true
}
func (a *AWUnknownRequest) Response() AWResponse {
	// Implementation note: AWUnknownRequest and AWUnknownResponse are separate
	// struct to avoid applications unknowingly casting them to the other type.
	return &AWUnknownResponse{}
}
func (a *AWUnknownRequest) requestSignature() (awHint, string) {
	return a.hint, a.text
}
func (a *AWUnknownRequest) unpackRequest(_ string) {}
func (a *AWUnknownRequest) packRequest() string {
	return a.text
}

// awRequestFactory is the lookup table entry for making AWRequest objects
type awRequestFactory struct {
	sig string
	new func() AWRequest
}

// awResponseFactory is the lookup table entry for making AWResponse objects
type awResponseFactory struct {
	sig string
	new func() AWResponse
}

// awRequestTable is the factory lookup table for AWRequests
var awRequestTable = [awHintCount][]awRequestFactory{}

// awResponseTable is the factory lookup table for AWResponses
var awResponseTable = [awHintCount][]awResponseFactory{}

// registerRequest registers a new request type with the factory table
func registerRequest(new func() AWRequest) {
	// TODO(zsh): These functions may be optimized away by code-generation instead.
	n := new()
	f, p := n.requestSignature()
	for i := range awHintCount {
		m := 1 << i
		if int(f)&m != 0 {
			awRequestTable[i] = append(awRequestTable[i], awRequestFactory{p, new})
		}
	}
}

// registerResponse registers a new response type with the factory table
func registerResponse(new func() AWResponse) {
	// TODO(zsh): These functions may be optimized away by code-generation instead.
	n := new()
	f, p := n.responseSignature()
	for i := range awHintCount {
		m := 1 << i
		if int(f)&m != 0 {
			awResponseTable[i] = append(awResponseTable[i], awResponseFactory{p, new})
		}
	}
}

// newRequest creates a new request via the factory
func newRequest(hint awHint, cmd string) AWRequest {
	// This function is within a latency-critical path of incoming requests.
	// Following is a tight-loop with everything inlined, but this may need
	// optimization if latency becomes an issue.
	for i := range awHintCount {
		m := 1 << i
		if int(hint)&m == 0 {
			continue
		}
		for _, e := range awRequestTable[i] {
			if match(e.sig, cmd) {
				req := e.new()
				req.unpackRequest(cmd)
				return req
			}
		}
	}
	return &AWUnknownRequest{
		hint: hint,
		text: cmd,
	}
}

// newResponse creates a new response via the factory
func newResponse(hint awHint, cmd string) AWResponse {
	// This function is less critical than newRequest(), because the object
	// returned by AWRequest.Response() is used in happy-path response creation.
	for i := range awHintCount {
		m := 1 << i
		if int(hint)&m == 0 {
			continue
		}
		debug := awResponseTable
		_ = debug
		for _, e := range awResponseTable[i] {
			if match(e.sig, cmd) {
				res := e.new()
				res.unpackResponse(cmd)
				return res
			}
		}
	}
	return &AWUnknownResponse{
		hint: hint,
		text: cmd,
	}
}

// NewAWError creates an AWError as a Panasonic device would.
// This is intended for simulating errors as a proxy or virtual device.
func NewAWError(n AWErrNo, c AWRequest) *AWError {
	f, _ := c.requestSignature()
	t := c.packRequest()
	return &AWError{
		cap:  (f & awPtz) == 0,
		No:   n,
		Flag: t[:min(len(t), 3)],
	}
}

// NetworkError is an error condition outside of the Panasonic protocol
type NetworkError struct {
	parent error
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("panasonic network failure: %s", e.parent.Error())
}
func (e *NetworkError) Unwrap() error {
	return e.parent
}

// CameraRemote represent a remote camera to be controlled via the AW protocol
type CameraRemote struct {
	// Addr is the IP address of the camera
	Addr netip.Addr
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
func (c *CameraRemote) httpInit() {
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
func (c *CameraRemote) httpGet(path string, query string) (*http.Response, error) {
	c.httpOnce.Do(c.httpInit)
	// The AW-RP50 just makes a one-liner HTTP/1.0 request, then proceeds to
	// provide a Host header anyway filled with an incorrectly zero-padded IP.
	// 	 GET /cgi-bin/aw_ptz?cmd=#R00&res=1 HTTP/1.0
	//   Host:198.051.100.008
	// We use a proper HTTP client instead.
	return c.Http.Do(&http.Request{
		Method: "GET",
		URL: &url.URL{
			Scheme:   "http",
			Host:     c.Addr.String(),
			Path:     path,
			RawQuery: query,
		},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Body:       nil,
		Host:       c.Addr.String(),
	})
}

// strCommand sends a command string to the camera over the http transport
func (c *CameraRemote) strCommand(hint awHint, cmd string) (string, error) {
	var path string

	if hint&awPtz != 0 {
		path = "/cgi-bin/aw_ptz"
	} else if hint&awCam != 0 {
		path = "/cgi-bin/aw_cam"
	} else {
		return "", fmt.Errorf("unsupported command on AW interface: %v", cmd)
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

func (c *CameraRemote) Command(req AWRequest) (AWResponse, error) {
	hint, _ := req.requestSignature()
	cmd := req.packRequest()

	ret, err := c.strCommand(hint, cmd)
	if err != nil {
		return nil, &NetworkError{err}
	}

	res := req.Response()
	_, sig := res.responseSignature()
	if match(sig, ret) {
		res.unpackResponse(ret)
		return res, nil
	}

	res = newResponse(hint, ret)

	if err, ok := res.(error); ok {
		return res, err
	}

	return res, nil
}

func (c *CameraRemote) NotificationListener() (*CameraNotifyListener, error) {
	listener, err := net.ListenTCP("tcp4", &net.TCPAddr{})
	if err != nil {
		return nil, &NetworkError{err}
	}
	return &CameraNotifyListener{
		lis: listener,
		cam: c,
	}, nil
}

// notifyUnpack retrieves a string response from the notification container
//
// This function ignores the undocumented metadata within the container.
// Format is:
// - 22 bytes reserved
// -  4 bytes of length, encoded as (l + 8) in uint16 big endian
// -  l bytes of string data (surrounded by CRLF)
// - 24 bytes reserved
//
// It is observed that string data sometimes have trailing null bytes. This
// function trims the string to printable ASCII characters.
func notifyUnpack(b []byte) (string, error) {
	if len(b) < 24 {
		return "", fmt.Errorf("panasonic.notifyUnpack: data too short %d (expected >=24)", len(b))
	}
	var l uint16
	_, err := binary.Decode(b[22:24], binary.BigEndian, &l)
	if err != nil {
		return "", fmt.Errorf("panasonic.tcpUnwrap: invalid length data: %w", err)
	}
	l -= 8
	dl := (22 + 2 + 4 + int(l) + 24)
	if len(b) != dl {
		return "", fmt.Errorf("panasonic.tcpUnwrap: data length mismatch %d (expected %d)", len(b), dl)
	}

	return string(trim(b[30 : 30+l])), nil
}

// notifyPack packages a string response into the notification container
//
// In addition to the documented format, this function guesses the undocumented
// bytes. Format is:
// -  4 bytes source IP
// -  2 bytes counter, which starts at 1, incremented for each packet
// -  6 bytes date, encoded as uint8 each: YEAR (last two decimal digits), MONTH, DAY, HOUR, MINUTE, SECOND
// - 10 bytes constant, hex 00 01 00 80 00 00 00 00 00 01
// -  2 bytes length, encoded as (l + 8) in uint16 big endian
// -  4 bytes constant, hex 01 00 00 00
// -  l bytes string data (surrounded by 2-2 bytes of CRLF)
// -  4 bytes constant, hex 00 02 00 18
// -  6 bytes source MAC
// -  2 bytes constant, hex 00 01
// -  6 bytes date, same as above
// -  6 bytes constant 00 00 00 00 00 00
// Date is in local time of the camera, not UTC.
func notifyPack(response string, session *NotifySession, date time.Time) []byte {
	offset := 30 + len(response)
	length := offset + 26

	b := make([]byte, length)
	_ = b[offset+25]
	_ = b[30]

	srcip := session.srcIP
	if srcip == (netip.Addr{}) {
		// avoid panicing on uninitialized sessions
		srcip = netip.IPv4Unspecified()
	}
	ipv4 := srcip.As4()
	copy(b, ipv4[:])

	counter := session.Counter.next()
	binary.BigEndian.PutUint16(b[4:], counter)

	b[6] = uint8(date.Year() % 100)
	b[7] = uint8(date.Month())
	b[8] = uint8(date.Day())
	b[9] = uint8(date.Hour())
	b[10] = uint8(date.Minute())
	b[11] = uint8(date.Second())

	copy(b[12:], "\x00\x01\x00\x80\x00\x00\x00\x00\x00\x01")

	replen := len(response) + 4 + 8 // reported length includes CRLFs + 8 bytes
	binary.BigEndian.PutUint16(b[22:], uint16(replen))

	copy(b[24:], "\x01\x00\x00\x00\r\n")
	copy(b[30:], response)
	copy(b[offset:], "\r\n\x00\x02\x00\x18")

	copy(b[offset+6:], session.srcMac)

	copy(b[offset+12:], "\x00\x01")
	copy(b[offset+14:], b[6:12]) // copy date instead of recalculating it
	copy(b[offset+20:], "\x00\x00\x00\x00\x00\x00")

	return b
}

type CameraNotifyListener struct {
	once sync.Once
	lis  *net.TCPListener
	cam  *CameraRemote
}

func (l *CameraNotifyListener) Start() error {
	err := l.start()
	l.once.Do(func() {})
	return err
}

func (l *CameraNotifyListener) start() error {
	port := netip.MustParseAddrPort(l.lis.Addr().String()).Port()
	res, err := l.cam.httpGet("/cgi-bin/event", "connect=start&my_port="+strconv.Itoa(int(port))+"&uid=0")
	if err != nil {
		return &NetworkError{err}
	}
	if res.StatusCode != http.StatusNoContent {
		return &NetworkError{fmt.Errorf("http status code: %d (expected %d)", res.StatusCode, http.StatusNoContent)}
	}
	return nil
}

func (l *CameraNotifyListener) Stop() error {
	port := netip.MustParseAddrPort(l.lis.Addr().String()).Port()
	res, err := l.cam.httpGet("/cgi-bin/event", "connect=start&my_port="+strconv.Itoa(int(port))+"&uid=0")
	if err != nil {
		return &NetworkError{err}
	}
	if res.StatusCode != http.StatusNoContent {
		return &NetworkError{fmt.Errorf("http status code: %d (expected %d)", res.StatusCode, http.StatusNoContent)}
	}
	return nil
}

func (l *CameraNotifyListener) Addr() netip.AddrPort {
	return netip.MustParseAddrPort(l.lis.Addr().String())
}

func (l *CameraNotifyListener) acceptTCP() (*net.TCPConn, error) {
	camaddr := l.cam.Addr
	for {
		conn, err := l.lis.AcceptTCP()
		if err != nil {
			return nil, err
		}

		remaddr := netip.MustParseAddrPort(conn.RemoteAddr().String()).Addr()
		if remaddr == camaddr {
			return conn, nil
		}

		conn.Close()
	}
}

func (l *CameraNotifyListener) Accept() (AWResponse, error) {
	l.once.Do(func() { l.start() })

	conn, err := l.acceptTCP()
	if err != nil {
		return nil, &NetworkError{err}
	}

	conn.SetDeadline(time.Now().Add(networkTimeout))
	b, err := io.ReadAll(conn)
	conn.Close()

	if err != nil {
		return nil, &NetworkError{err}
	}

	cmd, err := notifyUnpack(b)
	if err != nil {
		return nil, &NetworkError{err}
	}

	return newResponse(awNty, cmd), nil
}

func (l *CameraNotifyListener) Close() error {
	_ = l.Stop()
	return l.lis.Close()
}

func sendTCP(b []byte, dst netip.AddrPort) error {
	conn, err := net.DialTimeout("tcp4", dst.String(), networkTimeout)
	if err != nil {
		return err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(networkTimeout))
	_, err = conn.Write(b)
	if err != nil {
		return err
	}
	return nil
}

// discoverMac returns the MAC address which would be used on the local network
// with the given source IP. This may differ from the actual src MAC if the
// packet is going through a gateway.
func discoverMac(ip netip.Addr) net.HardwareAddr {
	interfaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	for _, i := range interfaces {
		addrs, err := i.Addrs()
		if err != nil {
			panic(err)
		}
		for _, a := range addrs {
			t := netip.MustParsePrefix(a.String())
			if t.Contains(ip) {
				return i.HardwareAddr
			}
		}
	}
	return nil
}

// discoverSource returns the IP address that would be autoselected for outgoing
// connections to the provided destination.
func discoverSource(dst netip.Addr) netip.Addr {
	// Dialing incorrectly is not the most elegant thing to do, but we can avoid
	// sending actual packets by using UDP instead of TCP for address discovery.
	c, err := net.DialUDP("udp4", nil, net.UDPAddrFromAddrPort(
		netip.AddrPortFrom(dst, 1),
	))
	if err != nil {
		panic(err)
	}
	defer c.Close()
	return netip.MustParseAddrPort(c.LocalAddr().String()).Addr()
}

// NotifyCounter is a 16-bit unsigned thread-safe counter for keeping track of
// notification packet numbers.
//
// Implementation uses atomic.Uint32 as there's no atomic.Uint16. Applications
// should never access the underlying value directly.
type NotifyCounter atomic.Uint32

// next yields the counter value for the next notification packet
func (c *NotifyCounter) next() uint16 {
	return uint16((*atomic.Uint32)(c).Add(1))
}

// Load returns the counter value of the last notification packet
func (c *NotifyCounter) Load() uint16 {
	return uint16((*atomic.Uint32)(c).Load())
}

// NotifySession is the metadata for the session
//
// Peer must not be modified after creation. Counter and Errors must be accessed
// atomically. NotifySession contains unexported autodiscovered metadata.
//
// Use NewNotifySession to create a NotifySession with autodiscovery.
type NotifySession struct {
	Peer netip.AddrPort
	// Figuring out the local source IP and MAC address is suprisingly complex.
	// We store them as session data to avoid re-discovering at every packet.
	srcIP   netip.Addr
	srcMac  net.HardwareAddr
	Counter NotifyCounter
	Errors  atomic.Int32
}

// NewNotifySession creates a new NotifySession with autodiscovered metadata.
func NewNotifySession(peer netip.AddrPort) *NotifySession {
	srcIP := discoverSource(peer.Addr())
	srcMac := discoverMac(srcIP)
	return &NotifySession{
		Peer:   peer,
		srcIP:  srcIP,
		srcMac: srcMac,
	}
}

// Send sends a notification packet to the session.
//
// It retries up to 3 times and returns the error of the last try.
func (s *NotifySession) Send(res AWResponse) error {
	b := notifyPack(res.packResponse(), s, time.Now())
	var err error
	for i := 0; i < 3; i++ {
		err = sendTCP(b, s.Peer)
		if err == nil {
			s.Errors.Store(0)
			return nil
		}
		s.Errors.Add(1)
		// Sleeping just 100ms is the backoff strategy of Panasonic hardware.
		time.Sleep(100 * time.Millisecond)
	}
	return &NetworkError{err}
}

// NotifyList is a thread-safe locked list of NotifySessions
type NotifyList struct {
	lock sync.Mutex
	// It has been observed that Panasonic allows only one notification session
	// per IP address, stopping any previous ones when a new port is registered,
	// effectively keying the map on IP instead of IP:PORT. Uncertain if it is
	// intentional, we stick to IP:PORT mapping.
	list map[netip.AddrPort]*NotifySession
}

// SendAll sends a notification to all active sessions.
//
// Sessions which have observed 3 or more errors (equal to one dropped message)
// are automatically removed from the active sessions list.
func (l *NotifyList) SendAll(res AWResponse) {
	l.lock.Lock()
	defer l.lock.Unlock()
	for dst, s := range l.list {
		if s.Errors.Load() > 2 {
			delete(l.list, dst)
			continue
		}
		s.Send(res)
	}
}

// Add adds a new peer to the session list
func (l *NotifyList) Add(peer netip.AddrPort) {
	l.lock.Lock()
	defer l.lock.Unlock()
	if l.list == nil {
		l.list = make(map[netip.AddrPort]*NotifySession)
	}
	l.list[peer] = NewNotifySession(peer)
}

// Remove removes a peer from the session list
func (l *NotifyList) Remove(peer netip.AddrPort) {
	l.lock.Lock()
	defer l.lock.Unlock()
	delete(l.list, peer)
}

// Len returns the number of active sessions
func (l *NotifyList) Len() int {
	l.lock.Lock()
	defer l.lock.Unlock()
	for dst, s := range l.list {
		if s.Errors.Load() > 2 {
			delete(l.list, dst)
			continue
		}
	}
	return len(l.list)
}

type Handler interface {
	ServeAW(AWRequest) AWResponse
}

type HandlerFunc func(AWRequest) AWResponse

func (f HandlerFunc) ServeAW(r AWRequest) AWResponse {
	return f(r)
}

// HttpHandler is an http.Handler that implements an endpoint for AW protocol.
type HttpHandler struct {
	once      sync.Once
	mux       http.ServeMux
	AWHandler Handler
	Sessions  NotifyList
}

func (c *HttpHandler) setup() {
	c.mux = http.ServeMux{}
	c.mux.HandleFunc("/cgi-bin/aw_ptz", c.servePtz)
	c.mux.HandleFunc("/cgi-bin/aw_cam", c.serveCam)
	c.mux.HandleFunc("/cgi-bin/event", c.serveEvent)
	c.mux.HandleFunc("/cgi-bin/man_session", c.serveManSession)
}

func (c *HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (c *HttpHandler) servePtz(w http.ResponseWriter, r *http.Request) {
	c.wrapAW(awPtz, w, r)
}

func (c *HttpHandler) serveCam(w http.ResponseWriter, r *http.Request) {
	c.wrapAW(awCam, w, r)
}

func (c *HttpHandler) wrapAW(hint awHint, w http.ResponseWriter, r *http.Request) {
	// Generate a "Bad Request" for missing parameters
	qry := r.URL.Query()
	if qry.Get("res") != "1" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	strcmd := qry.Get("cmd")
	if strcmd == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	awcmd := newRequest(hint, strcmd)
	awres := c.AWHandler.ServeAW(awcmd)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(awres.packResponse()))
}

func (c *HttpHandler) serveEvent(w http.ResponseWriter, r *http.Request) {
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
	// There's also a uid parameter, which seem to be ignored by cameras.
	// Documentation states to set it to 0. AW-RP50 sets it to 50. We ignore it.

	ip := netip.MustParseAddrPort(r.RemoteAddr).Addr()
	client := netip.AddrPortFrom(ip, uint16(portNo))

	switch connect {
	case "start":
		c.Sessions.Add(client)
	case "stop":
		c.Sessions.Remove(client)
	default:
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *HttpHandler) serveManSession(w http.ResponseWriter, r *http.Request) {
	// The command=get parameter is required. We opt to provide a 400 instead of
	// the random mix of 204/403 codes observed in real devices.
	if r.URL.Query().Get("command") != "get" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	// Quick workaround to make AW-RP50 think it is connected to us.
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Event session:"))
	w.Write([]byte(strconv.Itoa(c.Sessions.Len())))
}
