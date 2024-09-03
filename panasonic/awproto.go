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

// Camera represent a remote camera to be controlled via the AW protocol
type Camera struct {
	// Addr is the IP address of the camera
	Addr netip.Addr
	// Http is the HTTP client to use for requests
	//
	// Some variables are adjusted for Panasonic protocol unless explicitly set
	// before the first request is sent.
	// CheckRedirects is set func(...) error { return http.ErrUseLastResponse }
	// Timeout is set 1 * time.Second
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
		c.Http.Timeout = 1 * time.Second
	}
}

// httpGet does an http.Get to the camera with the quirks of the AW protocol
func (c *Camera) httpGet(path string, query string) (*http.Response, error) {
	c.httpOnce.Do(c.httpInit)
	return c.Http.Do(&http.Request{
		Method: "GET",
		URL: &url.URL{
			Scheme: "http",
			Host:   c.Addr.String(),
			Path:   path,
			// Note that query is NOT urlencoded and will contain non-compliant
			// characters like '#'. This matches the behavior of remote panels.
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
func (c *Camera) strCommand(hint awHint, cmd string) (string, error) {
	var path string

	if hint&awPtz != 0 {
		path = "/cgi-bin/aw_ptz"
	} else if hint&awCam != 0 {
		path = "/cgi-bin/aw_cam"
	} else {
		return "", fmt.Errorf("this command %v")
	}

	res, err := c.httpGet(path, "cmd="+cmd+"&res=1")
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

func (c *Camera) Command(req AWRequest) (AWResponse, error) {
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

func (c *Camera) UpdateListener() (*CameraUpdateListener, error) {
	listener, err := net.ListenTCP("tcp4", &net.TCPAddr{})
	if err != nil {
		return nil, &NetworkError{err}
	}
	return &CameraUpdateListener{
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
// -  l bytes string data (surrounded by CRLF)
// -  4 bytes constant, hex 00 02 00 18
// -  6 bytes source MAC
// -  2 bytes constant, hex 00 01
// -  6 bytes date, same as above
// -  6 bytes constant 00 00 00 00 00 00
// Date is in local time of the camera, not UTC.
func notifyPack(response string, counter uint16, camIP netip.Addr, camMAC net.HardwareAddr, date time.Time) []byte {
	// Surround the response with CRLF fluff
	// TODO(zsh): Should we mimick the 0-padding quirks of the camera?
	data := make([]byte, len(response)+4)
	data[0] = '\r'
	data[1] = '\n'
	copy(data[2:], response)
	data[len(data)-2] = '\r'
	data[len(data)-1] = '\n'

	// Create the data packet as described above
	f := 4 + 2 + 6 + 10 + 2 + 4 + len(data) + 4 + 6 + 2 + 6 + 6
	b := make([]byte, f)
	_ = b[28]
	copy(b, camIP.AsSlice())
	binary.BigEndian.PutUint16(b[4:], counter)
	b[6] = uint8(date.Year() % 100)
	b[7] = uint8(date.Month())
	b[8] = uint8(date.Day())
	b[9] = uint8(date.Hour())
	b[10] = uint8(date.Minute())
	b[11] = uint8(date.Second())
	copy(b[12:], []byte{0x00, 0x01, 0x00, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01})
	binary.BigEndian.PutUint16(b[22:], uint16(len(data)+8))
	copy(b[24:], []byte{0x01, 0x00, 0x00, 0x00})
	copy(b[28:], data)
	copy(b[f-25:f-21], []byte{0x00, 0x02, 0x00, 0x18})
	copy(b[f-21:f-15], camMAC)
	copy(b[f-15:f-13], []byte{0x00, 0x01})
	copy(b[f-13:f-7], b[6:12])
	copy(b[f-7:f-1], []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

	return b
}

type CameraUpdateListener struct {
	once sync.Once
	lis  *net.TCPListener
	cam  *Camera
}

func (u *CameraUpdateListener) Start() error {
	port := netip.MustParseAddrPort(u.lis.Addr().String()).Port()
	res, err := u.cam.httpGet("/cgi-bin/event", "connect=start&my_port="+strconv.Itoa(int(port))+"&uid=0")
	if err != nil {
		return &NetworkError{err}
	}
	if res.StatusCode != http.StatusNoContent {
		return &NetworkError{fmt.Errorf("http status code: %d (expected %d)", res.StatusCode, http.StatusNoContent)}
	}
	return nil
}

func (u *CameraUpdateListener) Stop() error {
	port := netip.MustParseAddrPort(u.lis.Addr().String()).Port()
	res, err := u.cam.httpGet("/cgi-bin/event", "connect=start&my_port="+strconv.Itoa(int(port))+"&uid=0")
	if err != nil {
		return &NetworkError{err}
	}
	if res.StatusCode != http.StatusNoContent {
		return &NetworkError{fmt.Errorf("http status code: %d (expected %d)", res.StatusCode, http.StatusNoContent)}
	}
	return nil
}

func (u *CameraUpdateListener) Addr() netip.AddrPort {
	return netip.MustParseAddrPort(u.lis.Addr().String())
}

func (u *CameraUpdateListener) acceptTCP() (*net.TCPConn, error) {
	camaddr := u.cam.Addr
	for {
		conn, err := u.lis.AcceptTCP()
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

func (u *CameraUpdateListener) Accept() (AWResponse, error) {
	u.once.Do(func() { u.Start() })

	conn, err := u.acceptTCP()
	if err != nil {
		return nil, &NetworkError{err}
	}

	conn.SetDeadline(time.Now().Add(1 * time.Second))
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

func (u *CameraUpdateListener) Close() error {
	_ = u.Stop()
	return u.lis.Close()
}
