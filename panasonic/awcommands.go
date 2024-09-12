package panasonic

import "strings"

// PowerSwitch represent the status of the virtual power switch
type PowerSwitch byte

const (
	PowerStandby       PowerSwitch = '0'
	PowerOn            PowerSwitch = '1'
	PowerTransitioning PowerSwitch = '3'
)

// AWPower command manages the Power On / Standby state of the camera.
type AWPower struct {
	Power   PowerSwitch
	altMode bool
}

func init() { registerRequest(func() AWRequest { return &AWPower{} }) }
func init() { registerResponse(func() AWResponse { return &AWPower{} }) }
func (a *AWPower) Acceptable() bool {
	if a.Power == PowerOn || a.Power == PowerStandby {
		return true
	}
	return false
}
func (a *AWPower) Response() AWResponse {
	return a
}
func (a *AWPower) requestSignature() string {
	return "#O\x00"
}
func (a *AWPower) unpackRequest(cmd string) {
	c := cmd[2]
	switch c {
	case 'f':
		a.altMode = true
		c = '0'
	case 'n':
		a.altMode = true
		c = '1'
	default:
		a.altMode = false
	}
	a.Power = PowerSwitch(c)
}
func (a *AWPower) packRequest() string {
	c := a.Power
	if a.altMode {
		// preserve alternate-mode request to allow for transparent proxying
		switch a.Power {
		case PowerOn:
			c = 'n'
		case PowerStandby:
			c = 'f'
		}
	}
	return "#O" + string(c)
}

func (a *AWPower) responseSignature() string {
	return "p\x00"
}
func (a *AWPower) unpackResponse(cmd string) {
	a.Power = PowerSwitch(cmd[1])
}
func (a *AWPower) packResponse() string {
	c := a.Power
	if !matchSets[anySet].contains(byte(c)) {
		// invalid bytes in go are wider than invalid bytes in the protocol
		// map those non-representable bytes to x to preserve invalid behavior
		c = 'x'
	}
	// note: the camare itself always responds with numbers, even for alternate
	// mode requests. We mirror that behavior here.
	return "p" + string(a.Power)
}

type AWPowerQuery struct{}

func init() { registerRequest(func() AWRequest { return &AWPowerQuery{} }) }
func (a *AWPowerQuery) Acceptable() bool {
	return true
}
func (a *AWPowerQuery) Response() AWResponse {
	return &AWPower{}
}
func (a *AWPowerQuery) requestSignature() string {
	return "#O"
}
func (a *AWPowerQuery) unpackRequest(_ string) {}
func (a *AWPowerQuery) packRequest() string {
	return "#O"
}

// InstallSwitch represents the installation position of the camera
type InstallSwitch int

const (
	DesktopPosition InstallSwitch = 0 // standing on base
	HangingPosition InstallSwitch = 1 // hanging from base
)

// AWInstall configures the installation position of the camera.
//
// This setting impacts image orientation and continous control directions
type AWInstall struct {
	Position InstallSwitch
}

func init() { registerRequest(func() AWRequest { return &AWInstall{} }) }
func init() { registerResponse(func() AWResponse { return &AWInstall{} }) }
func (a *AWInstall) Acceptable() bool {
	switch a.Position {
	case DesktopPosition, HangingPosition:
		return true
	}
	return false
}
func (a *AWInstall) Response() AWResponse {
	return a
}
func (a *AWInstall) requestSignature() string {
	return "#INS\x02"
}
func (a *AWInstall) unpackRequest(cmd string) {
	a.Position = InstallSwitch(dec2int(cmd[4:5]))
}
func (a *AWInstall) packRequest() string {
	return "#INS" + int2dec(int(a.Position), 1)
}

func (a *AWInstall) responseSignature() string {
	return "iNS\x02"
}
func (a *AWInstall) unpackResponse(cmd string) {
	a.Position = InstallSwitch(dec2int(cmd[3:4]))
}
func (a *AWInstall) packResponse() string {
	return "iNS" + int2dec(int(a.Position), 1)
}

// AWInstallQuery queries the configured installation position.
//
// See AWInstall for response format.
type AWInstallQuery struct{}

func init() { registerRequest(func() AWRequest { return &AWInstallQuery{} }) }
func (a *AWInstallQuery) Acceptable() bool {
	return true
}
func (a *AWInstallQuery) Response() AWResponse {
	return &AWInstall{}
}
func (a *AWInstallQuery) requestSignature() string {
	return "#INS"
}
func (a *AWInstallQuery) unpackRequest(_ string) {}
func (a *AWInstallQuery) packRequest() string {
	return "#INS"
}

// MoveUnit represents the unit of pan or tilt movement for the camera.
//
// Following the go convention zero-value is the home. One degree rotation is
// approx 121.35 units. Sign of displacement follows right-hand convention
// (right:+ left:-) of the viewer when the camera is in DesktopPosition.
type MoveUnit int

func (m MoveUnit) toWire() string {
	// offset and inversion for Panasonic home 0 => 0x8000
	return int2hex(-int(m)+0x8000, 4)
}
func toMoveUnit(hex string) MoveUnit {
	// offset and inversion from Panasonic home 0x8000 => 0
	return MoveUnit(-hex2int(hex[0:4]) + 0x8000)
}
func (m MoveUnit) Acceptable() bool {
	// Despite the camera range being limited in reality, cameras report
	// acceptable for any value, and just stop at the end of real range.
	return true
}

// AWPanTiltTo command manages the absolute pan and tilt position of the camera
type AWPanTiltTo struct {
	Pan  MoveUnit
	Tilt MoveUnit
}

func init() { registerRequest(func() AWRequest { return &AWPanTiltTo{} }) }
func init() { registerResponse(func() AWResponse { return &AWPanTiltTo{} }) }
func (a *AWPanTiltTo) Acceptable() bool {
	return a.Pan.Acceptable() && a.Tilt.Acceptable()
}
func (a *AWPanTiltTo) Response() AWResponse {
	return a
}
func (a *AWPanTiltTo) requestSignature() string {
	return "#APC\x01\x01\x01\x01\x01\x01\x01\x01"
}
func (a *AWPanTiltTo) unpackRequest(cmd string) {
	_ = cmd[11]
	a.Pan = toMoveUnit(cmd[4:8])
	a.Tilt = toMoveUnit(cmd[8:12])
}
func (a *AWPanTiltTo) packRequest() string {
	return "#APC" + a.Pan.toWire() + a.Tilt.toWire()
}

func (a *AWPanTiltTo) responseSignature() string {
	// #APC not supported in awNty notifications unfortunately
	return "aPC\x01\x01\x01\x01\x01\x01\x01\x01"
}
func (a *AWPanTiltTo) unpackResponse(cmd string) {
	_ = cmd[10]
	a.Pan = toMoveUnit(cmd[3:7])
	a.Tilt = toMoveUnit(cmd[7:11])
}
func (a *AWPanTiltTo) packResponse() string {
	return "aPC" + a.Pan.toWire() + a.Tilt.toWire()
}

// AWPanTiltQuery requests the current pan and tilt position of the camera.
//
// Note that cameras typically report a coordinates which only approximately
// match their commanded position. If the camera is commanded to move it's
// reported position, it may also move to a different approximation of it.
type AWPanTiltQuery struct{}

func init() { registerRequest(func() AWRequest { return &AWPanTiltQuery{} }) }
func (a *AWPanTiltQuery) Acceptable() bool {
	return true
}
func (a *AWPanTiltQuery) Response() AWResponse {
	return &AWPanTiltTo{}
}
func (a *AWPanTiltQuery) requestSignature() string {
	return "#APC"
}
func (a *AWPanTiltQuery) unpackRequest(_ string) {}
func (a *AWPanTiltQuery) packRequest() string {
	return "#APC"
}

// SpeedUnit is the arbitrary unit of speed for Panasonic cameras.
//
// Speed is a value between 1 and 30, the higher the quicker.
// Zero Speed chooses the factory default speed of 10.
// Table is a firmware-defined lookup table for interpreting the speed value.
// The Zero table chooses the factory default FastTable
type SpeedUnit struct {
	Speed int
	Table SpeedTable
}

// SpeedTable is the lookup table used by SpeedUnit
type SpeedTable int

func (t SpeedTable) toWire() string {
	tb := t
	if tb == DefaultSpeed {
		tb = FastSpeed
	}
	return int2dec(int(tb)-1, 1)
}

func toSpeedTable(dec string) SpeedTable {
	return SpeedTable(dec2int(dec[0:1]) + 1)
}

func (t SpeedTable) Acceptable() bool {
	return t >= DefaultSpeed && t <= FastSpeed
}

const (
	// Values are offset by one to Panasonic definition, to allow for default
	DefaultSpeed SpeedTable = 0
	SlowSpeed    SpeedTable = 1
	MedSpeed     SpeedTable = 2
	FastSpeed    SpeedTable = 3
)

func (s SpeedUnit) Acceptable() bool {
	return s.Speed >= 0 && s.Speed <= 30 && s.Table.Acceptable()
}
func (s SpeedUnit) toWire() string {
	sp := s.Speed
	if sp == 0 {
		sp = 9
	}
	return int2hex(sp, 2) + s.Table.toWire()
}
func toSpeedUnit(data string) SpeedUnit {
	_ = data[2]
	return SpeedUnit{
		Speed: int(hex2int(data[0:2])),
		Table: toSpeedTable(data[2:3]),
	}
}

// AWPanTiltSpeedTo command manages the absolute pan and tilt position to be
// reached at a given speed. See MoveUnit and SpeedUnit for details on values
type AWPanTiltSpeedTo struct {
	Pan   MoveUnit
	Tilt  MoveUnit
	Speed SpeedUnit
}

func init() { registerRequest(func() AWRequest { return &AWPanTiltSpeedTo{} }) }
func init() { registerResponse(func() AWResponse { return &AWPanTiltSpeedTo{} }) }
func (a *AWPanTiltSpeedTo) Acceptable() bool {
	return a.Pan.Acceptable() && a.Tilt.Acceptable() && a.Speed.Acceptable()
}
func (a *AWPanTiltSpeedTo) Response() AWResponse {
	return a
}
func (a *AWPanTiltSpeedTo) requestSignature() string {
	return "#APS\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01\x02"
}
func (a *AWPanTiltSpeedTo) unpackRequest(cmd string) {
	_ = cmd[14]
	a.Pan = toMoveUnit(cmd[4:8])
	a.Tilt = toMoveUnit(cmd[8:12])
	a.Speed = toSpeedUnit(cmd[12:15])
}
func (a *AWPanTiltSpeedTo) packRequest() string {
	return "#APS" + a.Pan.toWire() + a.Tilt.toWire() + a.Speed.toWire()
}

func (a *AWPanTiltSpeedTo) responseSignature() string {
	// #APS not supported in awNty notifications unfortunately
	return "aPS\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01\x02"
}
func (a *AWPanTiltSpeedTo) unpackResponse(cmd string) {
	_ = cmd[13]
	a.Pan = toMoveUnit(cmd[3:7])
	a.Tilt = toMoveUnit(cmd[7:11])
	a.Speed = toSpeedUnit(cmd[11:14])
}
func (a *AWPanTiltSpeedTo) packResponse() string {
	return "aPS" + a.Pan.toWire() + a.Tilt.toWire() + a.Speed.toWire()
}

