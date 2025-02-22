package yamaha

import (
	"bytes"
	"fmt"
	"strconv"
)

const FaderMin int = -32768 // -Inf
const FaderMax int = 10000  // +10 dB
const FaderBase int = 0     // 0 dB

type StringParam struct {
	Set      bool
	Address  string
	AddressX int
	AddressY int
	Value    string
}

func (p *StringParam) _recv() {}
func (p *StringParam) _send() {}

type IntParam struct {
	Set      bool
	Address  string
	AddressX int
	AddressY int
	Value    int
}

func (p *IntParam) _recv() {}
func (p *IntParam) _send() {}

func parseParam(line []byte) (Message, error) {
	l := line
	action, l := cutSpace(l)
	set := bytes.Equal(action, []byte("set"))
	if !set && !bytes.Equal(action, []byte("get")) {
		return nil, fmt.Errorf("syntax error: %s, invalid action", line)
	}

	address, l := cutWord(l)
	if len(address) == 0 {
		return nil, fmt.Errorf("syntax error: %s, missing address", line)
	}

	if !startsSpace(l) {
		return nil, fmt.Errorf("syntax error: %s, missing separator", line)
	}
	bX, l := cutSpace(l)
	if len(bX) == 0 {
		return nil, fmt.Errorf("syntax error: %s, missing x parameter", line)
	}
	pX, err := strconv.Atoi(string(bX))
	if err != nil {
		return nil, fmt.Errorf("syntax error: %s, x parameter not a number", line)
	}

	if !startsSpace(l) {
		return nil, fmt.Errorf("syntax error: %s, missing separator", line)
	}
	bY, l := cutSpace(l)
	if len(bX) == 0 {
		return nil, fmt.Errorf("syntax error: %s, missing y parameter", line)
	}
	pY, err := strconv.Atoi(string(bY))
	if err != nil {
		return nil, fmt.Errorf("syntax error: %s, y parameter not a number", line)
	}

	// TODO(zsh): For Yamaha emulation, support required for missing values.
	if !startsSpace(l) {
		return nil, fmt.Errorf("syntax error: %s, missing separator", line)
	}
	bV, _ := cutWord(l)
	if len(bV) == 0 {
		return nil, fmt.Errorf("syntax error: %s, missing value", line)
	}

	if l[0] != '"' {
		v, err := strconv.Atoi(string(bV))
		if err != nil {
			return nil, fmt.Errorf("syntax error: %s, value not a number", line)
		}
		return &IntParam{
			Set:      set,
			Address:  string(address),
			AddressX: pX,
			AddressY: pY,
			Value:    v,
		}, nil
	}

	return &StringParam{
		Set:      set,
		Address:  string(address),
		AddressX: pX,
		AddressY: pY,
		Value:    string(bV),
	}, nil
}
