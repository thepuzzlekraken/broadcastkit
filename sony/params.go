package sony

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

//
// Common parameter units used
//

// Button represents a state of a physical button.
//
// A press should be followed by a release. Long press is generally understood
// when there's >1s delay between the press and the release commands.
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

// SpeedStep is a speed represented as 0-50.
//
// For some parameters, representation may be 0-24 depending on the value of
// PanTiltSpeedStep setting for compatibility with older Sony devices.
type SpeedStep int

func (s SpeedStep) Valid() bool {
	return (s >= 0) && (s <= 50)
}

// SteppedPosition is an absolute distance measurement with 235.9 steps/degree
type SteppedPosition int

const SteppedPositionByDegree = 235.9

func (s SteppedPosition) Valid() bool {
	return (s >= -0x7FFFF) && (s <= 0x7FFFF)
}

func (s SteppedPosition) String() string {
	return hex20Encoder(int(s))
}

// SteppedRange is an absolute range measurement within a scale
type SteppedRange int

const SteppedRangeMax SteppedRange = 0xFFFF
const SteppedRangeFocusMax SteppedRange = 0xFFFF
const SteppedRangeZoomMax SteppedRange = 0x4000

func (s SteppedRange) Valid() bool {
	return (s >= 0) && (s <= 0xFFFF)
}

func (s SteppedRange) String() string {
	return hex16Encoder(int(s))
}

// Direction represents a phisical direction relative to the current image
// orientation, in the perspective of the camera operator (remove viewer).
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

// ZoomDirection represents a depth direction within the image
type ZoomDirection string

const (
	TeleDirection     ZoomDirection = "tele"
	WideDirection     ZoomDirection = "wide"
	StopZoomDirection ZoomDirection = "stop"
)

func (d ZoomDirection) Valid() bool {
	switch d {
	case TeleDirection,
		WideDirection,
		StopZoomDirection:
		return true
	default:
		return false
	}
}

// Preset is the number of a preset slot, valid between 1 and 100.
type Preset int

func (p Preset) Valid() bool {
	return (p >= 1) && (p <= 100)
}

// PresetName is the name of a preset slot, up to 32 bytes without commas.
type PresetName string

func (p PresetName) Valid() bool {
	return len(p) <= 32 && !strings.ContainsAny(string(p), ",")
}

// Switch is a two-state boolean value
type Switch string

const (
	SwitchOn  Switch = "on"
	SwitchOff Switch = "off"
)

func (s Switch) Valid() bool {
	switch s {
	case SwitchOn, SwitchOff:
		return true
	default:
		return false
	}
}

//
// Parameters for assignableEndpoint
//

type AssignableButton1Param struct {
	state Button
}

func (AssignableButton1Param) parameterKey() string {
	return "AssignableButton1"
}
func (a AssignableButton1Param) parameterValue() string {
	return string(a.state)
}
func (AssignableButton1Param) parameterParse(s string) (Parameter, error) {
	return AssignableButton1Param{Button(s)}, nil
}
func (a AssignableButton1Param) Valid() bool {
	return a.state.Valid()
}
func (AssignableButton1Param) _assignableParameter() {}

func init() {
	registerParameter(func() Parameter { return AssignableButton1Param{} })
}

type AssignableButton2Param struct {
	state Button
}

func (AssignableButton2Param) parameterKey() string {
	return "AssignableButton2"
}
func (a AssignableButton2Param) parameterValue() string {
	return string(a.state)
}
func (AssignableButton2Param) parameterParse(s string) (Parameter, error) {
	return AssignableButton2Param{Button(s)}, nil
}
func (a AssignableButton2Param) Valid() bool {
	return a.state.Valid()
}
func (AssignableButton2Param) _assignableParameter() {}

func init() {
	registerParameter(func() Parameter { return AssignableButton2Param{} })
}

type AssignableButton3Param struct {
	state Button
}

func (AssignableButton3Param) parameterKey() string {
	return "AssignableButton3"
}
func (a AssignableButton3Param) parameterValue() string {
	return string(a.state)
}
func (AssignableButton3Param) parameterParse(s string) (Parameter, error) {
	return AssignableButton3Param{Button(s)}, nil
}
func (a AssignableButton3Param) Valid() bool {
	return a.state.Valid()
}
func (AssignableButton3Param) _assignableParameter() {}

func init() {
	registerParameter(func() Parameter { return AssignableButton3Param{} })
}

type AssignableButton4Param struct {
	state Button
}

