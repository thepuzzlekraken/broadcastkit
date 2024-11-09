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
func (e AWError) Error() string {
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

func (e AWError) responseSignature() string {
	sig := "\x03R\x02:\x7F"
	// return the correct-length pattern depending on flag length
	return sig[0:min(4+len(e.Flag), 5)]
}
func (e AWError) unpackResponse(s string) AWResponse {
	e.cap = s[0] == 'E'
	e.No = AWErrNo(dec2int(s[2:3]))
	e.Flag = s[4:]
	return e
}
func (e AWError) packResponse() string {
	if e.cap {
		return "ER" + int2dec(int(e.No), 1) + ":" + e.Flag
	}
	return "eR" + int2dec(int(e.No), 1) + ":" + e.Flag
}
func init() {
	// The \xF7 matches any character of 1+ length, but error may have 0
	registerResponse(func() AWResponse { return AWError{Flag: ""} })
	registerResponse(func() AWResponse { return AWError{Flag: " "} })
}

// NewAWError creates an AWError as a Panasonic device would.
// This is intended for simulating errors as a proxy or virtual device.
func NewAWError(n AWErrNo, c AWRequest) AWError {
	t := c.packRequest()
	return AWError{
		cap:  len(t) > 0 && t[0] != '#',
		No:   n,
		Flag: t[:min(len(t), 3)],
	}
}

// SystemError is an error condition outside of the Panasonic protocol
type SystemError struct {
	parent error
}

func (e *SystemError) Error() string {
	return fmt.Sprintf("panasonic system failure: %s", e.parent.Error())
}
func (e *SystemError) Unwrap() error {
	return e.parent
}
