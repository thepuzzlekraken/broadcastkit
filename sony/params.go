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
	BoolOn  Switch = "on"
	BoolOff Switch = "off"
)

func (s Switch) Valid() bool {
	switch s {
	case BoolOn, BoolOff:
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
		parts[i*2+1] = string(p.Names[k])
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