func (AssignableButton4Param) parameterKey() string {
	return "AssignableButton4"
}
func (a AssignableButton4Param) parameterValue() string {
	return string(a.state)
}
func (AssignableButton4Param) parameterParse(s string) (Parameter, error) {
	return AssignableButton4Param{Button(s)}, nil
}
func (a AssignableButton4Param) Valid() bool {
	return a.state.Valid()
}
func (AssignableButton4Param) _assignableParameter() {}

func init() {
	registerParameter(func() Parameter { return AssignableButton4Param{} })
}

type AssignableButton5Param struct {
	state Button
}

func (AssignableButton5Param) parameterKey() string {
	return "AssignableButton5"
}
func (a AssignableButton5Param) parameterValue() string {
	return string(a.state)
}
func (AssignableButton5Param) parameterParse(s string) (Parameter, error) {
	return AssignableButton5Param{Button(s)}, nil
}
func (a AssignableButton5Param) Valid() bool {
	return a.state.Valid()
}
func (AssignableButton5Param) _assignableParameter() {}

func init() {
	registerParameter(func() Parameter { return AssignableButton5Param{} })
}

type AssignableButton6Param struct {
	state Button
}

func (AssignableButton6Param) parameterKey() string {
	return "AssignableButton6"
}
func (a AssignableButton6Param) parameterValue() string {
	return string(a.state)
}
func (AssignableButton6Param) parameterParse(s string) (Parameter, error) {
	return AssignableButton6Param{Button(s)}, nil
}
func (a AssignableButton6Param) Valid() bool {
	return a.state.Valid()
}
func (AssignableButton6Param) _assignableParameter() {}

func init() {
	registerParameter(func() Parameter { return AssignableButton6Param{} })
}

type AssignableButton7Param struct {
	state Button
}

func (AssignableButton7Param) parameterKey() string {
	return "AssignableButton7"
}
func (a AssignableButton7Param) parameterValue() string {
	return string(a.state)
}
func (AssignableButton7Param) parameterParse(s string) (Parameter, error) {
	return AssignableButton7Param{Button(s)}, nil
}
func (a AssignableButton7Param) Valid() bool {
	return a.state.Valid()
}
func (AssignableButton7Param) _assignableParameter() {}

func init() {
	registerParameter(func() Parameter { return AssignableButton7Param{} })
}

type AssignableButton8Param struct {
	state Button
}

func (AssignableButton8Param) parameterKey() string {
	return "AssignableButton8"
}
func (a AssignableButton8Param) parameterValue() string {
	return string(a.state)
}
func (AssignableButton8Param) parameterParse(s string) (Parameter, error) {
	return AssignableButton8Param{Button(s)}, nil
}
func (a AssignableButton8Param) Valid() bool {
	return a.state.Valid()
}
func (AssignableButton8Param) _assignableParameter() {}

func init() {
	registerParameter(func() Parameter { return AssignableButton8Param{} })
}

type AssignableButton9Param struct {
	state Button
}

func (AssignableButton9Param) parameterKey() string {
	return "AssignableButton9"
}
func (a AssignableButton9Param) parameterValue() string {
	return string(a.state)
}
func (AssignableButton9Param) parameterParse(s string) (Parameter, error) {
	return AssignableButton9Param{Button(s)}, nil
}
func (a AssignableButton9Param) Valid() bool {
	return a.state.Valid()
}
func (AssignableButton9Param) _assignableParameter() {}

func init() {
	registerParameter(func() Parameter { return AssignableButton9Param{} })
}

type AssignableButtonFocusHold struct {
	state Button
}

func (AssignableButtonFocusHold) parameterKey() string {
	return "AssignableButtonFocusHold"
}
func (a AssignableButtonFocusHold) parameterValue() string {
	return string(a.state)
}
func (AssignableButtonFocusHold) parameterParse(s string) (Parameter, error) {
	return AssignableButtonFocusHold{Button(s)}, nil
}
func (a AssignableButtonFocusHold) Valid() bool {
	return a.state.Valid()
}
func (AssignableButtonFocusHold) _assignableParameter() {}

//
// Parameters for ptzfEndpoint
//

// PanTiltMoveParam performs a continous movement of the camera
type PanTiltMoveParam struct {
	Direction       Direction
	HorizontalSpeed SpeedStep
	VerticalSpeed   SpeedStep
}