// AWPanTiltBy commands a movement relative to the cameras current
// position. See MoveUnit for details on values
type AWPanTiltBy struct {
	Pan  MoveUnit
	Tilt MoveUnit
}

func init() { registerRequest(func() AWRequest { return &AWPanTiltBy{} }) }
func init() { registerResponse(func() AWResponse { return &AWPanTiltBy{} }) }
func (a *AWPanTiltBy) Acceptable() bool {
	return a.Pan.Acceptable() && a.Tilt.Acceptable()
}
func (a *AWPanTiltBy) Response() AWResponse {
	return a
}
func (a *AWPanTiltBy) requestSignature() string {
	return "#RPC\x01\x01\x01\x01\x01\x01\x01\x01"
}
func (a *AWPanTiltBy) unpackRequest(cmd string) {
	_ = cmd[11]
	a.Pan = toMoveUnit(cmd[4:8])
	a.Tilt = toMoveUnit(cmd[8:12])
}
func (a *AWPanTiltBy) packRequest() string {
	return "#RPC" + a.Pan.toWire() + a.Tilt.toWire()
}

func (a *AWPanTiltBy) responseSignature() string {
	// #RPC not supported in awNty notifications unfortunately
	return "rPC\x01\x01\x01\x01\x01\x01\x01\x01"
}
func (a *AWPanTiltBy) unpackResponse(cmd string) {
	_ = cmd[10]
	a.Pan = toMoveUnit(cmd[3:7])
	a.Tilt = toMoveUnit(cmd[7:11])
}
func (a *AWPanTiltBy) packResponse() string {
	return "rPC" + a.Pan.toWire() + a.Tilt.toWire()
}

// AWPanTiltSpeedBy commands a movement relative to the cameras current
// position via a given speed. See MoveUnit and SpeedUnit for details on values.
type AWPanTiltSpeedBy struct {
	Pan   MoveUnit
	Tilt  MoveUnit
	Speed SpeedUnit
}

func init() { registerRequest(func() AWRequest { return &AWPanTiltSpeedBy{} }) }
func init() { registerResponse(func() AWResponse { return &AWPanTiltSpeedBy{} }) }
func (a *AWPanTiltSpeedBy) Acceptable() bool {
	return a.Pan.Acceptable() && a.Tilt.Acceptable() && a.Speed.Acceptable()
}
func (a *AWPanTiltSpeedBy) Response() AWResponse {
	return a
}
func (a *AWPanTiltSpeedBy) requestSignature() string {
	return "#RPS\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01\x02"
}
func (a *AWPanTiltSpeedBy) unpackRequest(cmd string) {
	_ = cmd[14]
	a.Pan = toMoveUnit(cmd[4:8])
	a.Tilt = toMoveUnit(cmd[8:12])
	a.Speed = toSpeedUnit(cmd[12:15])
}
func (a *AWPanTiltSpeedBy) packRequest() string {
	return "#RPS" + a.Pan.toWire() + a.Tilt.toWire() + a.Speed.toWire()
}

func (a *AWPanTiltSpeedBy) responseSignature() string {
	// #RPS not supported in awNty notifications
	return "rPS\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01\x02"
}
func (a *AWPanTiltSpeedBy) unpackResponse(cmd string) {
	_ = cmd[13]
	a.Pan = toMoveUnit(cmd[3:7])
	a.Tilt = toMoveUnit(cmd[7:11])
	a.Speed = toSpeedUnit(cmd[11:14])
}
func (a *AWPanTiltSpeedBy) packResponse() string {
	return "rPS" + a.Pan.toWire() + a.Tilt.toWire() + a.Speed.toWire()
}

// ContinuousSpeed is an arbitrary speed value for a continuous movement
//
// Zero value commands a stop
// Negative values move leftwards or downwards
// Positive values move rightwards or upwards
// Directions respect the actual AWInstall configuration
// Maximum values are +/- 49, values outside the range cause ErrUnacceptable
type ContinuousSpeed int

func (c ContinuousSpeed) toWire() string {
	if !c.Acceptable() {
		return "00"
	}
	return int2dec(int(c)+50, 2)
}
func toInteractiveSpeed(s string) ContinuousSpeed {
	return ContinuousSpeed(dec2int(s[0:2]) - 50)
}
func (c ContinuousSpeed) Acceptable() bool {
	if c < -49 || c > 49 {
		return false
	}
	return true
}

// AWPan commands a continuous pan movement
//
// See ContinuousSpeed for details. To coordinate Pan and Tilt movements,
// prefer AWPanTilt instead.
type AWPan struct {
	Pan ContinuousSpeed
}

func init() { registerRequest(func() AWRequest { return &AWPan{} }) }
func init() { registerResponse(func() AWResponse { return &AWPan{} }) }
func (a *AWPan) Acceptable() bool {
	return a.Pan.Acceptable()
}
func (a *AWPan) Response() AWResponse {
	return a
}
func (a *AWPan) requestSignature() string {
	return "#P\x02\x02"
}
func (a *AWPan) unpackRequest(cmd string) {
	a.Pan = toInteractiveSpeed(cmd[2:4])
}
func (a *AWPan) packRequest() string {
	return "#P" + a.Pan.toWire()
}

func (a *AWPan) responseSignature() string {
	return "pS\x02\x02"
}
func (a *AWPan) unpackResponse(cmd string) {
	a.Pan = toInteractiveSpeed(cmd[2:4])
}
func (a *AWPan) packResponse() string {
	return "pS" + a.Pan.toWire()
}

// AWTilt commands a continuous tilt movement.
// See ContinuousSpeed for details. To coordinate Pan and Tilt movements,
// prefer AWPanTilt instead.
type AWTilt struct {
	Tilt ContinuousSpeed
}

func init() { registerRequest(func() AWRequest { return &AWTilt{} }) }
func init() { registerResponse(func() AWResponse { return &AWTilt{} }) }
func (a *AWTilt) Acceptable() bool {
	return a.Tilt.Acceptable()
}
func (a *AWTilt) Response() AWResponse {
	return a
}
func (a *AWTilt) requestSignature() string {
	return "#T\x02\x02"
}
func (a *AWTilt) unpackRequest(cmd string) {
	a.Tilt = toInteractiveSpeed(cmd[2:4])
}
func (a *AWTilt) packRequest() string {
	return "#T" + a.Tilt.toWire()
}

func (a *AWTilt) responseSignature() string {
	return "tS\x02\x02"
}
func (a *AWTilt) unpackResponse(cmd string) {
	a.Tilt = toInteractiveSpeed(cmd[2:4])
}
func (a *AWTilt) packResponse() string {
	return "tS" + a.Tilt.toWire()
}

// AWPanTilt commands a continuous pan and tilt movement.
// This is the roughly the same as AWPan and AWTilt, but avoids the timing
// issues of commanding them separately.
// See ContinuousSpeed for details.
type AWPanTilt struct {
	Pan  ContinuousSpeed
	Tilt ContinuousSpeed
}

func init() { registerRequest(func() AWRequest { return &AWPanTilt{} }) }
func init() { registerResponse(func() AWResponse { return &AWPanTilt{} }) }
func (a *AWPanTilt) Acceptable() bool {
	return a.Pan.Acceptable() && a.Tilt.Acceptable()
}
func (a *AWPanTilt) Response() AWResponse {
	return a
}
func (a *AWPanTilt) requestSignature() string {
	return "#PTS\x02\x02\x02\x02"
}
func (a *AWPanTilt) unpackRequest(cmd string) {
	_ = cmd[7]
	a.Pan = toInteractiveSpeed(cmd[4:6])
	a.Tilt = toInteractiveSpeed(cmd[6:8])
}
func (a *AWPanTilt) packRequest() string {
	return "#PTS" + a.Pan.toWire() + a.Tilt.toWire()
}

func (a *AWPanTilt) responseSignature() string {
	return "pTS\x02\x02\x02\x02"
}
func (a *AWPanTilt) unpackResponse(cmd string) {
	_ = cmd[6]
	a.Pan = toInteractiveSpeed(cmd[3:5])
	a.Tilt = toInteractiveSpeed(cmd[5:7])
}
func (a *AWPanTilt) packResponse() string {
	return "pTS" + a.Pan.toWire() + a.Tilt.toWire()
}

// ScaleUnit indicates a position on a preset scale
//
// Zero is the "near-end" of the range: as near-focus, wide-angle, closed-iris
// 4095 is the "far-end" of the range: as far-focus, tele-angle, open-iris
type ScaleUnit int

func (s ScaleUnit) toWire() string {
	if s > 4095 {
		// Avoid invalid high values becoming a valid FFF
		return "000"
	}
	// Panasonic API uses the range 0x555 to 0xFFF
	return int2hex(int(s)+0x555, 3)
}

func toScaleUnit(s string) ScaleUnit {
	return ScaleUnit(hex2int(s[0:3]) - 0x555)
}

func (s ScaleUnit) Acceptable() bool {
	if s < 0 || s > 4095 {
		return false
	}
	return true
}

// AWZoomTo commands a zooms to a specific position on the scale
type AWZoomTo struct {
	Zoom ScaleUnit
}

func init() { registerRequest(func() AWRequest { return &AWZoomTo{} }) }
func init() { registerResponse(func() AWResponse { return &AWZoomTo{} }) }

func (a *AWZoomTo) Acceptable() bool {
	return a.Zoom.Acceptable()
}
func (a *AWZoomTo) Response() AWResponse {
	return a
}
func (a *AWZoomTo) requestSignature() string {
	return "#AXZ\x01\x01\x01"
}
func (a *AWZoomTo) unpackRequest(cmd string) {
	a.Zoom = toScaleUnit(cmd[4:7])
}
func (a *AWZoomTo) packRequest() string {
	return "#AXZ" + a.Zoom.toWire()
}
func (a *AWZoomTo) responseSignature() string {
	return "axz\x01\x01\x01"
}
func (a *AWZoomTo) unpackResponse(cmd string) {
	a.Zoom = toScaleUnit(cmd[3:6])
}
func (a *AWZoomTo) packResponse() string {
	return "axz" + a.Zoom.toWire()
}

// AWZoomQuery is a request for the current zoom position.
type AWZoomQuery struct{}

func init() { registerRequest(func() AWRequest { return &AWZoomQuery{} }) }
func (a *AWZoomQuery) Acceptable() bool {
	return true
}
func (a *AWZoomQuery) Response() AWResponse {
	return &AWZoomTo{}
}
func (a *AWZoomQuery) requestSignature() string {
	return "#AXZ"
}
func (a *AWZoomQuery) unpackRequest(_ string) {}
func (a *AWZoomQuery) packRequest() string {
	return "#AXZ"
}

// AWZoomResponseAlternate is the answer to AWZoomQuery requests
//
// This response is functionally equivalent to an AWZoomTo response, but has
// different on-wire format. This is yielded by AWZoomQueryAlternate and
type AWZoomResponseAlternate struct {
	Zoom ScaleUnit
}

func init() { registerResponse(func() AWResponse { return &AWZoomResponseAlternate{} }) }

