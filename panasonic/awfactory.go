package panasonic

// AWResponse is the interface implemented by all responses sent from a camera.
//
// For processing information, the application has to type-assert the response
// to the specific implementation. Note that Responses returned may be AWError
// anytime.
type AWResponse interface {
	// responseSignature returns the pattern of Panasonic string literals.
	//
	// Pattern is in a custom format of this package. The match() function
	// should be used to test the pattern on any string literals before passing
	// them to unpackResponse().
	responseSignature() string
	// unpackResponse parses response values from Panasonic string literal.
	//
	// It is the responsibility of the caller to ensure pattern match before
	// a call. Behavior is undefined if the passed string does not match the
	// signature possibly resulting in panic or invalid command data!
	unpackResponse(string) AWResponse
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
	// requestSignature returns the pattern of Panasonic string.
	//
	// Limitations are the same as AWResponse.responseSignature()
	requestSignature() string
	// unpackRequest parses request values from Panasonic string.
	//
	// The passed string must match the requestSignature() pattern. Limitations
	// are the same as AWResponse.unpackResponse().
	unpackRequest(string) AWRequest
	// packRequest returns the Panasonic string representation of the request.
	//
	// The returned string is guaranteed to match the resquestSignature()
	// pattern, but it may still be invalid semantically.
	packRequest() string
}

// awQuirkedPack is implemented by an AWResponse when it needs to be aware of
// it's context to produce a valid packing or unpacking.
type awQuirkedPacking interface {
	packingQuirk(mode quirkMode) AWResponse
}

type quirkMode int

const (
	quirkBatch  quirkMode = iota // set when in a batch reply
	quirkNotify                  // set when in a notification
	quirkCamera                  // set when in an aw_cam endpoint
	quirkPtz                     // set when in an aw_ptz endpoint
)

// AWUknownResponse is a placeholder implementation for AWResponse.
//
// Used when a non-error response is not recognized by this library. It is not
// possible to understand the meaning of such replies. They are intended for
// proxying only.
type AWUnknownResponse struct {
	text string
}

func (a AWUnknownResponse) responseSignature() string {
	return a.text
}
func (a AWUnknownResponse) unpackResponse(_ string) AWResponse {
	return a
}
func (a AWUnknownResponse) packResponse() string {
	return a.text
}

// AWUknownRequest is a placeholder implementation for AWRequest.
//
// Used when a request is not recognized by this library. It is not possible to
// understand the meaning of such requests. Their intended use is proxying or to
// be replied with AWErrUnsupported.
type AWUnknownRequest struct {
	text string
}

func (a AWUnknownRequest) Acceptable() bool {
	return true
}
func (a AWUnknownRequest) Response() AWResponse {
	// Implementation note: AWUnknownRequest and AWUnknownResponse are separate
	// struct to avoid applications unknowingly casting them to the other type.
	return AWUnknownResponse{}
}
func (a AWUnknownRequest) requestSignature() string {
	return a.text
}
func (a AWUnknownRequest) unpackRequest(_ string) AWRequest {
	return a
}
func (a AWUnknownRequest) packRequest() string {
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
var awRequestTable = []awRequestFactory{}

// awResponseTable is the factory lookup table for AWResponses
var awResponseTable = []awResponseFactory{}

// registerRequest registers a new request type with the factory table
func registerRequest(new func() AWRequest) {
	// TODO(zsh): These functions may be optimized away by code-generation instead.
	n := new()
	p := n.requestSignature()
	awRequestTable = append(awRequestTable, awRequestFactory{p, new})
}

// registerResponse registers a new response type with the factory table
func registerResponse(new func() AWResponse) {
	// TODO(zsh): These functions may be optimized away by code-generation instead.
	n := new()
	p := n.responseSignature()
	awResponseTable = append(awResponseTable, awResponseFactory{p, new})
}

// newRequest creates a new request via the factory
func newRequest(cmd string) AWRequest {
	// This function is within a latency-critical path of incoming requests.
	// Following is a tight-loop with everything inlined, but this may need
	// optimization if latency becomes an issue.
	for _, e := range awRequestTable {
		if match(e.sig, cmd) {
			req := e.new()
			req.unpackRequest(cmd)
			return req
		}
	}
	return AWUnknownRequest{
		text: cmd,
	}
}

// newResponse creates a new response via the factory
func newResponse(cmd string) AWResponse {
	// This function is less critical than newRequest(), because the object
	// returned by AWRequest.Response() is used in happy-path response creation.
	for _, e := range awResponseTable {
		if match(e.sig, cmd) {
			res := e.new()
			res.unpackResponse(cmd)
			return res
		}
	}
	return AWUnknownResponse{
		text: cmd,
	}
}