func (PanTiltMoveParam) parameterKey() string {
	return "PanTiltMove"
}
func (p PanTiltMoveParam) parameterValue() string {
	return commaJoin(string(p.Direction), itoa(int(p.HorizontalSpeed)), itoa(int(p.VerticalSpeed)))
}
func (PanTiltMoveParam) parameterParse(s string) (Parameter, error) {
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
	return PanTiltMoveParam{
		Direction:       Direction(dir),
		HorizontalSpeed: SpeedStep(hS),
		VerticalSpeed:   SpeedStep(vS),
	}, nil
}
func (p PanTiltMoveParam) Valid() bool {
	return p.Direction.Valid() && p.HorizontalSpeed.Valid() && p.VerticalSpeed.Valid()
}
func (PanTiltMoveParam) _ptzfParameter() {}
func init() {

	registerParameter(func() Parameter { return PanTiltMoveParam{} })
}

type AbsolutePanTiltParam struct {
	Pan   SteppedPosition
	Tilt  SteppedPosition
	Speed SpeedStep
}

func (AbsolutePanTiltParam) parameterKey() string {
	return "AbsolutePanTilt"
}
func (p AbsolutePanTiltParam) parameterValue() string {
	return commaJoin(hex20Encoder(int(p.Pan)),
		hex20Encoder(int(p.Tilt)),
		itoa(int(p.Speed)))
}
func (AbsolutePanTiltParam) parameterParse(s string) (Parameter, error) {
	sp := commaSplit(s)
	if len(sp) != 3 {
		return nil, fmt.Errorf("invalid comma-joined-list length: %d, expects 3", len(sp))
	}
	pan, err := hex20Decoder(sp[0])
	if err != nil {
		return nil, err
	}
	tilt, err := hex20Decoder(sp[1])
	if err != nil {
		return nil, err
	}
	speed, err := atoi(sp[2])
	if err != nil {
		return nil, err
	}
	return AbsolutePanTiltParam{
		Pan:   SteppedPosition(pan),
		Tilt:  SteppedPosition(tilt),
		Speed: SpeedStep(speed),
	}, nil
}
func (p AbsolutePanTiltParam) Valid() bool {
	return p.Pan.Valid() && p.Tilt.Valid() && p.Speed.Valid()
}
func (AbsolutePanTiltParam) _ptzfParameter() {}
func init() {
	registerParameter(func() Parameter { return AbsolutePanTiltParam{} })
}

type AbsolutePTZFParam struct {
	Pan   SteppedPosition
	Tilt  SteppedPosition
	Zoom  SteppedRange
	Focus SteppedRange
}

func (AbsolutePTZFParam) parameterKey() string {
	return "AbsolutePTZF"
}
func (p AbsolutePTZFParam) parameterValue() string {
	return commaJoin(hex20Encoder(int(p.Pan)),
		hex20Encoder(int(p.Tilt)),
		hex16Encoder(int(p.Zoom)),
		hex16Encoder(int(p.Focus)))
}
func (AbsolutePTZFParam) parameterParse(s string) (Parameter, error) {
	sp := commaSplit(s)
	if len(sp) != 4 {
		return nil, fmt.Errorf("invalid comma-joined-list length: %d, expects 4", len(sp))
	}
	pan, err := hex20Decoder(sp[0])
	if err != nil {
		return nil, err
	}
	tilt, err := hex20Decoder(sp[1])
	if err != nil {
		return nil, err
	}
	zoom, err := hex16Decoder(sp[2])
	if err != nil {
		return nil, err
	}
	focus, err := hex16Decoder(sp[3])
	if err != nil {
		return nil, err
	}
	return AbsolutePTZFParam{
		Pan:   SteppedPosition(pan),
		Tilt:  SteppedPosition(tilt),
		Zoom:  SteppedRange(zoom),
		Focus: SteppedRange(focus),
	}, nil
}
func (p AbsolutePTZFParam) Valid() bool {
	return p.Pan.Valid() && p.Tilt.Valid() && p.Zoom.Valid() && p.Focus.Valid()
}
func (AbsolutePTZFParam) _ptzfParameter() {}
func init() {
	registerParameter(func() Parameter { return AbsolutePTZFParam{} })
}

type FocusMode string

const (
	FocusModeAuto   FocusMode = "auto"
	FocusModeManual FocusMode = "manual"
)

func (f FocusMode) Valid() bool {
	return f == FocusModeAuto || f == FocusModeManual
}

type FocusModeParam struct {
	Mode FocusMode
}

