package sony

import "fmt"

type endpoint string

const inquiryEndpoint endpoint = "inquiry"

// inquiryParameter is the parameter passed to the special inquery endpoint
// to download all parameters of a different endpoint. It is intentionally not
// registered as it should never be returned from the camera.
type inquiryParameter endpoint

func (_ inquiryParameter) parameterKey() string {
	return "inq"
}
func (ep inquiryParameter) parameterValue() string {
	return string(ep) // cut leading slash and trailing .cgi
}
func (_ inquiryParameter) parameterParse(s string) (Parameter, error) {
	return inquiryParameter(s), nil
}
func (_ inquiryParameter) Valid() bool {
	return true
}

type Parameter interface {
	Valid() bool
	parameterKey() string
	parameterValue() string
	parameterParse(string) (Parameter, error)
}

var parameterTable = make(map[endpoint]map[string]func() Parameter)

func registerParameter(ep endpoint, new func() Parameter) {
	key := new().parameterKey()
	_, ok := parameterTable[ep]
	if !ok {
		parameterTable[ep] = make(map[string]func() Parameter)
	}
	parameterTable[ep][key] = new
}

func createParameter(ep endpoint, key string, val string) (Parameter, error) {
	_, ok := parameterTable[ep]
	if !ok {
		// This should never happen, aid debugging
		panic("lookup of non-existent sony parameter table")
	}
	new, ok := parameterTable[ep][key]
	if !ok {
		return nil, fmt.Errorf("unknown parameter: %s.cgi?%s=...", ep, key)
	}
	p, err := new().parameterParse(val)
	if err != nil {
		return nil, err
	}
	return p, nil
}
