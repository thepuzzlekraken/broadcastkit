package yamaha

import "fmt"

type InfoMessage struct {
	Action  string
	Address string
	Value   string
}

func (m *InfoMessage) _recv() {}
func (m *InfoMessage) _send() {}

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
		Address: string(address),
		Value:   string(value),
	}, nil
}
