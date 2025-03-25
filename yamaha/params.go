package yamaha

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

const DbMin int = -32768 // -Inf
const DbMax int = 10000  // +10 dB
const DbZero int = 0     // 0 dB

// StringParam represents a string parameter
type StringParam struct {
	Set      bool
	Address  AddressString
	AddressX int
	AddressY int
	Value    string
}

func (p *StringParam) _msg() {}

// IntParam represents an integer parameter
//
// Values often represent 1/1000dB, from FaderMin to FaderMax. The exact meaning
// is determined by the parameter address.
type IntParam struct {
	Set      bool
	Address  AddressString
	AddressX int
	AddressY int
	Value    int
}

func (p *IntParam) _msg() {}

func parseParam(line []byte) (Message, error) {
	l := line
	action, l := cutSpace(l)
	set := bytes.Equal(action, []byte("set"))
	if !set && !bytes.Equal(action, []byte("get")) {
		return nil, fmt.Errorf("broadcastkit/yamaha: invalid action: %s", line)
	}

	address, l := cutWord(l)
	if len(address) == 0 {
		return nil, fmt.Errorf("broadcastkit/yamaha: syntax: missing address: %s", line)
	}

	if !startsSpace(l) {
		return nil, fmt.Errorf("broadcastkit/yamaha: syntax: missing address separator: %s", line)
	}
	bX, l := cutSpace(l)
	if len(bX) == 0 {
		return nil, fmt.Errorf("broadcastkit/yamaha: syntax: missing x parameter: %s", line)
	}
	pX, err := strconv.Atoi(string(bX))
	if err != nil {
		return nil, fmt.Errorf("broadcastkit/yamaha: syntax: x parameter not a number: %s", line)
	}

	if !startsSpace(l) {
		return nil, fmt.Errorf("broadcastkit/yamaha: syntax: missing x-y separator: %s", line)
	}
	bY, l := cutSpace(l)
	if len(bX) == 0 {
		return nil, fmt.Errorf("broadcastkit/yamaha: syntax: missing y parameter: %s", line)
	}
	pY, err := strconv.Atoi(string(bY))
	if err != nil {
		return nil, fmt.Errorf("broadcastkit/yamaha: syntax: y parameter not a number: %s", line)
	}

	// TODO(zsh): For Yamaha emulation, support required for missing values.
	if !startsSpace(l) {
		return nil, fmt.Errorf("broadcastkit/yamaha: syntax: missing parameter separator: %s", line)
	}
	bV, _ := cutWord(l)
	if len(bV) == 0 {
		return nil, fmt.Errorf("broadcastkit/yamaha: syntax: missing value: %s", line)
	}

	if l = trimSpace(l); l[0] != '"' {
		v, err := strconv.Atoi(string(bV))
		if err != nil {
			return nil, fmt.Errorf("broadcastkit/yamaha: syntax: value not a number: %s", line)
		}
		return &IntParam{
			Set:      set,
			Address:  AddressString(address),
			AddressX: pX,
			AddressY: pY,
			Value:    v,
		}, nil
	}

	return &StringParam{
		Set:      set,
		Address:  AddressString(address),
		AddressX: pX,
		AddressY: pY,
		Value:    string(bV),
	}, nil
}

// AddressString is a string which represent a Yamaha parameter address.
//
// This is a convenience wrapper for enum-like autocomple and type-check.
type AddressString string

func autoquote(s AddressString) string {
	if strings.ContainsAny(string(s), " \"") {
		return fmt.Sprintf("%q", s)
	}
	return string(s)
}

func (a AddressString) String() string {
	return autoquote(a)
}

const (
	MasterFaderAddr  AddressString = "MIXER:Current/St/Fader/Level"
	ChFaderAddr      AddressString = "MIXER:Current/InCh/Fader/Level"
	StChFaderAddr    AddressString = "MIXER:Current/StInCh/Fader/Level"
	ChToMixAddr      AddressString = "MIXER:Current/InCh/ToMix/Level"
	StChToMixAddr    AddressString = "MIXER:Current/StInCh/ToMix/Level"
	ChToMatrixAddr   AddressString = "MIXER:Current/InCh/ToMtrx/Level"
	StChToMatrixAddr AddressString = "MIXER:Current/StInCh/ToMtrx/Level"
	MixFaderAddr     AddressString = "MIXER:Current/Mix/Fader/Level"
	MatrixFaderAddr  AddressString = "MIXER:Current/Mtrx/Fader/Level"
	MixToMatrixAddr  AddressString = "MIXER:Current/Mix/ToMtrx/Level"
	ChNameAddr       AddressString = "MIXER:Current/InCh/Label/Name"
	StChNameAddr     AddressString = "MIXER:Current/StInCh/Label/Name"
	MixNameAddr      AddressString = "MIXER:Current/Mix/Label/Name"
	MatrixNameAddr   AddressString = "MIXER:Current/Mtrx/Label/Name"
	ChColorAddr      AddressString = "MIXER:Current/InCh/Label/Color"
	StChColorAddr    AddressString = "MIXER:Current/StInCh/Label/Color"
	MixColorAddr     AddressString = "MIXER:Current/Mix/Label/Color"
	MatrixColorAddr  AddressString = "MIXER:Current/Mtrx/Label/Color"
)

const (
	Blue    string = "Blue"
	Orange  string = "Orange"
	Yellow  string = "Yellow"
	Purple  string = "Purple"
	Cyan    string = "Cyan"
	Magenta string = "Magenta"
	Red     string = "Red"
	Green   string = "Green"
	Off     string = "Off"
)

// Yamaha CL/QL mixers use both /StIn/ and /StInCh/ for many parameters. They
// accept either in set/get and send both upon NOTIFY set, presumably for
// backwards compatibility. Some parameters available via /StInCh/ only.
// You may wish to drop all /StIn/ notifications to processing overhead.
//
//	socket := yamaha.DialSCP("203.0.113.123")
//	for {
//		_, msg, _ := socket.Read()
//		if strings.HasPrefix(msg.Address, LegacyStPrefix) {
//			continue
//		}
//		// process msg normally
//	}
const LegacyStPrefix AddressString = "MIXER:Current/StIn/"