func (a *AWZoomResponseAlternate) responseSignature() string {
	// There's a special case of gz--- which is returned instead of an eR2 error
	// when the camera is suspended. We'll just let that for UnknownResponse.
	return "gz\x01\x01\x01"
}
func (a *AWZoomResponseAlternate) unpackResponse(cmd string) {
	a.Zoom = toScaleUnit(cmd[2:5])
}
func (a *AWZoomResponseAlternate) packResponse() string {
	return "gz" + a.Zoom.toWire()
}

// AWZoomQueryAltenate requests informationabout the current zoom level.
//
// This is functionally equivalent to an AWZoomQuery request, but it is sent as
// a different command to the camera. Yields AWZoomResponseAlternate instead of
// AWZoomTo.
type AWZoomQueryAltenate struct{}

func init() { registerRequest(func() AWRequest { return &AWZoomQueryAltenate{} }) }
func (a *AWZoomQueryAltenate) Acceptable() bool {
	return true
}
func (a *AWZoomQueryAltenate) Response() AWResponse {
	return &AWZoomResponseAlternate{}
}
func (a *AWZoomQueryAltenate) requestSignature() string {
	return "#GZ"
}
func (a *AWZoomQueryAltenate) unpackRequest(cmd string) {}
func (a *AWZoomQueryAltenate) packRequest() string {
	return "#GZ"
}

// AWZoom commands a continuous zoom movement with a given speed.
type AWZoom struct {
	Zoom ContinuousSpeed
}

func init() { registerRequest(func() AWRequest { return &AWZoom{} }) }
func init() { registerResponse(func() AWResponse { return &AWZoom{} }) }
func (a *AWZoom) Acceptable() bool {
	return a.Zoom.Acceptable()
}
func (a *AWZoom) Response() AWResponse {
	return a
}
func (a *AWZoom) requestSignature() string {
	return "#Z\x02\x02"
}
func (a *AWZoom) unpackRequest(cmd string) {
	a.Zoom = toInteractiveSpeed(cmd[2:4])
}
func (a *AWZoom) packRequest() string {
	return "#Z" + a.Zoom.toWire()
}

func (a *AWZoom) responseSignature() string {
	return "zS\x02\x02"
}
func (a *AWZoom) unpackResponse(cmd string) {
	a.Zoom = toInteractiveSpeed(cmd[2:4])
}
func (a *AWZoom) packResponse() string {
	return "zS" + a.Zoom.toWire()
}

// AWFocusTo commands a focus movement to a specific position on the scale.
type AWFocusTo struct {
	Focus ScaleUnit
}

func init() { registerRequest(func() AWRequest { return &AWFocusTo{} }) }
func init() { registerResponse(func() AWResponse { return &AWFocusTo{} }) }

func (a *AWFocusTo) Acceptable() bool {
	return a.Focus.Acceptable()
}
func (a *AWFocusTo) Response() AWResponse {
	return a
}
func (a *AWFocusTo) requestSignature() string {
	return "#AXF\x01\x01\x01"
}
func (a *AWFocusTo) unpackRequest(cmd string) {
	a.Focus = toScaleUnit(cmd[4:7])
}
func (a *AWFocusTo) packRequest() string {
	return "#AXF" + a.Focus.toWire()
}
func (a *AWFocusTo) responseSignature() string {
	return "axf\x01\x01\x01"
}
func (a *AWFocusTo) packResponse() string {
	return "axf" + a.Focus.toWire()
}
func (a *AWFocusTo) unpackResponse(cmd string) {
	a.Focus = toScaleUnit(cmd[3:6])
}

// AWFocusQuery is a request for the current focus position.
type AWFocusQuery struct{}

func init() { registerRequest(func() AWRequest { return &AWFocusQuery{} }) }
func (a *AWFocusQuery) Acceptable() bool {
	return true
}
func (a *AWFocusQuery) Response() AWResponse {
	return &AWFocusTo{}
}
func (a *AWFocusQuery) requestSignature() string {
	return "#AXF"
}
func (a *AWFocusQuery) unpackRequest(_ string) {}
func (a *AWFocusQuery) packRequest() string {
	return "#AXF"
}

// AWFocusResponseAlternate is the answer to AWFocusQueryAlternate requests
type AWFocusResponseAlternate struct {
	Focus ScaleUnit
}

func init() { registerResponse(func() AWResponse { return &AWFocusResponseAlternate{} }) }

func (a *AWFocusResponseAlternate) responseSignature() string {
	// There's a special case of gz--- which is returned instead of an eR2 error
	// when the camera is suspended. We'll just leave that for UnknownResponse.
	return "gf\x01\x01\x01"
}
func (a *AWFocusResponseAlternate) unpackResponse(cmd string) {
	a.Focus = toScaleUnit(cmd[2:5])
}
func (a *AWFocusResponseAlternate) packResponse() string {
	return "gf" + a.Focus.toWire()
}

// AWFocusQueryAlternate requests informationabout the current focus position.
// This is functionally equivalent to an AWFocusQuery request, but it is sent as
// a different command to the camera. Yields AWFocusResponseAlternate instead of
// AWFocusTo.
type AWFocusQueryAlternate struct{}

func init() { registerRequest(func() AWRequest { return &AWFocusQueryAlternate{} }) }
func (a *AWFocusQueryAlternate) Acceptable() bool {
	return true
}
func (a *AWFocusQueryAlternate) Response() AWResponse {
	return &AWFocusResponseAlternate{}
}
func (a *AWFocusQueryAlternate) requestSignature() string {
	return "#GF"
}
func (a *AWFocusQueryAlternate) unpackRequest(_ string) {}
func (a *AWFocusQueryAlternate) packRequest() string {
	return "#GF"
}

// AWFocus commands a continuous focus movement with a given speed.
type AWFocus struct {
	Focus ContinuousSpeed
}

func init() { registerRequest(func() AWRequest { return &AWFocus{} }) }
func init() { registerResponse(func() AWResponse { return &AWFocus{} }) }
func (a *AWFocus) Acceptable() bool {
	return a.Focus.Acceptable()
}
func (a *AWFocus) Response() AWResponse {
	return a
}

func (a *AWFocus) requestSignature() string {
	return "#F\x02\x02"
}
func (a *AWFocus) unpackRequest(cmd string) {
	a.Focus = toInteractiveSpeed(cmd[2:4])
}
func (a *AWFocus) packRequest() string {
	return "#F" + a.Focus.toWire()
}
func (a *AWFocus) responseSignature() string {
	return "fS\x02\x02"
}
func (a *AWFocus) unpackResponse(cmd string) {
	a.Focus = toInteractiveSpeed(cmd[2:4])
}
func (a *AWFocus) packResponse() string {
	return "fS" + a.Focus.toWire()
}

// AWAutoFocus configures the camera's autofocus functionality.
type AWAutoFocus struct {
	AutoFocus Toggle
}

func init() { registerRequest(func() AWRequest { return &AWAutoFocus{} }) }
func init() { registerResponse(func() AWResponse { return &AWAutoFocus{} }) }
func (a *AWAutoFocus) Acceptable() bool {
	return a.AutoFocus.Acceptable()
}
func (a *AWAutoFocus) Response() AWResponse {
	return a
}

func (a *AWAutoFocus) requestSignature() string {
	return "#D1\x02"
}
func (a *AWAutoFocus) unpackRequest(cmd string) {
	a.AutoFocus = toToggle(cmd[3:4])
}
func (a *AWAutoFocus) packRequest() string {
	return "#D1" + a.AutoFocus.toWire()
}
func (a *AWAutoFocus) responseSignature() string {
	return "d1\x02"
}
func (a *AWAutoFocus) unpackResponse(cmd string) {
	a.AutoFocus = toToggle(cmd[2:3])
}
func (a *AWAutoFocus) packResponse() string {
	return "d1" + a.AutoFocus.toWire()
}

// AWAutoFocusQuery requests information about the current autofocus status.
type AWAutoFocusQuery struct{}

func init() { registerRequest(func() AWRequest { return &AWAutoFocusQuery{} }) }
func (a *AWAutoFocusQuery) Acceptable() bool {
	return true
}
func (a *AWAutoFocusQuery) Response() AWResponse {
	return &AWAutoFocus{}
}
func (a *AWAutoFocusQuery) requestSignature() string {
	return "#D1"
}
func (a *AWAutoFocusQuery) unpackRequest(_ string) {}
func (a *AWAutoFocusQuery) packRequest() string {
	return "#D1"
}

// AWIrisTo commands the camera to set the iris to a specific value.
type AWIrisTo struct {
	Iris ScaleUnit
}

func init() { registerRequest(func() AWRequest { return &AWIrisTo{} }) }
func init() { registerResponse(func() AWResponse { return &AWIrisTo{} }) }
func (a *AWIrisTo) Acceptable() bool {
	return a.Iris.Acceptable()
}
func (a *AWIrisTo) Response() AWResponse {
	return a
}

func (a *AWIrisTo) requestSignature() string {
	return "#AXI\x01\x01\x01"
}
func (a *AWIrisTo) unpackRequest(cmd string) {
	a.Iris = toScaleUnit(cmd[4:7])
}
func (a *AWIrisTo) packRequest() string {
	return "#AXI" + a.Iris.toWire()
}
func (a *AWIrisTo) responseSignature() string {
	return "axi\x01\x01\x01"
}
func (a *AWIrisTo) unpackResponse(cmd string) {
	a.Iris = toScaleUnit(cmd[3:6])
}
func (a *AWIrisTo) packResponse() string {
	return "axi" + a.Iris.toWire()
}

// AWIrisQuery requests information about the current iris position.
type AWIrisQuery struct{}

func init() { registerRequest(func() AWRequest { return &AWIrisQuery{} }) }
func (a *AWIrisQuery) Acceptable() bool {
	return true
}
func (a *AWIrisQuery) Response() AWResponse {
	return &AWIrisTo{}
}
func (a *AWIrisQuery) requestSignature() string {
	return "#AXI"
}
func (a *AWIrisQuery) unpackRequest(_ string) {}
func (a *AWIrisQuery) packRequest() string {
	return "#AXI"
}

// LimitedScaleUnit represents a scale unit on a specific range.
//
// Zero value is an unacceptable value.
// A value of 1 is the near-and (closed) 99 is the far-end (open)
// Values outside the range 1 to 99 cause ErrUnacceptable
type LimitedScaleUnit int

func (l LimitedScaleUnit) toWire() string {
	if !l.Acceptable() {
		return "00"
	}
	return int2dec(int(l), 2)
}
func toLimitedScaleUnit(s string) LimitedScaleUnit {
	return LimitedScaleUnit(dec2int(s[0:2]))
}
func (c LimitedScaleUnit) Acceptable() bool {
	if c < 1 || c > 99 {
		return false
	}
	return true
}

// AWIris commands the camera iris to a specific value, just like AWIrisTo.
//
// Although syntactically alike, this is *not* a continous movement. This
// command is important for compatibility, but new codes should prefer AWIrisTo.
type AWIris struct {
	Iris LimitedScaleUnit
}

func init() { registerRequest(func() AWRequest { return &AWIris{} }) }
func init() { registerResponse(func() AWResponse { return &AWIris{} }) }
func (a *AWIris) Acceptable() bool {
	return a.Iris.Acceptable()
}
func (a *AWIris) Response() AWResponse {
	return a
}