func (FocusModeParam) parameterKey() string {
	return "FocusMode"
}
func (p FocusModeParam) parameterValue() string {
	return string(p.Mode)
}
func (FocusModeParam) parameterParse(s string) (Parameter, error) {
	return FocusModeParam{FocusMode(s)}, nil
}
func (p FocusModeParam) Valid() bool {
	return p.Mode.Valid()
}
func (FocusModeParam) _ptzfParameter() {}
func init() {
	registerParameter(func() Parameter { return FocusModeParam{} })
}

// ZoomSpeed is an arbitrary value between 0 and 32766, the higer the quicker.
type ZoomSpeed int

const ZoomSpeedMax = ZoomSpeed(32766)

func (z ZoomSpeed) Valid() bool {
	return z >= 0 && z <= ZoomSpeedMax
}

// ZoomMoveParam performs a continous zoom motion
type ZoomMoveParam struct {
	Direction ZoomDirection
	Speed     ZoomSpeed
}

func (ZoomMoveParam) parameterKey() string {
	return "ZoomMove"
}
func (p ZoomMoveParam) parameterValue() string {
	return commaJoin(string(p.Direction), itoa(int(p.Speed)))
}
func (ZoomMoveParam) parameterParse(s string) (Parameter, error) {
	sp := commaSplit(s)
	if len(sp) != 2 {
		return nil, fmt.Errorf("invalid comma-joined-list length: %d, expects 2", len(sp))
	}
	dir := sp[0]
	speed, err := atoi(sp[1])
	if err != nil {
		return nil, err
	}
	return ZoomMoveParam{
		Direction: ZoomDirection(dir),
		Speed:     ZoomSpeed(speed),
	}, nil
}
func (p ZoomMoveParam) Valid() bool {
	return p.Direction.Valid() && p.Speed.Valid()
}
func (ZoomMoveParam) _ptzfParameter() {}
func init() {
	registerParameter(func() Parameter { return ZoomMoveParam{} })
}

type PushAFMode string

const (
	PushAFModeAF  PushAFMode = "af"
	PushAFModeAFS PushAFMode = "af-s"
)

func (p PushAFMode) Valid() bool {
	return p == PushAFModeAF || p == PushAFModeAFS
}

type PushAFModeParam struct {
	Mode PushAFMode
}

func (PushAFModeParam) parameterKey() string {
	return "PushAFMode"
}
func (p PushAFModeParam) parameterValue() string {
	return string(p.Mode)
}
func (PushAFModeParam) parameterParse(s string) (Parameter, error) {
	return PushAFModeParam{
		Mode: PushAFMode(s),
	}, nil
}
func (p PushAFModeParam) Valid() bool {
	return p.Mode.Valid()
}
func (PushAFModeParam) _ptzfParameter() {}
func init() {
	registerParameter(func() Parameter { return PushAFModeParam{} })
}

type AbsoluteFocusParam struct {
	Position SteppedRange
}

func (AbsoluteFocusParam) parameterKey() string {
	return "AbsoluteFocus"
}
func (p AbsoluteFocusParam) parameterValue() string {
	return hex16Encoder(int(p.Position))
}
func (AbsoluteFocusParam) parameterParse(s string) (Parameter, error) {
	fp, err := hex16Decoder(s)
	if err != nil {
		return nil, err
	}
	return AbsoluteFocusParam{
		Position: SteppedRange(fp),
	}, nil
}
func (p AbsoluteFocusParam) Valid() bool {
	return p.Position.Valid()
}
func (AbsoluteFocusParam) _ptzfParameter() {}
func init() {
	registerParameter(func() Parameter { return AbsoluteFocusParam{} })
}

type FocusPushAFMFParam struct {
	Button Button
}

func (FocusPushAFMFParam) parameterKey() string {
	return "FocusPushAFMF"
}
func (p FocusPushAFMFParam) parameterValue() string {
	return string(p.Button)
}
func (FocusPushAFMFParam) parameterParse(s string) (Parameter, error) {
	return FocusPushAFMFParam{
		Button: Button(s),
	}, nil
}
func (p FocusPushAFMFParam) Valid() bool {
	return p.Button.Valid()
}
func (FocusPushAFMFParam) _ptzfParameter() {}
func init() {
	registerParameter(func() Parameter { return FocusPushAFMFParam{} })
}

//
// Parameters for presetpositionEndpoint
//

// Sets the pan/tilt speed of presets when SpeedSelect parameter is common.
type CommonSpeedParam struct {
	Speed SpeedStep
}

