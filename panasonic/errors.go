package panasonic

import "fmt"

// AWErrNo is the error number within the Panasonic AW protocol
type AWErrNo int

const (
	AWErrUnsupported  AWErrNo = 1 // The command is not understood by the device
	AWErrBusy         AWErrNo = 2 // The device is not ready for the command
	AWErrUnacceptable AWErrNo = 3 // The command values are not acceptable
	// Numbers higher than 3 are unused, we parse them for future-proofing only
)

// AWError is an error response over the Panasonic AW protocol
type AWError struct {
	cap  bool
	No   AWErrNo // The error number reported
	Flag string  // The textual flag reported (usually the begining of command)
}

// Error implements the error interface
// This is intended for the ability to return Panasonic errors as go error
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

func (e *AWError) Ok() bool {
	return false
}
func (e *AWError) responseSignature() (awHint, string) {
	sig := "\x03R\x01:\x00\x00\x00"
	// return the correct-length pattern depending on flag length
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
	// AW commands are fixed length, except for errors.	Pattern matching is
	// fixed-length, so we register a separate error for each possible length.
	registerResponse(func() AWResponse { return &AWError{Flag: ""} })
	registerResponse(func() AWResponse { return &AWError{Flag: " "} })
	registerResponse(func() AWResponse { return &AWError{Flag: "  "} })
	registerResponse(func() AWResponse { return &AWError{Flag: "   "} })
}