func (a *AWIris) requestSignature() string {
	return "#I\x02\x02"
}
func (a *AWIris) unpackRequest(cmd string) {
	a.Iris = toLimitedScaleUnit(cmd[2:4])
}
func (a *AWIris) packRequest() string {
	return "#I" + a.Iris.toWire()
}
func (a *AWIris) responseSignature() string {
	return "iC\x02\x02"
}
func (a *AWIris) unpackResponse(cmd string) {
	a.Iris = toLimitedScaleUnit(cmd[2:4])
}
func (a *AWIris) packResponse() string {
	return "iC" + a.Iris.toWire()
}

// AWAutoIris configures the camera automatic iris control
type AWAutoIris struct {
	AutoIris Toggle
}

func init() { registerRequest(func() AWRequest { return &AWAutoIris{} }) }
func init() { registerResponse(func() AWResponse { return &AWAutoIris{} }) }
func (a *AWAutoIris) Acceptable() bool {
	return a.AutoIris.Acceptable()
}
func (a *AWAutoIris) Response() AWResponse {
	return a
}

func (a *AWAutoIris) requestSignature() string {
	return "#D3\x02"
}
func (a *AWAutoIris) unpackRequest(cmd string) {
	a.AutoIris = toToggle(cmd[3:4])
}
func (a *AWAutoIris) packRequest() string {
	return "#D3" + a.AutoIris.toWire()
}
func (a *AWAutoIris) responseSignature() string {
	return "d3\x02"
}
func (a *AWAutoIris) unpackResponse(cmd string) {
	a.AutoIris = toToggle(cmd[2:3])
}
func (a *AWAutoIris) packResponse() string {
	return "d3" + a.AutoIris.toWire()
}

// AWAutoIrisQuery requests the current automatic iris control configuration.
type AWAutoIrisQuery struct{}

func init() { registerRequest(func() AWRequest { return &AWAutoIrisQuery{} }) }
func (a *AWAutoIrisQuery) Acceptable() bool {
	return true
}
func (a *AWAutoIrisQuery) Response() AWResponse {
	return &AWAutoIris{}
}
func (a *AWAutoIrisQuery) requestSignature() string {
	return "#D3"
}
func (a *AWAutoIrisQuery) unpackRequest(_ string) {}
func (a *AWAutoIrisQuery) packRequest() string {
	return "#D3"
}

// AWCombinedIrisQuery requests the current iris position and configuration.
type AWCombinedIrisQuery struct{}

func init() { registerRequest(func() AWRequest { return &AWCombinedIrisQuery{} }) }
func (a *AWCombinedIrisQuery) Acceptable() bool {
	return true
}
func (a *AWCombinedIrisQuery) Response() AWResponse {
	return &AWCombinedIrisInfo{}
}
func (a *AWCombinedIrisQuery) requestSignature() string {
	return "#GI"
}
func (a *AWCombinedIrisQuery) unpackRequest(_ string) {}
func (a *AWCombinedIrisQuery) packRequest() string {
	return "#GI"
}

// AWCombinedIrisInfo is a response to AWCombinedIrisQuery.
type AWCombinedIrisInfo struct {
	Iris     ScaleUnit
	AutoIris Toggle
}

func init() { registerResponse(func() AWResponse { return &AWCombinedIrisInfo{} }) }

func (a *AWCombinedIrisInfo) responseSignature() string {
	// The --- as iris position is a "busy error". It will be left for Unknown
	// response in this case.
	return "gi\x01\x01\x01\x02"
}
func (a *AWCombinedIrisInfo) unpackResponse(cmd string) {
	_ = cmd[5]
	a.Iris = toScaleUnit(cmd[2:5])
	a.AutoIris = toToggle(cmd[5:6])
}
func (a *AWCombinedIrisInfo) packResponse() string {
	return "gi" + a.Iris.toWire() + a.AutoIris.toWire()
}

// Preset is a camera-stored preset number
//
// The camera has 100 presets, numbered 0-99. Values outside of that range will
// be capped to the nearest valid value (0 or 99).
type Preset int

func (p Preset) toWire() string {
	return int2dec(int(p), 2)
}
func toPreset(s string) Preset {
	return Preset(dec2int(s[0:2]))
}
func (p Preset) Acceptable() bool {
	// All wire-representable values are acceptable.
	return true
}

// AWPreset is a generic reply, indicating the affected preset number.
//
// Most requests affecting presets will return this response instead of
// themselves.
type AWPreset struct {
	Preset Preset
}

func init() { registerResponse(func() AWResponse { return &AWPreset{} }) }

func (a *AWPreset) responseSignature() string {
	return "s\x02\x02"
}
func (a *AWPreset) unpackResponse(cmd string) {
	a.Preset = toPreset(cmd[1:3])
}
func (a *AWPreset) packResponse() string {
	return "s" + a.Preset.toWire()
}

// AWPresetRegister saves the current camera position as a preset.
type AWPresetRegister struct {
	Preset Preset
}

func init() { registerRequest(func() AWRequest { return &AWPresetRegister{} }) }
func (a *AWPresetRegister) Acceptable() bool {
	return a.Preset.Acceptable()
}
func (a *AWPresetRegister) Response() AWResponse {
	return &AWPreset{Preset: a.Preset}
}
func (a *AWPresetRegister) requestSignature() string {
	return "#P\x02\x02"
}
func (a *AWPresetRegister) unpackRequest(cmd string) {
	a.Preset = toPreset(cmd[2:4])
}
func (a *AWPresetRegister) packRequest() string {
	return "#P" + a.Preset.toWire()
}

// AWPresetRecall commands the camera to obtain the specified preset position.
type AWPresetRecall struct {
	Preset Preset
}

func init() { registerRequest(func() AWRequest { return &AWPresetRecall{} }) }
func (a *AWPresetRecall) Acceptable() bool {
	return a.Preset.Acceptable()
}
func (a *AWPresetRecall) Response() AWResponse {
	return &AWPreset{Preset: a.Preset}
}
func (a *AWPresetRecall) requestSignature() string {
	return "#R\x02\x02"
}
func (a *AWPresetRecall) unpackRequest(cmd string) {
	a.Preset = toPreset(cmd[2:4])
}
func (a *AWPresetRecall) packRequest() string {
	return "#R" + a.Preset.toWire()
}

// AWPresetClear cleares the specified preset position data.
type AWPresetClear struct {
	Preset Preset
}

func init() { registerRequest(func() AWRequest { return &AWPresetClear{} }) }
func (a *AWPresetClear) Acceptable() bool {
	return a.Preset.Acceptable()
}
func (a *AWPresetClear) Response() AWResponse {
	return &AWPreset{Preset: a.Preset}
}
func (a *AWPresetClear) requestSignature() string {
	return "#C\x02\x02"
}
func (a *AWPresetClear) unpackRequest(cmd string) {
	a.Preset = toPreset(cmd[2:4])
}
func (a *AWPresetClear) packRequest() string {
	return "#C" + a.Preset.toWire()
}

// AWPresetQuery returns the last preset position the camera was commanded to.
type AWPresetQuery struct{}

func init() { registerRequest(func() AWRequest { return &AWPresetQuery{} }) }
func (a *AWPresetQuery) Acceptable() bool {
	return true
}
func (a *AWPresetQuery) Response() AWResponse {
	return &AWPreset{}
}
func (a *AWPresetQuery) requestSignature() string {
	return "#S"
}
func (a *AWPresetQuery) unpackRequest(_ string) {}
func (a *AWPresetQuery) packRequest() string {
	return "#S"
}

// HighSpeedUnit is a higher-resolution unit of motion speed.
//
// A Zero value is the factory-default value of approx 780.
// A HighSpeedUnit of 1 is equivalent to a SpeedUnit of 1.
// A HighSpeedUnit of 780 is equivalent to a SpeedUnit of 30.
//
// SpeedTable also applies to HighSpeedUnit, but it is provided separately.
// Mapping between HighSpeedUnit and SpeedUnit is a lossy conversion.
//
// TODO(zsh): Support timing-based HighSpeedUnit values for AW-UE150 cameras
type HighSpeedUnit int

func (s HighSpeedUnit) Acceptable() bool {
	// Altough not all values are valid, cameras never return AWErrUnacceptable
	return true
}
func (s HighSpeedUnit) toWire() string {
	var sp HighSpeedUnit
	// s == 0 is handled by the camera
	switch {
	case s == 0:
		// unlike SpeedUnit, zero HighSpeedUnit is handled by the camera
		sp = 0
	case s > 0 && s <= 750:
		// map 1-750 to 250-999
		sp = s + 249
	case s < 0 && s >= -999:
		// keep invalid values for proxying
		sp = -s
	default:
		// #UPVS never returns invalid in practice. Use default instead.
		sp = 0
	}
	return int2dec(int(sp), 3)
}
func toHighSpeedUnit(data string) HighSpeedUnit {
	u := HighSpeedUnit(dec2int(data[0:3]))
	switch {
	case u == 0:
		return HighSpeedUnit(0)
	case u >= 250 && u <= 999:
		// map 250-999 to 1-750
		return HighSpeedUnit(u - 249)
	default:
		// keep invalid values for proxying
		return HighSpeedUnit(-u)
	}
}

// AWPresetSpeed configures the speed at which presets are recalled.
type AWPresetSpeed struct {
	Speed HighSpeedUnit
}

func init() { registerRequest(func() AWRequest { return &AWPresetSpeed{} }) }
func init() { registerResponse(func() AWResponse { return &AWPresetSpeed{} }) }
func (a *AWPresetSpeed) Acceptable() bool {
	return a.Speed.Acceptable()
}
func (a *AWPresetSpeed) Response() AWResponse {
	return a
}

func (a *AWPresetSpeed) requestSignature() string {
	return "#UPVS\x02\x02\x02"
}
func (a *AWPresetSpeed) unpackRequest(cmd string) {
	a.Speed = toHighSpeedUnit(cmd[5:8])
}
func (a *AWPresetSpeed) packRequest() string {
	return "#UPVS" + a.Speed.toWire()
}
func (a *AWPresetSpeed) responseSignature() string {
	return "uPVS\x02\x02\x02"
}
func (a *AWPresetSpeed) unpackResponse(cmd string) {
	a.Speed = toHighSpeedUnit(cmd[4:7])
}
func (a *AWPresetSpeed) packResponse() string {
	return "uPVS" + a.Speed.toWire()
}

// AWPresetSpeedQuery requests the current AWPresetSpeed setting.
type AWPresetSpeedQuery struct{}

func init() { registerRequest(func() AWRequest { return &AWPresetSpeedQuery{} }) }
func (a *AWPresetSpeedQuery) Acceptable() bool {
	return true
}
func (a *AWPresetSpeedQuery) Response() AWResponse {
	return &AWPresetSpeed{}
}
func (a *AWPresetSpeedQuery) requestSignature() string {
	return "#UPVS"
}
func (a *AWPresetSpeedQuery) unpackRequest(_ string) {}
func (a *AWPresetSpeedQuery) packRequest() string {
	return "#UPVS"
}