func (CommonSpeedParam) parameterKey() string {
	return "CommonSpeed"
}
func (c CommonSpeedParam) parameterValue() string {
	return itoa(int(c.Speed))
}
func (CommonSpeedParam) parameterParse(s string) (Parameter, error) {
	i, err := atoi(s)
	if err != nil {
		return nil, err
	}
	return CommonSpeedParam{SpeedStep(i)}, nil
}
func (c CommonSpeedParam) Valid() bool {
	return c.Speed.Valid()
}
func (CommonSpeedParam) _presetpositionParameter() {}
func init() {
	registerParameter(func() Parameter { return CommonSpeedParam{} })
}

// HomePosParam forces a recall of the home position when set.
type HomePosParam struct{}

func (HomePosParam) parameterKey() string {
	return "HomePos"
}
func (HomePosParam) parameterValue() string {
	return "recall"
}
func (HomePosParam) parameterParse(_ string) (Parameter, error) {
	return HomePosParam{}, nil
}
func (HomePosParam) Valid() bool {
	return true
}
func (HomePosParam) _presetpositionParameter() {}
func init() {
	registerParameter(func() Parameter { return HomePosParam{} })
}

type PresetCallParam struct {
	Preset Preset
}

func (PresetCallParam) parameterKey() string {
	return "PresetCall"
}
func (p PresetCallParam) parameterValue() string {
	return itoa(int(p.Preset))
}
func (PresetCallParam) parameterParse(s string) (Parameter, error) {
	i, err := atoi(s)
	if err != nil {
		return nil, err
	}
	return PresetCallParam{
		Preset: Preset(i),
	}, nil
}
func (p PresetCallParam) Valid() bool {
	return p.Preset.Valid()
}
func (PresetCallParam) _presetpositionParameter() {}
func init() {
	registerParameter(func() Parameter { return PresetCallParam{} })
}

type PresetClearParam struct {
	Preset Preset
}

func (PresetClearParam) parameterKey() string {
	return "PresetClear"
}
func (p PresetClearParam) parameterValue() string {
	return itoa(int(p.Preset))
}
func (PresetClearParam) parameterParse(s string) (Parameter, error) {
	i, err := atoi(s)
	if err != nil {
		return nil, err
	}
	return PresetClearParam{
		Preset: Preset(i),
	}, nil
}
func (p PresetClearParam) Valid() bool {
	return p.Preset.Valid()
}
func (PresetClearParam) _presetpositionParameter() {}
func init() {
	registerParameter(func() Parameter { return PresetClearParam{} })
}

type PresetNameParam struct {
	Names map[Preset]PresetName
}

