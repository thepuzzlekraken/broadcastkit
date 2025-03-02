package yamaha

import "fmt"

// InfoMessage is a special action message within the Yamaha SCP protocol.
//
// Most common actions are: devinfo, devicename, devstatus, scpmode
type InfoMessage struct {
	Action  string
	Address AddressString
	Value   string
}

func (m *InfoMessage) _msg() {}

func parseInfo(line []byte) (Message, error) {
	l := trimSpace(line)
	action, l := cutSpace(l)
	if len(action) == 0 {
		return nil, fmt.Errorf("syntax error: %s, missing action", line)
	}
	if !startsSpace(l) {
		return nil, fmt.Errorf("syntax error: %s, missing separator", line)
	}
	address, l := cutWord(l)
	if len(address) == 0 {
		return nil, fmt.Errorf("syntax error: %s, missing address", line)
	}
	value, _ := cutWord(l)
	if len(value) == 0 {
		return nil, fmt.Errorf("syntax error: %s, missing value", line)
	}
	return &InfoMessage{
		Action:  string(action),
		Address: AddressString(address),
		Value:   string(value),
	}, nil
}