// AWPresetFreeze configures camera image freeze during preset operations.
type AWPresetFreeze struct {
	Freeze Toggle
}

func init() { registerRequest(func() AWRequest { return &AWPresetFreeze{} }) }
func init() { registerResponse(func() AWResponse { return &AWPresetFreeze{} }) }
func (a *AWPresetFreeze) Acceptable() bool {
	return a.Freeze.Acceptable()
}
func (a *AWPresetFreeze) Response() AWResponse {
	return a
}

func (a *AWPresetFreeze) requestSignature() string {
	return "#PRF\x02"
}
func (a *AWPresetFreeze) unpackRequest(cmd string) {
	a.Freeze = toToggle(cmd[4:5])
}
func (a *AWPresetFreeze) packRequest() string {
	return "#PRF" + a.Freeze.toWire()
}
func (a *AWPresetFreeze) responseSignature() string {
	return "pRF\x02"
}
func (a *AWPresetFreeze) unpackResponse(cmd string) {
	a.Freeze = toToggle(cmd[3:4])
}
func (a *AWPresetFreeze) packResponse() string {
	return "pRF" + a.Freeze.toWire()
}

// AWPresetFreezeQuery requests the current AWPresetFreeze setting.
type AWPresetFreezeQuery struct{}

func init() { registerRequest(func() AWRequest { return &AWPresetFreezeQuery{} }) }
func (a *AWPresetFreezeQuery) Acceptable() bool {
	return true
}
func (a *AWPresetFreezeQuery) Response() AWResponse {
	return &AWPresetFreeze{}
}
func (a *AWPresetFreezeQuery) requestSignature() string {
	return "#PRF"
}
func (a *AWPresetFreezeQuery) unpackRequest(_ string) {}
func (a *AWPresetFreezeQuery) packRequest() string {
	return "#PRF"
}

type AWPresetSpeedTable struct {
	Table SpeedTable
}

func init() { registerRequest(func() AWRequest { return &AWPresetSpeedTable{} }) }
func init() { registerResponse(func() AWResponse { return &AWPresetSpeedTable{} }) }
func (a *AWPresetSpeedTable) Acceptable() bool {
	return a.Table.Acceptable()
}
func (a *AWPresetSpeedTable) Response() AWResponse {
	return a
}

func (a *AWPresetSpeedTable) requestSignature() string {
	return "#PST\x02"
}
func (a *AWPresetSpeedTable) unpackRequest(cmd string) {
	a.Table = toSpeedTable(cmd[4:5])
}
func (a *AWPresetSpeedTable) packRequest() string {
	return "#PST" + a.Table.toWire()
}
func (a *AWPresetSpeedTable) responseSignature() string {
	return "pST\x02"
}
func (a *AWPresetSpeedTable) unpackResponse(cmd string) {
	a.Table = toSpeedTable(cmd[3:4])
}
func (a *AWPresetSpeedTable) packResponse() string {
	return "pST" + a.Table.toWire()
}

type FuseOffset int

func (f FuseOffset) Acceptable() bool {
	return f >= 0 && f <= 2
}
func (f FuseOffset) toWire() string {
	return int2hex(int(f), 2)
}
func toFuseOffset(data string) FuseOffset {
	return FuseOffset(hex2int(data[0:2]))
}

// AWPresetEntries returns a bitmask of the stored presets.
//
// Due the the on-wire representation, a single command can not represent all
// 100 presets available. Presets are returned in groups of 40, defined by
// the offset parameter:
// offset(0) -> preset 0-39
// offset(1) -> preset 40-79
// offset(2) -> preset 80-119 (unused above 99)
// The Entries field is always a full bitmask, with non-representative bits
// set to 0 when received. Non-represnetative bits are ignored when sending.
type AWPresetEntries struct {
	Offset  FuseOffset
	Entries FuseSet
}

func init() { registerResponse(func() AWResponse { return &AWPresetEntries{} }) }

func (a *AWPresetEntries) responseSignature() string {
	return "pE\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01"
}
func (a *AWPresetEntries) unpackResponse(cmd string) {
	_ = cmd[13]
	a.Offset = toFuseOffset(cmd[2:4])
	f := FuseSet{}
	f[1] |= uint32(strings.IndexByte(hexAlphabet, cmd[4])&0xF) << 4
	f[1] |= uint32(strings.IndexByte(hexAlphabet, cmd[5])&0xF) << 0
	f[0] |= uint32(strings.IndexByte(hexAlphabet, cmd[6])&0xF) << 28
	f[0] |= uint32(strings.IndexByte(hexAlphabet, cmd[7])&0xF) << 24
	f[0] |= uint32(strings.IndexByte(hexAlphabet, cmd[8])&0xF) << 20
	f[0] |= uint32(strings.IndexByte(hexAlphabet, cmd[9])&0xF) << 16
	f[0] |= uint32(strings.IndexByte(hexAlphabet, cmd[10])&0xF) << 12
	f[0] |= uint32(strings.IndexByte(hexAlphabet, cmd[11])&0xF) << 8
	f[0] |= uint32(strings.IndexByte(hexAlphabet, cmd[12])&0xF) << 4
	f[0] |= uint32(strings.IndexByte(hexAlphabet, cmd[13])&0xF) << 0
	a.Entries = f.ShiftLeft(uint(40 * a.Offset))
}
func (a *AWPresetEntries) packResponse() string {
	o := a.Offset
	if o < 0 {
		// represent invalid offsets as invalid
		return "pEFF0000000000"
	}
	s := make([]byte, 14)
	copy(s, "pE")
	copy(s[2:], o.toWire())
	f := a.Entries.ShiftRight(uint(40 * a.Offset))
	s[4] = hexAlphabet[(f[1]>>4)&0xF]
	s[5] = hexAlphabet[(f[1]>>0)&0xF]
	s[6] = hexAlphabet[(f[0]>>28)&0xF]
	s[7] = hexAlphabet[(f[0]>>24)&0xF]
	s[8] = hexAlphabet[(f[0]>>20)&0xF]
	s[9] = hexAlphabet[(f[0]>>16)&0xF]
	s[10] = hexAlphabet[(f[0]>>12)&0xF]
	s[11] = hexAlphabet[(f[0]>>8)&0xF]
	s[12] = hexAlphabet[(f[0]>>4)&0xF]
	s[13] = hexAlphabet[(f[0]>>0)&0xF]
	return string(s)
}

// Mask returns a bitmask of the bits valid on the given preset offset.
//
// This function helps to determine the bits valid inside an AWPresetEntries
func (o FuseOffset) Mask() FuseSet {
	if o < 0 || o > 3 {
		return FuseSet{0x0, 0x0, 0x0, 0x0}
	}
	return FuseSet{0xFFFFFFFF, 0xFF, 0x0, 0x0}.ShiftLeft(uint(o * 40))
}

// Offset returns the Offset capable of representing this Preset.
//
// This function helps to determine the offset to use in AWPresetEntries
func (p Preset) Offset() FuseOffset {
	return FuseOffset(int(p) / 10)
}

type AWPresetEntriesQuery struct {
	Offset FuseOffset
}

func init() { registerRequest(func() AWRequest { return &AWPresetEntriesQuery{} }) }
func (a *AWPresetEntriesQuery) Acceptable() bool {
	return a.Offset.Acceptable()
}
func (a *AWPresetEntriesQuery) Response() AWResponse {
	return &AWPresetEntries{Offset: a.Offset}
}
func (a *AWPresetEntriesQuery) requestSignature() string {
	return "#PE\x01\x01"
}
func (a *AWPresetEntriesQuery) unpackRequest(cmd string) {
	a.Offset = toFuseOffset(cmd[3:5])
}
func (a *AWPresetEntriesQuery) packRequest() string {
	return "#PE" + a.Offset.toWire()
}

// AWPresetPlayback is a response indicating a preset position has just been
// reached by the camera.
//
// This indicates that a camera has reached a specific preset location that it
// was previously commanded to.
type AWPresetPlayback struct {
	Preset Preset
}

func init() { registerResponse(func() AWResponse { return &AWPresetPlayback{} }) }

func (a *AWPresetPlayback) responseSignature() string {
	return "q\x02\x02"
}
func (a *AWPresetPlayback) unpackResponse(cmd string) {
	a.Preset = toPreset(cmd[1:3])
}
func (a *AWPresetPlayback) packResponse() string {
	return "q" + a.Preset.toWire()
}

// AWTallyEnable is a request to turn on/off the tally light functionality.
//
// This is a global configuration to enable/disable the functionality. To light
// up the tally light, a tally GPI also have to be pulled or the AWTallySet
// command used.
type AWTallyEnable struct {
	TallyEnable Toggle
}

func init() { registerRequest(func() AWRequest { return &AWTallyEnable{} }) }
func init() { registerResponse(func() AWResponse { return &AWTallyEnable{} }) }
func (a *AWTallyEnable) Acceptable() bool {
	return a.TallyEnable.Acceptable()
}
func (a *AWTallyEnable) Response() AWResponse {
	return a
}

func (a *AWTallyEnable) requestSignature() string {
	return "#TAE\x02"
}
func (a *AWTallyEnable) unpackRequest(cmd string) {
	a.TallyEnable = toToggle(cmd[4:5])
}
func (a *AWTallyEnable) packRequest() string {
	return "#TAE" + a.TallyEnable.toWire()
}
func (a *AWTallyEnable) responseSignature() string {
	return "tAE\x02"
}
func (a *AWTallyEnable) unpackResponse(cmd string) {
	a.TallyEnable = toToggle(cmd[3:4])
}
func (a *AWTallyEnable) packResponse() string {
	return "tAE" + a.TallyEnable.toWire()
}

// AWTallyEnableQuery is a request to query the current tally configuration
type AWTallyEnableQuery struct{}

func init() { registerRequest(func() AWRequest { return &AWTallyEnableQuery{} }) }
func (a *AWTallyEnableQuery) Acceptable() bool {
	return true
}
func (a *AWTallyEnableQuery) Response() AWResponse {
	return &AWTallyEnable{}
}
func (a *AWTallyEnableQuery) requestSignature() string {
	return "#TAE"
}
func (a *AWTallyEnableQuery) unpackRequest(_ string) {}
func (a *AWTallyEnableQuery) packRequest() string {
	return "#TAE"
}

// AWTallySet is a request to turn on/off the tally light
type AWTallySet struct {
	TallyLight Toggle
}

func init() { registerRequest(func() AWRequest { return &AWTallySet{} }) }
func init() { registerResponse(func() AWResponse { return &AWTallySet{} }) }
func (a *AWTallySet) Acceptable() bool {
	return a.TallyLight.Acceptable()
}
func (a *AWTallySet) Response() AWResponse {
	return a
}

func (a *AWTallySet) requestSignature() string {
	return "#DA\x02"
}
func (a *AWTallySet) unpackRequest(cmd string) {
	a.TallyLight = toToggle(cmd[3:4])
}
func (a *AWTallySet) packRequest() string {
	return "#DA" + a.TallyLight.toWire()
}
func (a *AWTallySet) responseSignature() string {
	return "dA\x02"
}
func (a *AWTallySet) unpackResponse(cmd string) {
	a.TallyLight = toToggle(cmd[2:3])
}
func (a *AWTallySet) packResponse() string {
	return "dA" + a.TallyLight.toWire()
}