func (PresetNameParam) parameterKey() string {
	return "PresetName"
}
func (p PresetNameParam) parameterValue() string {
	keys := make([]Preset, 0, len(p.Names))
	for k := range p.Names {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	parts := make([]string, len(keys)*2)
	for i, k := range keys {
		parts[i*2] = itoa(int(k))
		name := strings.ReplaceAll(string(p.Names[k]), ",", "")
		parts[i*2+1] = name
	}
	return commaJoin(parts...)
}
func (PresetNameParam) parameterParse(s string) (Parameter, error) {
	sp := commaSplit(s)
	if len(sp)%2 != 0 {
		return nil, fmt.Errorf("invalid comma-joined-list length: %d, expects pairs", len(sp))
	}
	names := make(map[Preset]PresetName, len(sp)/2)
	var errs []error
	for i := 0; i < len(sp)-1; i += 2 {
		no, err := atoi(sp[i])
		if err != nil {
			errs = append(errs, err)
			continue
		}
		names[Preset(no)] = PresetName(sp[i+1])
	}
	return PresetNameParam{
		Names: names,
	}, errors.Join(errs...)
}
func (p PresetNameParam) Valid() bool {
	for k, v := range p.Names {
		if !k.Valid() {
			return false
		}
		if !v.Valid() {
			return false
		}
	}
	return true
}
func (PresetNameParam) _presetpositionParameter() {}
func init() {
	registerParameter(func() Parameter { return PresetNameParam{} })
}

// PresetNumParam returns the maximum number of presets
//
// This is always 100
type PresetNumParam struct {
	Max int
}

func (PresetNumParam) parameterKey() string {
	return "PresetNum"
}
func (p PresetNumParam) parameterValue() string {
	return itoa(p.Max)
}
func (PresetNumParam) parameterParse(s string) (Parameter, error) {
	i, err := atoi(s)
	if err != nil {
		return nil, err
	}
	return PresetNumParam{
		Max: i,
	}, nil
}
func (p PresetNumParam) Valid() bool {
	return p.Max == 100
}
func (PresetNumParam) _presetpositionParameter() {}
func init() {
	registerParameter(func() Parameter { return PresetNumParam{} })
}

// PresetSetParam registers the current pan/tilt/focus/zoom position and camera
// settings as a preset position
type PresetSetParam struct {
	Preset    Preset
	Name      PresetName
	Thumbnail Switch
}

func (PresetSetParam) parameterKey() string {
	return "PresetSet"
}
func (p PresetSetParam) parameterValue() string {
	return commaJoin(itoa(int(p.Preset)), string(p.Name), string(p.Thumbnail))
}
func (PresetSetParam) parameterParse(s string) (Parameter, error) {
	sp := commaSplit(s)
	if len(sp) != 3 {
		return nil, fmt.Errorf("invalid comma-joined-list length: %d, expects 3", len(sp))
	}
	no, err := atoi(sp[0])
	if err != nil {
		return nil, err
	}
	return PresetSetParam{
		Preset:    Preset(no),
		Name:      PresetName(sp[1]),
		Thumbnail: Switch(sp[2]),
	}, nil
}
func (p PresetSetParam) Valid() bool {
	return p.Preset.Valid() && p.Name.Valid() && p.Thumbnail.Valid()
}
func (PresetSetParam) _presetpositionParameter() {}
func init() {
	registerParameter(func() Parameter { return PresetSetParam{} })
}

// PresettThumbnailClearParam removes the thumbnail image of a preset
type PresetThumbnailClearParam struct {
	Preset Preset
}

func (PresetThumbnailClearParam) parameterKey() string {
	return "PresetThumbnailClear"
}
func (p PresetThumbnailClearParam) parameterValue() string {
	return itoa(int(p.Preset))
}
func (PresetThumbnailClearParam) parameterParse(s string) (Parameter, error) {
	i, err := atoi(s)
	if err != nil {
		return nil, err
	}
	return PresetThumbnailClearParam{
		Preset: Preset(i),
	}, nil
}
func (p PresetThumbnailClearParam) Valid() bool {
	return p.Preset.Valid()
}
func (PresetThumbnailClearParam) _presetpositionParameter() {}
func init() {
	registerParameter(func() Parameter { return PresetThumbnailClearParam{} })
}

// SeparateSpeedParam sets the speed of each preset when SpeedSelect is separate.
type SeparateSpeedParam struct {
	Speed map[Preset]SpeedStep
}

func (SeparateSpeedParam) parameterKey() string {
	return "SeparateSpeed"
}
func (p SeparateSpeedParam) parameterValue() string {
	keys := make([]Preset, 0, len(p.Speed))
	for k := range p.Speed {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	parts := make([]string, len(keys)*2)
	for i, k := range keys {
		parts[i*2] = itoa(int(k))
		parts[i*2+1] = itoa(int(p.Speed[k]))
	}
	return commaJoin(parts...)
}
func (SeparateSpeedParam) parameterParse(s string) (Parameter, error) {
	sp := commaSplit(s)
	if len(sp)%2 != 0 {
		return nil, fmt.Errorf("invalid comma-joined-list length: %d, expects pairs", len(sp))
	}
	speeds := make(map[Preset]SpeedStep, len(sp)/2)
	var errs []error
	for i := 0; i < len(sp)-1; i += 2 {
		no, err := atoi(sp[i])
		if err != nil {
			errs = append(errs, err)
			continue
		}
		speed, err := atoi(sp[i+1])
		if err != nil {
			errs = append(errs, err)
			continue
		}
		speeds[Preset(no)] = SpeedStep(speed)
	}
	return SeparateSpeedParam{
		Speed: speeds,
	}, errors.Join(errs...)
}
func (p SeparateSpeedParam) Valid() bool {
	for k, v := range p.Speed {
		if !k.Valid() {
			return false
		}
		if !v.Valid() {
			return false
		}
	}
	return true
}
func (SeparateSpeedParam) _presetpositionParameter() {}
func init() {
	registerParameter(func() Parameter { return SeparateSpeedParam{} })
}

type SpeedSelect string

const (
	SpeedSelectCommon   SpeedSelect = "common"
	SpeedSelectSeparate SpeedSelect = "separate"
)

func (s SpeedSelect) Valid() bool {
	return s == SpeedSelectCommon || s == SpeedSelectSeparate
}

type SpeedSelectParam struct {
	Select SpeedSelect
}

func (SpeedSelectParam) parameterKey() string {
	return "SpeedSelect"
}
func (p SpeedSelectParam) parameterValue() string {
	return string(p.Select)
}
func (SpeedSelectParam) parameterParse(s string) (Parameter, error) {
	return SpeedSelectParam{
		Select: SpeedSelect(s),
	}, nil
}
func (p SpeedSelectParam) Valid() bool {
	return p.Select.Valid()
}
func (SpeedSelectParam) _presetpositionParameter() {}
func init() {
	registerParameter(func() Parameter { return SpeedSelectParam{} })
}

//
// Parameters for NetworkEndpoint
//

type CameraNameParam struct {
	Name string
}

func (CameraNameParam) parameterKey() string {
	return "CameraName"
}
func (p CameraNameParam) parameterValue() string {
	return p.Name
}
func (CameraNameParam) parameterParse(s string) (Parameter, error) {
	return CameraNameParam{
		Name: s,
	}, nil
}
func (p CameraNameParam) Valid() bool {
	return len(p.Name) <= 8
}
func (CameraNameParam) _networkParameter() {}
func init() {
	registerParameter(func() Parameter { return CameraNameParam{} })
}

//
// Parameters for ImagingEndpoint
//

type NDState string

const (
	NDClear    NDState = "clear"
	NDFiltered NDState = "filtered"
)

func (n NDState) Valid() bool {
	return n == NDClear || n == NDFiltered
}

type ExposureNDClearParam struct {
	FilterState NDState
}

func (ExposureNDClearParam) parameterKey() string {
	return "ExposureNDClear"
}
func (p ExposureNDClearParam) parameterValue() string {
	return string(p.FilterState)
}
func (ExposureNDClearParam) parameterParse(s string) (Parameter, error) {
	return ExposureNDClearParam{
		FilterState: NDState(s),
	}, nil
}
func (p ExposureNDClearParam) Valid() bool {
	return p.FilterState.Valid()
}
func (ExposureNDClearParam) _imagingParameter() {}
func init() {
	registerParameter(func() Parameter { return ExposureNDClearParam{} })
}

type NDLevel int

const (
	ND1o4 NDLevel = iota
	ND1o5
	ND1o6
	ND1o7
	ND1o8
	ND1o10
	ND1o11
	ND1o13
	ND1o16
	ND1o19
	ND1o23
	ND1o27
	ND1o32
	ND1o38
	ND1o45
	ND1o54
	ND1o64
	ND1o76
	ND1o91
	ND1o108
	ND1o128
)

func (n NDLevel) Valid() bool {
	return (n >= ND1o4) && (n <= ND1o128)
}

type ExposureNDVariableParam struct {
	Level NDLevel
}

func (ExposureNDVariableParam) parameterKey() string {
	return "ExposureNDVariable"
}
func (p ExposureNDVariableParam) parameterValue() string {
	return itoa(int(p.Level))
}
func (ExposureNDVariableParam) parameterParse(s string) (Parameter, error) {
	i, err := atoi(s)
	if err != nil {
		return nil, err
	}
	return ExposureNDVariableParam{
		Level: NDLevel(i),
	}, nil
}
func (p ExposureNDVariableParam) Valid() bool {
	return p.Level.Valid()
}
func (ExposureNDVariableParam) _imagingParameter() {}
func init() {
	registerParameter(func() Parameter { return ExposureNDVariableParam{} })
}

type ExposureAutoIrisParam struct {
	Auto Switch
}

func (ExposureAutoIrisParam) parameterKey() string {
	return "ExposureAutoIris"
}

func (p ExposureAutoIrisParam) parameterValue() string {
	return string(p.Auto)
}

func (ExposureAutoIrisParam) parameterParse(s string) (Parameter, error) {
	return ExposureAutoIrisParam{
		Auto: Switch(s),
	}, nil
}

func (p ExposureAutoIrisParam) Valid() bool {
	return p.Auto.Valid()
}

func (ExposureAutoIrisParam) _imagingParameter() {}

func init() {
	registerParameter(func() Parameter { return ExposureAutoIrisParam{} })
}

type ExposureIrisRangeParam struct {
	Min, Max int
}

func (ExposureIrisRangeParam) parameterKey() string {
	return "ExposureIrisRange"
}

func (p ExposureIrisRangeParam) parameterValue() string {
	return itoa(p.Min) + ":" + itoa(p.Max)
}

func (ExposureIrisRangeParam) parameterParse(s string) (Parameter, error) {
	parts := commaSplit(s)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid comma-joined-list length: %d, expects 2", len(parts))
	}
	min, err := atoi(parts[0])
	if err != nil {
		return nil, err
	}
	max, err := atoi(parts[1])
	if err != nil {
		return nil, err
	}
	return ExposureIrisRangeParam{
		Min: min,
		Max: max,
	}, nil
}

func (p ExposureIrisRangeParam) Valid() bool {
	return p.Min <= p.Max
}

func (ExposureIrisRangeParam) _imagingParameter() {}

func init() {
	registerParameter(func() Parameter { return ExposureIrisRangeParam{} })
}

type ExposureIrisParam struct {
	Iris int
}

func (ExposureIrisParam) parameterKey() string {
	return "ExposureIris"
}

func (p ExposureIrisParam) parameterValue() string {
	return itoa(p.Iris)
}

func (ExposureIrisParam) parameterParse(s string) (Parameter, error) {
	iris, err := atoi(s)
	if err != nil {
		return nil, err
	}
	return ExposureIrisParam{
		Iris: iris,
	}, nil
}

func (p ExposureIrisParam) Valid() bool {
	return p.Iris >= 0 && p.Iris <= 65535
}

func (ExposureIrisParam) _imagingParameter() {}

func init() {
	registerParameter(func() Parameter { return ExposureIrisParam{} })
}

type WhiteBalanceColorTempParam struct {
	Kelvin int
}

func (p WhiteBalanceColorTempParam) parameterKey() string {
	return "WhiteBalanceColorTemp"
}

func (p WhiteBalanceColorTempParam) parameterValue() string {
	return itoa(p.Kelvin)
}

func (WhiteBalanceColorTempParam) parameterParse(s string) (Parameter, error) {
	kelvin, err := atoi(s)
	if err != nil {
		return nil, err
	}
	return WhiteBalanceColorTempParam{
		Kelvin: kelvin,
	}, nil
}

func (p WhiteBalanceColorTempParam) Valid() bool {
	return p.Kelvin >= 2000 && p.Kelvin <= 15000
}

func (WhiteBalanceColorTempParam) _imagingParameter() {}

func init() {
	registerParameter(func() Parameter { return WhiteBalanceColorTempParam{} })
}

type WhiteBalanceTintParam struct {
	Tint int
}

func (p WhiteBalanceTintParam) parameterKey() string {
	return "WhiteBalanceTint"
}

func (p WhiteBalanceTintParam) parameterValue() string {
	return itoa(p.Tint)
}

func (WhiteBalanceTintParam) parameterParse(s string) (Parameter, error) {
	tint, err := atoi(s)
	if err != nil {
		return nil, err
	}
	return WhiteBalanceTintParam{
		Tint: tint,
	}, nil
}

func (p WhiteBalanceTintParam) Valid() bool {
	return p.Tint >= -99 && p.Tint <= 99
}

func (WhiteBalanceTintParam) _imagingParameter() {}

func init() {
	registerParameter(func() Parameter { return WhiteBalanceTintParam{} })
}

type WhiteBalanceCbGain struct {
	Gain int
}

func (p WhiteBalanceCbGain) parameterKey() string {
	return "WhiteBalanceCbGain"
}

func (p WhiteBalanceCbGain) parameterValue() string {
	return itoa(p.Gain)
}

func (WhiteBalanceCbGain) parameterParse(s string) (Parameter, error) {
	gain, err := atoi(s)
	if err != nil {
		return nil, err
	}
	return WhiteBalanceCbGain{
		Gain: gain,
	}, nil
}

func (p WhiteBalanceCbGain) Valid() bool {
	return p.Gain >= -990 && p.Gain <= 990
}

func (p WhiteBalanceCbGain) _imagingParameter() {}

func init() {
	registerParameter(func() Parameter { return WhiteBalanceCbGain{} })
}

type WhiteBalanceCrGain struct {
	Gain int
}

func (p WhiteBalanceCrGain) parameterKey() string {
	return "WhiteBalanceCrGain"
}

func (p WhiteBalanceCrGain) parameterValue() string {
	return itoa(p.Gain)
}

func (WhiteBalanceCrGain) parameterParse(s string) (Parameter, error) {
	gain, err := atoi(s)
	if err != nil {
		return nil, err
	}
	return WhiteBalanceCrGain{
		Gain: gain,
	}, nil
}

func (p WhiteBalanceCrGain) Valid() bool {
	return p.Gain >= -990 && p.Gain <= 990
}

func (p WhiteBalanceCrGain) _imagingParameter() {}

func init() {
	registerParameter(func() Parameter { return WhiteBalanceCrGain{} })
}
