package sony

import "fmt"

type Button string

const (
	ButtonPress   Button = "press"
	ButtonRelease Button = "release"
)

func (b Button) Valid() bool {
	switch b {
	case ButtonPress, ButtonRelease:
		return true
	default:
		return false
	}
}

type Direction string

const (
	UpDirection        Direction = "up"
	DownDirection      Direction = "down"
	LeftDirection      Direction = "left"
	RightDirection     Direction = "right"
	UpLeftDirection    Direction = "up-left"
	UpRightDirection   Direction = "up-right"
	DownLeftDirection  Direction = "down-left"
	DownRightDirection Direction = "down-right"
	StopDirection      Direction = "stop"
)

func (d Direction) Valid() bool {
	switch d {
	case UpDirection,
		DownDirection,
		LeftDirection,
		RightDirection,
		UpLeftDirection,
		UpRightDirection,
		DownLeftDirection,
		DownRightDirection,
		StopDirection:
		return true
	default:
		return false
	}
}

// SpeedStep is a speed represented as 0-50 or 0-24 depending on the value of
// PanTiltSpeedStep setting
type SpeedStep int

func (s SpeedStep) Valid() bool {
	return (s >= 0) && (s <= 50)
}

type PanTiltMoveParameter struct {
	Direction       Direction
	HorizontalSpeed SpeedStep
	VerticalSpeed   SpeedStep
}

func (p PanTiltMoveParameter) parameterKey() string {
	return "PanTiltMove"
}
func (p PanTiltMoveParameter) parameterValue() string {
	return commaJoin(string(p.Direction), itoa(int(p.HorizontalSpeed)), itoa(int(p.VerticalSpeed)))
}
func (p PanTiltMoveParameter) parameterParse(s string) (Parameter, error) {
	sp := commaSplit(s)
	if len(sp) != 3 {
		return nil, fmt.Errorf("invalid comma-joined-list length: %d, expects 3", len(sp))
	}
	dir := sp[0]
	hS, err := atoi(sp[1])
	if err != nil {
		return nil, err
	}
	vS, err := atoi(sp[2])
	if err != nil {
		return nil, err
	}
	return PanTiltMoveParameter{
		Direction:       Direction(dir),
		HorizontalSpeed: SpeedStep(hS),
		VerticalSpeed:   SpeedStep(vS),
	}, nil
}
func (p PanTiltMoveParameter) Valid() bool {
	return p.Direction.Valid() && p.HorizontalSpeed.Valid() && p.VerticalSpeed.Valid()
}
func (p PanTiltMoveParameter) _ptzfParameter() {}

func init() { registerParameter(ptzfEndpoint, func() Parameter { return PanTiltMoveParameter{} }) }