// AWTallyQuery is a request to query the current tally light status
type AWTallyQuery struct{}

func init() { registerRequest(func() AWRequest { return &AWTallyQuery{} }) }
func (a *AWTallyQuery) Acceptable() bool {
	return true
}
func (a *AWTallyQuery) Response() AWResponse {
	return &AWTallySet{}
}
func (a *AWTallyQuery) requestSignature() string {
	return "#DA"
}
func (a *AWTallyQuery) unpackRequest(_ string) {}
func (a *AWTallyQuery) packRequest() string {
	return "#DA"
}

// AWWirelessRemote controls the status of remote controller functionality
type AWWirelessRemote struct {
	RemoteEnable Toggle
}

func init() { registerRequest(func() AWRequest { return &AWWirelessRemote{} }) }
func init() { registerResponse(func() AWResponse { return &AWWirelessRemote{} }) }
func (a *AWWirelessRemote) Acceptable() bool {
	return a.RemoteEnable.Acceptable()
}
func (a *AWWirelessRemote) Response() AWResponse {
	return a
}

func (a *AWWirelessRemote) requestSignature() string {
	return "#WLC\x02"
}
func (a *AWWirelessRemote) unpackRequest(cmd string) {
	a.RemoteEnable = toToggle(cmd[4:5])
}
func (a *AWWirelessRemote) packRequest() string {
	return "#WLC" + a.RemoteEnable.toWire()
}
func (a *AWWirelessRemote) responseSignature() string {
	return "wLC\x02"
}
func (a *AWWirelessRemote) unpackResponse(cmd string) {
	a.RemoteEnable = toToggle(cmd[3:4])
}
func (a *AWWirelessRemote) packResponse() string {
	return "wLC" + a.RemoteEnable.toWire()
}

// AWWirelessRemoteQuery queries the current wireless remote status
type AWWirelessRemoteQuery struct{}

func init() { registerRequest(func() AWRequest { return &AWWirelessRemoteQuery{} }) }
func (a *AWWirelessRemoteQuery) Acceptable() bool {
	return true
}
func (a *AWWirelessRemoteQuery) Response() AWResponse {
	return &AWWirelessRemote{}
}
func (a *AWWirelessRemoteQuery) requestSignature() string {
	return "#WLC"
}
func (a *AWWirelessRemoteQuery) unpackRequest(_ string) {}
func (a *AWWirelessRemoteQuery) packRequest() string {
	return "#WLC"
}

type WirelessRemoteID int

const (
	RemoteCAM1 = iota
	RemoteCAM2
	RemoteCAM3
	RemoteCAM4
)

func (w WirelessRemoteID) toWire() string {
	if w < 0 {
		w = 9
	}
	return int2dec(int(w), 1)
}
func toWirelessRemoteID(s string) WirelessRemoteID {
	return WirelessRemoteID(dec2int(s[0:1]))
}
func (w WirelessRemoteID) Acceptable() bool {
	return w >= RemoteCAM1 && w <= RemoteCAM4
}

// AWWirelessRemoteID configures the number of the camera on the remote
type AWWirelessRemoteID struct {
	RemoteID WirelessRemoteID
}

func init() { registerRequest(func() AWRequest { return &AWWirelessRemoteID{} }) }
func init() { registerResponse(func() AWResponse { return &AWWirelessRemoteID{} }) }
func (a *AWWirelessRemoteID) Acceptable() bool {
	return a.RemoteID.Acceptable()
}
func (a *AWWirelessRemoteID) Response() AWResponse {
	return a
}

func (a *AWWirelessRemoteID) requestSignature() string {
	return "#RID\x02"
}
func (a *AWWirelessRemoteID) unpackRequest(cmd string) {
	a.RemoteID = toWirelessRemoteID(cmd[4:5])
}
func (a *AWWirelessRemoteID) packRequest() string {
	return "#RID" + a.RemoteID.toWire()
}
func (a *AWWirelessRemoteID) responseSignature() string {
	return "rID\x02"
}
func (a *AWWirelessRemoteID) unpackResponse(cmd string) {
	a.RemoteID = toWirelessRemoteID(cmd[3:4])
}
func (a *AWWirelessRemoteID) packResponse() string {
	return "rID" + a.RemoteID.toWire()
}

// AWWirelessRemoteIDQuery queries the current wireless remote ID
type AWWirelessRemoteIDQuery struct{}

func init() { registerRequest(func() AWRequest { return &AWWirelessRemoteIDQuery{} }) }
func (a *AWWirelessRemoteIDQuery) Acceptable() bool {
	return true
}
func (a *AWWirelessRemoteIDQuery) Response() AWResponse {
	return &AWWirelessRemoteID{}
}
func (a *AWWirelessRemoteIDQuery) requestSignature() string {
	return "#RID"
}
func (a *AWWirelessRemoteIDQuery) unpackRequest(_ string) {}
func (a *AWWirelessRemoteIDQuery) packRequest() string {
	return "#RID"
}

// AWSpeedWithZoom sets the Pan-Tilt speed slower when zoomed in.
type AWSpeedWithZoom struct {
	EnableSlowdown Toggle
}

func init() { registerRequest(func() AWRequest { return &AWSpeedWithZoom{} }) }
func init() { registerResponse(func() AWResponse { return &AWSpeedWithZoom{} }) }
func (a *AWSpeedWithZoom) Acceptable() bool {
	return a.EnableSlowdown.Acceptable()
}
func (a *AWSpeedWithZoom) Response() AWResponse {
	return a
}

func (a *AWSpeedWithZoom) requestSignature() string {
	return "#SWZ\x02"
}
func (a *AWSpeedWithZoom) unpackRequest(cmd string) {
	a.EnableSlowdown = toToggle(cmd[4:5])
}
func (a *AWSpeedWithZoom) packRequest() string {
	return "#SWZ" + a.EnableSlowdown.toWire()
}
func (a *AWSpeedWithZoom) responseSignature() string {
	return "sWZ\x02"
}
func (a *AWSpeedWithZoom) unpackResponse(cmd string) {
	a.EnableSlowdown = toToggle(cmd[3:4])
}
func (a *AWSpeedWithZoom) packResponse() string {
	return "sWZ" + a.EnableSlowdown.toWire()
}

// AWSpeedWithZoomQuery requests the status of speed with zoom slowdown
type AWSpeedWithZoomQuery struct{}

func init() { registerRequest(func() AWRequest { return &AWSpeedWithZoomQuery{} }) }
func (a *AWSpeedWithZoomQuery) Acceptable() bool {
	return true
}
func (a *AWSpeedWithZoomQuery) Response() AWResponse {
	return &AWSpeedWithZoom{}
}
func (a *AWSpeedWithZoomQuery) requestSignature() string {
	return "#SWZ"
}
func (a *AWSpeedWithZoomQuery) unpackRequest(_ string) {}
func (a *AWSpeedWithZoomQuery) packRequest() string {
	return "#SWZ"
}

type HealthCode int

const HealthOk HealthCode = 0x00

func (h HealthCode) toWire() string {
	if int(h) < 0 {
		h = 0xFF
	}
	return int2hex(int(h), 2)
}
func toHealthCode(s string) HealthCode {
	return HealthCode(hex2int(s[0:2]))
}

// Problem decides if the health code indicates an issue or if it's fine :-)
func (h HealthCode) Problem() bool {
	return h != HealthOk
}

// AWHealthStatus is information about the camera's phisical healthű
//
// The camera documentation refers to these as error codes. We intentionally
// don't use the term "error" because these are different from both go errors
// and error replies.
type AWHealthStatus struct {
	Code HealthCode
}

func init() { registerResponse(func() AWResponse { return &AWHealthStatus{} }) }

func (a *AWHealthStatus) responseSignature() string {
	return "rER\x01\x01"
}
func (a *AWHealthStatus) unpackResponse(cmd string) {
	a.Code = toHealthCode(cmd[3:5])
}
func (a *AWHealthStatus) packResponse() string {
	return "rER" + a.Code.toWire()
}

type AWHealthQuery struct{}

func init() { registerRequest(func() AWRequest { return &AWHealthQuery{} }) }
func (a *AWHealthQuery) Acceptable() bool {
	return true
}
func (a *AWHealthQuery) Response() AWResponse {
	return &AWHealthStatus{}
}
func (a *AWHealthQuery) requestSignature() string {
	return "#RER"
}
func (a *AWHealthQuery) unpackRequest(_ string) {}
func (a *AWHealthQuery) packRequest() string {
	return "#RER"
}

// AWOptionSwitch enables or disables the camera option. This is night-mode for
// all supported cameras.
type AWOptionSwitch struct {
	Enable Toggle
}

func init() { registerRequest(func() AWRequest { return &AWOptionSwitch{} }) }
func (a *AWOptionSwitch) Acceptable() bool {
	return a.Enable.Acceptable()
}
func (a *AWOptionSwitch) Response() AWResponse {
	return a
}

func (a *AWOptionSwitch) requestSignature() string {
	return "#D6\x02"
}
func (a *AWOptionSwitch) unpackRequest(cmd string) {
	a.Enable = toToggle(cmd[3:4])
}
func (a *AWOptionSwitch) packRequest() string {
	return "#D6" + a.Enable.toWire()
}
func (a *AWOptionSwitch) responseSignature() string {
	return "d6\x02"
}
func (a *AWOptionSwitch) unpackResponse(cmd string) {
	a.Enable = toToggle(cmd[2:3])
}
func (a *AWOptionSwitch) packResponse() string {
	return "d6" + a.Enable.toWire()
}

// AWOptionSwitchQuery requests the status of the camera option. This is
// night-mode for all supported cameras.
type AWOptionSwitchQuery struct{}

func init() { registerRequest(func() AWRequest { return &AWOptionSwitchQuery{} }) }
func (a *AWOptionSwitchQuery) Acceptable() bool {
	return true
}
func (a *AWOptionSwitchQuery) Response() AWResponse {
	return &AWOptionSwitch{}
}
func (a *AWOptionSwitchQuery) requestSignature() string {
	return "#D6"
}
func (a *AWOptionSwitchQuery) unpackRequest(_ string) {}
func (a *AWOptionSwitchQuery) packRequest() string {
	return "#D6"
}

// Toggle is a boolean on/off value which also have invalid values
type Toggle int

const (
	Off Toggle = 0
	On  Toggle = 1
)

func (t Toggle) toWire() string {
	if int(t) < 0 {
		// keep unrepresentable invalid values invalid when sent on-wire
		t = 9
	}
	return int2dec(int(t), 1)
}
func toToggle(s string) Toggle {
	return Toggle(dec2int(s[0:1]))
}
func (t Toggle) Acceptable() bool {
	return t == Off || t == On
}

// AWLensInformation is a response returning combined lens status information
//
// This information may be queried by AWLensInformationQuery, but also sent
// periodically as a notification if AWLensInformationNotify is enabled.
type AWLensInformation struct {
	Zoom  ScaleUnit
	Focus ScaleUnit
	Iris  ScaleUnit
}

func init() { registerResponse(func() AWResponse { return &AWLensInformation{} }) }

