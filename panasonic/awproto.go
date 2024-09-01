package panasonic

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"sync"
	"time"
)

// awHint is a bitmask marking where this command is expected to appear.
//
// Because the Panasonic commands textual representation is not unique, this
// is used to determine look-up tables and endpoints.
type awHint int

const (
	awPtz awHint = 1 << iota // expected over aw_ptz interface
	awCam                    // expected over aw_cam interface
	awNty                    // expected over notifications interface
	awHintMax
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
var awRequestTable = [awHintMax][]awRequestFactory{}

// awResponseTable is the factory lookup table for AWResponses
var awResponseTable = [awHintMax][]awResponseFactory{}

// registerRequest registers a new request type with the factory table
func registerRequest(new func() AWRequest) {
	// TODO(zsh): These functions may be optimized away by code-generation instead.
	n := new()
	f, p := n.requestSignature()
	for i := range awHintMax {
		m := 1 << i
		if int(f)&m != 0 {
			awRequestTable[i] = append(awRequestTable[i], awRequestFactory{p, new})
			return
		}
	}
}

// registerResponse registers a new response type with the factory table
func registerResponse(new func() AWResponse) {
	// TODO(zsh): These functions may be optimized away by code-generation instead.
	n := new()
	f, p := n.responseSignature()
	for i := range awHintMax {
		m := 1 << i
		if int(f)&m != 0 {
			awResponseTable[i] = append(awResponseTable[i], awResponseFactory{p, new})
			return
		}
	}
}

// newRequest creates a new request via the factory
func newRequest(hint awHint, cmd string) AWRequest {
	// This function is within a latency-critical path of incoming requests.
	// Following is a tight-loop with everything inlined, but this may need
	// optimization if latency becomes an issue.
	for i := range awHintMax {
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
	for i := range awHintMax {
		m := 1 << i
		if int(hint)&m == 0 {
			continue
		}
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

// tcpUnwrap peels out the string command from a tcp connection buffer
func tcpUnwrap(b []byte) (string, error) {
	if len(b) < 24 {
		return "", fmt.Errorf("panasonic.tcpUnwrap: data too short %d (expected >=24)", len(b))
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

// awListener is the go routine loop that keeps accepting incoming notifications
// listener is the net.Listener to be used
// camaddr is the source address of notifications
// ch is the channel to send AWResponses to
func awListener(listener net.Listener, camaddr netip.Addr, ch chan<- AWResponse) {
	defer listener.Close()
	defer close(ch)
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}

		addr, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
		remaddr, _ := netip.ParseAddr(addr)
		if remaddr != camaddr {
			conn.Close()
			continue
		}

		conn.SetDeadline(time.Now().Add(1 * time.Second))
		b, err := io.ReadAll(conn)
		conn.Close()
		if err != nil {
			continue
		}

		cmd, err := tcpUnwrap(b)
		if err != nil {
			continue
		}

		ch <- newResponse(awNty, cmd)
	}
}

// AWErrNo is the error number used by the Panasonic AW protocol
type AWErrNo int

const (
	AWErrUnsupported  AWErrNo = 1 // The command is not understood by the device
	AWErrBusy         AWErrNo = 2 // The device is not ready for the command
	AWErrUnacceptable AWErrNo = 3 // The command values are not acceptable
	// Numbers higher than 3 are unused, we parse them for future-proofing only
)

// AWError is a response indicating error reported by the Panasonic device
type AWError struct {
	cap  bool
	No   AWErrNo // The error number reported
	Flag string  // The textual flag reported (usually the begining of command)
}

func (e *AWError) Ok() bool {
	return false
}
func (e *AWError) responseSignature() (awHint, string) {
	sig := "\x03R\x01:\x00\x00\x00"
	return awPtz | awCam, sig[0:min(4+len(e.Flag), 7)]
}
func (e *AWError) unpackResponse(s string) {
	e.cap = s[0] == 'E'
	e.No = AWErrNo(dec2int(s[2:3]))
	e.Flag = s[4:]
}
func (e *AWError) packResponse() string {
	if e.cap {
		return "ER" + int2dec(int(e.No), 1) + ":" + e.Flag
	}
	return "eR" + int2dec(int(e.No), 1) + ":" + e.Flag
}
func init() {
	// All commands are fixed length, except for errors which take up to 3 chars
	// we just register it 4 times so we don't need to teach the match enginge
	// about variable-length responses
	registerResponse(func() AWResponse { return &AWError{Flag: ""} })
	registerResponse(func() AWResponse { return &AWError{Flag: " "} })
	registerResponse(func() AWResponse { return &AWError{Flag: "  "} })
	registerResponse(func() AWResponse { return &AWError{Flag: "   "} })
}

// Error implements the error interface
// This is so that Panasonic protocol errors can be returned both via AWResponse
// and via the go standard error interface
func (e *AWError) Error() string {
	switch e.No {
	case AWErrUnsupported:
		return fmt.Sprintf("unsupported AW command (%s)", e.Flag)
	case AWErrBusy:
		return fmt.Sprintf("busy status for AW command (%s)", e.Flag)
	case AWErrUnacceptable:
		return fmt.Sprintf("unacceptable value for AW command (%s)", e.Flag)
	default:
		return fmt.Sprintf("unknown error (%d) for AW command (%s)", e.No, e.Flag)
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

// httpInit sets good defaults of the Http client for the AW protocol
//
// Called via httpOnce.Do(..) before the first request.
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

func (c *Camera) strCommand(hint awHint, cmd string) (string, error) {
	var path string

	if hint&awPtz != 0 {
		path = "/cgi-bin/aw_ptz"
	} else {
		path = "/cgi-bin/aw_cam"
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

func (c *Camera) Notify(ch chan<- AWResponse) (stop func(), err error) {
	c.httpOnce.Do(c.httpInit)

	if err != nil {
		close(ch)
		return nil, &NetworkError{err}
	}

	listener, err := net.Listen("tcp4", "0.0.0.0:0")
	if err != nil {
		close(ch)
		return nil, &NetworkError{err}
	}
	stop = func() {
		listener.Close()
	}

	go awListener(listener, c.Addr, ch)

	_, port, _ := net.SplitHostPort(listener.Addr().String())
	res, err := c.httpGet("/cgi-bin/event", "connect=start&my_port="+port+"&uid=0")

	if err != nil {
		stop()
		return nil, &NetworkError{err}
	}
	if res.StatusCode != http.StatusNoContent {
		stop()
		err := fmt.Errorf("http status code: %d (expected %d)", res.StatusCode, http.StatusNoContent)
		return nil, &NetworkError{err}
	}

	return stop, nil
}