func (a *AWLensInformation) responseSignature() string {
	return "lPI\x01\x01\x01\x01\x01\x01\x01\x01\x01"
}
func (a *AWLensInformation) unpackResponse(cmd string) {
	_ = cmd[11]
	a.Zoom = toScaleUnit(cmd[3:6])
	a.Focus = toScaleUnit(cmd[6:9])
	a.Iris = toScaleUnit(cmd[9:12])
}
func (a *AWLensInformation) packResponse() string {
	return "lPI" + a.Zoom.toWire() + a.Focus.toWire() + a.Iris.toWire()
}

type AWLensInformationQuery struct{}

func init() { registerRequest(func() AWRequest { return &AWLensInformationQuery{} }) }
func (a *AWLensInformationQuery) Acceptable() bool {
	return true
}
func (a *AWLensInformationQuery) Response() AWResponse {
	return &AWLensInformation{}
}
func (a *AWLensInformationQuery) requestSignature() string {
	return "#LPI"
}
func (a *AWLensInformationQuery) unpackRequest(_ string) {}
func (a *AWLensInformationQuery) packRequest() string {
	return "#LPI"
}

// AWLensInformationNotify configures the automatic sending of AWLensInformation.
type AWLensInformationNotify struct {
	Enabled Toggle
}

func init() { registerRequest(func() AWRequest { return &AWLensInformationNotify{} }) }
func init() { registerResponse(func() AWResponse { return &AWLensInformationNotify{} }) }
func (a *AWLensInformationNotify) Acceptable() bool {
	return a.Enabled.Acceptable()
}
func (a *AWLensInformationNotify) Response() AWResponse {
	return a
}

func (a *AWLensInformationNotify) requestSignature() string {
	return "#LPC\x02"
}
func (a *AWLensInformationNotify) unpackRequest(cmd string) {
	a.Enabled = toToggle(cmd[4:5])
}
func (a *AWLensInformationNotify) packRequest() string {
	return "#LPC" + a.Enabled.toWire()
}
func (a *AWLensInformationNotify) responseSignature() string {
	return "lPC\x02"
}
func (a *AWLensInformationNotify) unpackResponse(cmd string) {
	a.Enabled = toToggle(cmd[3:4])
}
func (a *AWLensInformationNotify) packResponse() string {
	return "lPC" + a.Enabled.toWire()
}

// AWLensInformationNotifyQuery requests the current AWLensInformationNofity.
type AWLensInformationNotifyQuery struct{}

func init() { registerRequest(func() AWRequest { return &AWLensInformationNotifyQuery{} }) }
func (a *AWLensInformationNotifyQuery) Acceptable() bool {
	return true
}
func (a *AWLensInformationNotifyQuery) Response() AWResponse {
	return &AWLensInformationNotify{}
}
func (a *AWLensInformationNotifyQuery) requestSignature() string {
	return "#LPC"
}
func (a *AWLensInformationNotifyQuery) unpackRequest(_ string) {}
func (a *AWLensInformationNotifyQuery) packRequest() string {
	return "#LPC"
}

// AWSoftwareVersion indicates the software version running on the camera
//
// Software version information is rarely useful in itself, but it is sent as a
// heartbeat every 60 seconds by the camera when no other notifications took
// place.
type AWSoftwareVersion struct {
	Component int
	Major     int
	Minor     int
	Flag      byte
	Revision  int
	Mode      int
}

func init() { registerResponse(func() AWResponse { return &AWSoftwareVersion{} }) }

func (a *AWSoftwareVersion) responseSignature() string {
	return "qSV\x02V\x02\x02.\x02\x02\x00\x02\x02\x02"
}
func (a *AWSoftwareVersion) unpackResponse(cmd string) {
	_ = cmd[13]
	a.Component = dec2int(cmd[3:4])
	a.Major = dec2int(cmd[5:7])
	a.Minor = dec2int(cmd[8:10])
	a.Flag = cmd[10]
	a.Revision = dec2int(cmd[11:13])
	a.Mode = dec2int(cmd[13:14])
}
func (a *AWSoftwareVersion) packResponse() string {
	return "qSV" + int2dec(a.Component, 1) + "V" + int2dec(a.Major, 2) + "." +
		int2dec(a.Minor, 2) + string(a.Flag) + int2dec(a.Revision, 2) +
		int2dec(a.Mode, 1)
}

// AWSoftwareVersionQuery is a request to query the software version.
type AWSoftwareVersionQuery struct {
	Component int
}

func init() { registerRequest(func() AWRequest { return &AWSoftwareVersionQuery{} }) }
func (a *AWSoftwareVersionQuery) Acceptable() bool {
	return true
}
func (a *AWSoftwareVersionQuery) Response() AWResponse {
	return &AWSoftwareVersion{
		Component: a.Component,
	}
}
func (a *AWSoftwareVersionQuery) requestSignature() string {
	return "#QSV\x02"
}
func (a *AWSoftwareVersionQuery) unpackRequest(cmd string) {
	a.Component = dec2int(cmd[4:5])
}
func (a *AWSoftwareVersionQuery) packRequest() string {
	return "#QSV" + int2dec(a.Component, 1)
}

// AWAutoFocusAlternate enables or disables the camera autofocus mode
//
// This reques is equivalent to the AWAutoFocus but has a different on-wire
// representation. Most cameras send both commands to improve compatibility.
type AWAutoFocusAlternate struct {
	Enabled Toggle
}

func init() { registerRequest(func() AWRequest { return &AWAutoFocusAlternate{} }) }
func init() { registerResponse(func() AWResponse { return &AWAutoFocusAlternate{} }) }
func (a *AWAutoFocusAlternate) Acceptable() bool {
	return a.Enabled.Acceptable()
}
func (a *AWAutoFocusAlternate) Response() AWResponse {
	return a
}

func (a *AWAutoFocusAlternate) requestSignature() string {
	return "OAF:\x02"
}
func (a *AWAutoFocusAlternate) unpackRequest(cmd string) {
	a.Enabled = toToggle(cmd[4:5])
}
func (a *AWAutoFocusAlternate) packRequest() string {
	return "OAF:" + a.Enabled.toWire()
}
func (a *AWAutoFocusAlternate) responseSignature() string {
	return a.requestSignature()
}
func (a *AWAutoFocusAlternate) unpackResponse(cmd string) {
	a.unpackRequest(cmd)
}
func (a *AWAutoFocusAlternate) packResponse() string {
	return a.packRequest()
}

// AWAutoFocusQueryAlternate requests the current AWAutoFocusAlternate.
type AWAutoFocusQueryAlternate struct{}

func init() { registerRequest(func() AWRequest { return &AWAutoFocusQueryAlternate{} }) }
func (a *AWAutoFocusQueryAlternate) Acceptable() bool {
	return true
}
func (a *AWAutoFocusQueryAlternate) Response() AWResponse {
	return &AWAutoFocusAlternate{}
}
func (a *AWAutoFocusQueryAlternate) requestSignature() string {
	return "QAF"
}
func (a *AWAutoFocusQueryAlternate) unpackRequest(_ string) {}
func (a *AWAutoFocusQueryAlternate) packRequest() string {
	return "QAF"
}

// AWOneTouchFocus instruct the camrea to autofocus one time only
type AWOneTouchFocus struct {
	Parameter int // this parameter is meaningless, but it's there anyway
}

func init() { registerRequest(func() AWRequest { return &AWOneTouchFocus{} }) }
func (a *AWOneTouchFocus) Acceptable() bool {
	return a.Parameter == 0
}
func (a *AWOneTouchFocus) Response() AWResponse {
	return a
}
func (a *AWOneTouchFocus) requestSignature() string {
	return "OSE:69:\x02"
}
func (a *AWOneTouchFocus) unpackRequest(cmd string) {
	// The only valid value of this parameter is 1. Let's offset it so we can
	// leave it as zero-value and ignore it's existence.
	a.Parameter = dec2int(cmd[7:8]) - 1
}
func (a *AWOneTouchFocus) packRequest() string {
	return "OSE:69:" + int2dec(a.Parameter+1, 1)
}

func (a *AWOneTouchFocus) responseSignature() string {
	return a.requestSignature()
}
func (a *AWOneTouchFocus) unpackResponse(cmd string) {
	a.unpackRequest(cmd)
}
func (a *AWOneTouchFocus) packResponse() string {
	return a.packRequest()
}

type AWAutoIrisAlternate struct {
	Enabled Toggle
}

func init() { registerRequest(func() AWRequest { return &AWAutoIrisAlternate{} }) }
func init() { registerResponse(func() AWResponse { return &AWAutoIrisAlternate{} }) }
func (a *AWAutoIrisAlternate) Acceptable() bool {
	return a.Enabled.Acceptable()
}
func (a *AWAutoIrisAlternate) Response() AWResponse {
	return a
}

func (a *AWAutoIrisAlternate) requestSignature() string {
	return "ORS:\x02"
}
func (a *AWAutoIrisAlternate) unpackRequest(cmd string) {
	a.Enabled = toToggle(cmd[4:5])
}
func (a *AWAutoIrisAlternate) packRequest() string {
	return "ORS:" + a.Enabled.toWire()
}
func (a *AWAutoIrisAlternate) responseSignature() string {
	return a.requestSignature()
}
func (a *AWAutoIrisAlternate) unpackResponse(cmd string) {
	a.unpackRequest(cmd)
}
func (a *AWAutoIrisAlternate) packResponse() string {
	return a.packRequest()
}

// AWAutoIrisQueryAlternate requests the current AWAutoIrisAlternate.
type AWAutoIrisQueryAlternate struct{}

func init() { registerRequest(func() AWRequest { return &AWAutoIrisQueryAlternate{} }) }
func (a *AWAutoIrisQueryAlternate) Acceptable() bool {
	return true
}
func (a *AWAutoIrisQueryAlternate) Response() AWResponse {
	return &AWAutoIrisAlternate{}
}
func (a *AWAutoIrisQueryAlternate) requestSignature() string {
	return "QRS"
}
func (a *AWAutoIrisQueryAlternate) unpackRequest(_ string) {}
func (a *AWAutoIrisQueryAlternate) packRequest() string {
	return "QRS"
}

// AWIrisAlternate is functionally identical to AWIris and AWIrisTo but uses a
// different scale and on-wire representation. Minimum acceptable value is 0,
// maximum is 1023.
type AWIrisAlternate struct {
	Iris int
}

func init() { registerRequest(func() AWRequest { return &AWIrisAlternate{} }) }
func init() { registerResponse(func() AWResponse { return &AWIrisAlternate{} }) }
func (a *AWIrisAlternate) Acceptable() bool {
	return a.Iris >= 0 && a.Iris <= 1023
}
func (a *AWIrisAlternate) Response() AWResponse {
	return a
}

func (a *AWIrisAlternate) requestSignature() string {
	return "ORV:\x01\x01\x01"
}
func (a *AWIrisAlternate) unpackRequest(cmd string) {
	a.Iris = hex2int(cmd[4:7])
}
func (a *AWIrisAlternate) packRequest() string {
	return "ORV:" + int2hex(a.Iris, 3)
}
func (a *AWIrisAlternate) responseSignature() string {
	return a.requestSignature()
}
func (a *AWIrisAlternate) unpackResponse(cmd string) {
	a.unpackRequest(cmd)
}
func (a *AWIrisAlternate) packResponse() string {
	return a.packRequest()
}

// AWIrisQueryAlternate requests the current AWIrisAlternate.
type AWIrisQueryAlternate struct{}

func init() { registerRequest(func() AWRequest { return &AWIrisQueryAlternate{} }) }
func (a *AWIrisQueryAlternate) Acceptable() bool {
	return true
}
func (a *AWIrisQueryAlternate) Response() AWResponse {
	return &AWIrisAlternate{}
}
func (a *AWIrisQueryAlternate) requestSignature() string {
	return "QRV"
}
func (a *AWIrisQueryAlternate) unpackRequest(_ string) {}
func (a *AWIrisQueryAlternate) packRequest() string {
	return "QRV"
}

// AWIrisAlternate2 is functionally identical to AWIrisQueryAlternate,
// AWIrisTo and AWIris, but uses yet another scale. Valid values are between
// 0 and 255
type AWIrisAlternate2 struct {
	Iris int
}

func init() { registerRequest(func() AWRequest { return &AWIrisAlternate2{} }) }
func init() { registerResponse(func() AWResponse { return &AWIrisAlternate2{} }) }
func (a *AWIrisAlternate2) Acceptable() bool {
	return a.Iris >= 0 && a.Iris <= 255
}
func (a *AWIrisAlternate2) Response() AWResponse {
	return a
}

func (a *AWIrisAlternate2) requestSignature() string {
	return "OSD:4F:\x01\x01"
}
func (a *AWIrisAlternate2) unpackRequest(cmd string) {
	a.Iris = hex2int(cmd[7:9])
}
func (a *AWIrisAlternate2) packRequest() string {
	return "OSD:4F:" + int2hex(a.Iris, 2)
}
func (a *AWIrisAlternate2) responseSignature() string {
	return a.requestSignature()
}
func (a *AWIrisAlternate2) unpackResponse(cmd string) {
	a.unpackRequest(cmd)
}
func (a *AWIrisAlternate2) packResponse() string {
	return a.packRequest()
}

// AWIrisQueryAlternate2 requests the current AWIrisAlternate2.
type AWIrisQueryAlternate2 struct{}

func init() { registerRequest(func() AWRequest { return &AWIrisQueryAlternate2{} }) }
func (a *AWIrisQueryAlternate2) Acceptable() bool {
	return true
}
func (a *AWIrisQueryAlternate2) Response() AWResponse {
	return &AWIrisAlternate2{}
}
func (a *AWIrisQueryAlternate2) requestSignature() string {
	return "QSD:4F"
}
func (a *AWIrisQueryAlternate2) unpackRequest(_ string) {}
func (a *AWIrisQueryAlternate2) packRequest() string {
	return "QSD:4F"
}

type NDFilter int

const (
	NDFilterNone NDFilter = iota
	NDFilter1_4
	NDFilter1_16
	NDFilter1_64
	// NDFilterAuto is value 8 is documented, but does not work in practice.
)

func (n NDFilter) toWire() string {
	return int2dec(int(n), 1)
}
func toNDFilter(s string) NDFilter {
	return NDFilter(dec2int(s[0:1]))
}
func (n NDFilter) Acceptable() bool {
	return n >= NDFilterNone && n <= NDFilter1_64
}

// AWNDFilter configures the ND filter on the camera.
type AWNDFilter struct {
	Level NDFilter
}

func init() { registerRequest(func() AWRequest { return &AWNDFilter{} }) }
func init() { registerResponse(func() AWResponse { return &AWNDFilter{} }) }
func (a *AWNDFilter) Acceptable() bool {
	return a.Level.Acceptable()
}
func (a *AWNDFilter) Response() AWResponse {
	return a
}

func (a *AWNDFilter) requestSignature() string {
	return "OFT:\x02"
}
func (a *AWNDFilter) unpackRequest(cmd string) {
	a.Level = toNDFilter(cmd[4:5])
}
func (a *AWNDFilter) packRequest() string {
	return "OFT:" + a.Level.toWire()
}
func (a *AWNDFilter) responseSignature() string {
	return a.requestSignature()
}
func (a *AWNDFilter) unpackResponse(cmd string) {
	a.unpackRequest(cmd)
}
func (a *AWNDFilter) packResponse() string {
	return a.packRequest()
}

// AWNDFilterQuery requests the current ND filter level.
type AWNDFilterQuery struct{}

func init() { registerRequest(func() AWRequest { return &AWNDFilterQuery{} }) }
func (a *AWNDFilterQuery) Acceptable() bool {
	return true
}
func (a *AWNDFilterQuery) Response() AWResponse {
	return &AWNDFilter{}
}
func (a *AWNDFilterQuery) requestSignature() string {
	return "QFT"
}
func (a *AWNDFilterQuery) unpackRequest(_ string) {}
func (a *AWNDFilterQuery) packRequest() string {
	return "QFT"
}

// CenteredScale is an arbitrary scale with a middle default. Valid values are
// between -100 and +100.
type CenteredScale int

func (c CenteredScale) toWire() string {
	return int2hex(int(c)+0x32, 2)
}
func toCenteredScale(s string) CenteredScale {
	return CenteredScale(hex2int(s[0:2]) - 0x32)
}
func (c CenteredScale) Acceptable() bool {
	return c >= -32 && c <= 32
}

// AW contrast level is the contrast level configuration of the camera.
// In case of AW-UE70, -100 equals to -10, +100 equals to +10 visible in OSD
type AWContrastLevel struct {
	Level CenteredScale
}

func init() { registerRequest(func() AWRequest { return &AWContrastLevel{} }) }
func init() { registerResponse(func() AWResponse { return &AWContrastLevel{} }) }
func (a *AWContrastLevel) Acceptable() bool {
	return a.Level.Acceptable()
}
func (a *AWContrastLevel) Response() AWResponse {
	return a
}

func (a *AWContrastLevel) requestSignature() string {
	return "OSD:48:\x01\x01"
}
func (a *AWContrastLevel) unpackRequest(cmd string) {
	a.Level = toCenteredScale(cmd[7:9])
}
func (a *AWContrastLevel) packRequest() string {
	return "OSD:48:" + a.Level.toWire()
}
func (a *AWContrastLevel) responseSignature() string {
	return a.requestSignature()
}
func (a *AWContrastLevel) unpackResponse(cmd string) {
	a.unpackRequest(cmd)
}
func (a *AWContrastLevel) packResponse() string {
	return a.packRequest()
}

// AWContrastLevelQuery requests the current AWContrastLevel.
type AWContrastLevelQuery struct{}

func init() { registerRequest(func() AWRequest { return &AWContrastLevelQuery{} }) }
func (a *AWContrastLevelQuery) Acceptable() bool {
	return true
}
func (a *AWContrastLevelQuery) Response() AWResponse {
	return &AWContrastLevel{}
}
func (a *AWContrastLevelQuery) requestSignature() string {
	return "QSD:48"
}
func (a *AWContrastLevelQuery) unpackRequest(_ string) {}
func (a *AWContrastLevelQuery) packRequest() string {
	return "QSD:48"
}

// AWLensInformationAlternate is equal to AWLensInformation but has a different
// wire representation.
type AWLensInformationAlternate struct {
	Zoom  ScaleUnit
	Focus ScaleUnit
	Iris  ScaleUnit
}

func init() { registerResponse(func() AWResponse { return &AWLensInformationAlternate{} }) }
func (a *AWLensInformationAlternate) Acceptable() bool {
	return true
}
func (a *AWLensInformationAlternate) Response() AWResponse {
	return a
}

func (a *AWLensInformationAlternate) responseSignature() string {
	return "OSI:18:\x01\x01\x01:\x01\x01\x01:\x01\x01\x01"
}
func (a *AWLensInformationAlternate) unpackResponse(cmd string) {
	_ = cmd[17]
	a.Zoom = toScaleUnit(cmd[7:10])
	a.Focus = toScaleUnit(cmd[11:14])
	a.Iris = toScaleUnit(cmd[15:18])
}
func (a *AWLensInformationAlternate) packResponse() string {
	return "OSI:18:" + a.Zoom.toWire() + ":" + a.Focus.toWire() + ":" + a.Iris.toWire()
}

// AWLensInformationAlternateQuery requests the current AWLensInformationAlternate.
type AWLensInformationAlternateQuery struct{}

func init() { registerRequest(func() AWRequest { return &AWLensInformationAlternateQuery{} }) }
func (a *AWLensInformationAlternateQuery) Acceptable() bool {
	return true
}
func (a *AWLensInformationAlternateQuery) Response() AWResponse {
	return &AWLensInformationAlternate{}
}
func (a *AWLensInformationAlternateQuery) requestSignature() string {
	return "QSI:18"
}
func (a *AWLensInformationAlternateQuery) unpackRequest(_ string) {}
func (a *AWLensInformationAlternateQuery) packRequest() string {
	return "QSI:18"
}

// AWOSDMenu configures the on-screen visible camera menu
type AWOSDMenu struct {
	Display Toggle
}

func init() { registerResponse(func() AWResponse { return &AWOSDMenu{} }) }
func init() { registerRequest(func() AWRequest { return &AWOSDMenu{} }) }
func (a *AWOSDMenu) Acceptable() bool {
	// 2 is undocumented but accepted
	return a.Display > 0 && a.Display < 3
}
func (a *AWOSDMenu) Response() AWResponse {
	return a
}
func (a *AWOSDMenu) requestSignature() string {
	return "DUS:\x02"
}
func (a *AWOSDMenu) unpackRequest(cmd string) {
	a.Display = toToggle(cmd[4:5])
}
func (a *AWOSDMenu) packRequest() string {
	return "DUS:" + a.Display.toWire()
}

func (a *AWOSDMenu) responseSignature() string {
	return a.requestSignature()
}
func (a *AWOSDMenu) unpackResponse(cmd string) {
	a.unpackRequest(cmd)
}
func (a *AWOSDMenu) packResponse() string {
	return a.packRequest()
}

// AWOSDMenuQuery requests an AWOSDMenu status information.
type AWOSDMenuQuery struct{}

func init() { registerRequest(func() AWRequest { return &AWOSDMenuQuery{} }) }
func (a *AWOSDMenuQuery) Acceptable() bool {
	return true
}
func (a *AWOSDMenuQuery) Response() AWResponse {
	return &AWOSDMenu{}
}
func (a *AWOSDMenuQuery) requestSignature() string {
	return "QUS"
}
func (a *AWOSDMenuQuery) unpackRequest(_ string) {}
func (a *AWOSDMenuQuery) packRequest() string {
	return "QUS"
}

type AWTitle struct {
	Title string
}

func init() { registerResponse(func() AWResponse { return &AWTitle{} }) }
func init() { registerResponse(func() AWResponse { return &AWTitle{Title: " "} }) }

func (a *AWTitle) responseSignature() string {
	return "TITLE:\xF7"[:min(len(a.Title)+6, 7)]
}
func (a *AWTitle) unpackResponse(cmd string) {
	a.Title = cmd[6:]
}
func (a *AWTitle) packResponse() string {
	return "TITLE:" + a.Title
}
