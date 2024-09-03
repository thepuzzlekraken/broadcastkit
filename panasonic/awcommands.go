package panasonic

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
func (a *AWPower) requestSignature() (awHint, string) {
	return awPtz, "#O\x00"
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
	a.Power = PowerSwitch(cmd[2])
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
func (a *AWPower) Ok() bool {
	return true
}
func (a *AWPower) responseSignature() (awHint, string) {
	return awPtz | awNty, "p\x00"
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
func (a *AWPowerQuery) requestSignature() (awHint, string) {
	return awPtz, "#O"
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
func (a *AWInstall) requestSignature() (awHint, string) {
	return awPtz, "#INS\x02"
}
func (a *AWInstall) unpackRequest(cmd string) {
	a.Position = InstallSwitch(dec2int(cmd[4:5]))
}
func (a *AWInstall) packRequest() string {
	return "#INS" + int2dec(int(a.Position), 1)
}
func (a *AWInstall) Ok() bool {
	return true
}
func (a *AWInstall) responseSignature() (awHint, string) {
	return awPtz | awNty, "iNS\x02"
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
func (a *AWInstallQuery) requestSignature() (awHint, string) {
	return awPtz, "#INS"
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
	return int2dec(-int(m)+0x8000, 4)
}
func toMoveUnit(hex string) MoveUnit {
	// offset and inversion from Panasonic home 0x8000 => 0
	return MoveUnit(-dec2int(hex[0:4]) + 0x8000)
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
func (a *AWPanTiltTo) requestSignature() (awHint, string) {
	return awPtz, "#APC\x01\x01\x01\x01\x01\x01\x01\x01"
}
func (a *AWPanTiltTo) unpackRequest(cmd string) {
	_ = cmd[11]
	a.Pan = toMoveUnit(cmd[4:8])
	a.Tilt = toMoveUnit(cmd[8:12])
}
func (a *AWPanTiltTo) packRequest() string {
	return "#APC" + a.Pan.toWire() + a.Tilt.toWire()
}
func (a *AWPanTiltTo) Ok() bool {
	return true
}
func (a *AWPanTiltTo) responseSignature() (awHint, string) {
	// #APC not supported in awNty notifications unfortunately
	return awPtz, "aPC\x01\x01\x01\x01\x01\x01\x01\x01"
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
func (a *AWPanTiltQuery) requestSignature() (awHint, string) {
	return awPtz, "#APC"
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

const (
	// Values are offset by one to Panasonic definition, to allow for default
	DefaultSpeed SpeedTable = 0
	SlowSpeed    SpeedTable = 1
	MedSpeed     SpeedTable = 2
	FastSpeed    SpeedTable = 3
)

func (s SpeedUnit) Acceptable() bool {
	return s.Speed >= 0 && s.Speed <= 30 && s.Table >= 0 && s.Table <= 3
}
func (s SpeedUnit) toWire() string {
	sp := s.Speed
	if sp == 0 {
		sp = 9
	}
	tb := s.Table
	if tb == DefaultSpeed {
		tb = FastSpeed
	}
	return int2hex(sp, 2) + int2dec(int(tb)-1, 1)
}
func toSpeedUnit(data string) SpeedUnit {
	_ = data[3]
	return SpeedUnit{
		Speed: hex2int(data[0:2]),
		Table: SpeedTable(dec2int(data[2:3]) + 1),
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
func (a *AWPanTiltSpeedTo) requestSignature() (awHint, string) {
	return awPtz, "#APS\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01\x02"
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
func (a *AWPanTiltSpeedTo) Ok() bool {
	return true
}
func (a *AWPanTiltSpeedTo) responseSignature() (awHint, string) {
	// #APS not supported in awNty notifications unfortunately
	return awPtz, "aPS\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01\x02"
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
func (a *AWPanTiltBy) requestSignature() (awHint, string) {
	return awPtz, "#RPC\x01\x01\x01\x01\x01\x01\x01\x01"
}
func (a *AWPanTiltBy) unpackRequest(cmd string) {
	_ = cmd[11]
	a.Pan = toMoveUnit(cmd[4:8])
	a.Tilt = toMoveUnit(cmd[8:12])
}
func (a *AWPanTiltBy) packRequest() string {
	return "#RPC" + a.Pan.toWire() + a.Tilt.toWire()
}
func (a *AWPanTiltBy) Ok() bool {
	return true
}
func (a *AWPanTiltBy) responseSignature() (awHint, string) {
	// #RPC not supported in awNty notifications unfortunately
	return awPtz, "rPC\x01\x01\x01\x01\x01\x01\x01\x01"
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
func (a *AWPanTiltSpeedBy) requestSignature() (awHint, string) {
	return awPtz, "#RPS\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01\x02"
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
func (a *AWPanTiltSpeedBy) Ok() bool {
	return true
}
func (a *AWPanTiltSpeedBy) responseSignature() (awHint, string) {
	// #RPS not supported in awNty notifications
	return awPtz, "rPS\x01\x01\x01\x01\x01\x01\x01\x01\x01\x01\x02"
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
func (a *AWPan) requestSignature() (awHint, string) {
	return awPtz, "#P\x02\x02"
}
func (a *AWPan) unpackRequest(cmd string) {
	a.Pan = toInteractiveSpeed(cmd[2:4])
}
func (a *AWPan) packRequest() string {
	return "#P" + a.Pan.toWire()
}
func (a *AWPan) Ok() bool {
	return true
}
func (a *AWPan) responseSignature() (awHint, string) {
	return awPtz, "pS\x02\x02"
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
func (a *AWTilt) requestSignature() (awHint, string) {
	return awPtz, "#T\x02\x02"
}
func (a *AWTilt) unpackRequest(cmd string) {
	a.Tilt = toInteractiveSpeed(cmd[2:4])
}
func (a *AWTilt) packRequest() string {
	return "#T" + a.Tilt.toWire()
}
func (a *AWTilt) Ok() bool {
	return true
}
func (a *AWTilt) responseSignature() (awHint, string) {
	return awPtz, "tS\x02\x02"
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
func (a *AWPanTilt) requestSignature() (awHint, string) {
	return awPtz, "#PTS\x02\x02\x02\x02"
}
func (a *AWPanTilt) unpackRequest(cmd string) {
	_ = cmd[7]
	a.Pan = toInteractiveSpeed(cmd[4:6])
	a.Tilt = toInteractiveSpeed(cmd[6:8])
}
func (a *AWPanTilt) packRequest() string {
	return "#PTS" + a.Pan.toWire() + a.Tilt.toWire()
}
func (a *AWPanTilt) Ok() bool {
	return true
}
func (a *AWPanTilt) responseSignature() (awHint, string) {
	return awPtz, "pTS\x02\x02\x02\x02"
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
	// Panasonic API uses 0x555 as lowest value
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
func (a *AWZoomTo) Ok() bool {
	return true
}
func (a *AWZoomTo) Acceptable() bool {
	return a.Zoom.Acceptable()
}
func (a *AWZoomTo) Response() AWResponse {
	return a
}
func (a *AWZoomTo) requestSignature() (awHint, string) {
	return awPtz, "#AXZ\x01\x01\x01"
}
func (a *AWZoomTo) unpackRequest(cmd string) {
	a.Zoom = toScaleUnit(cmd[4:7])
}
func (a *AWZoomTo) packRequest() string {
	return "#AXZ" + a.Zoom.toWire()
}
func (a *AWZoomTo) responseSignature() (awHint, string) {
	return awPtz, "aXZ\x01\x01\x01"
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
func (a *AWZoomQuery) requestSignature() (awHint, string) {
	return awPtz, "#AXZ"
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
func (a *AWZoomResponseAlternate) Ok() bool {
	return true
}
func (a *AWZoomResponseAlternate) responseSignature() (awHint, string) {
	// There's a special case of gz--- which is returned instead of an eR2 error
	// when the camera is suspended. We'll just let that for UnknownResponse.
	return awPtz, "gz\x01\x01\x01"
}
func (a *AWZoomResponseAlternate) unpackResponse(cmd string) {
	a.Zoom = toScaleUnit(cmd[2:4])
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
func (a *AWZoomQueryAltenate) requestSignature() (awHint, string) {
	return awPtz, "#GZ"
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
func (a *AWZoom) requestSignature() (awHint, string) {
	return awPtz, "#Z\x02\x02"
}
func (a *AWZoom) unpackRequest(cmd string) {
	a.Zoom = toInteractiveSpeed(cmd[2:4])
}
func (a *AWZoom) packRequest() string {
	return "#Z" + a.Zoom.toWire()
}
func (a *AWZoom) Ok() bool {
	return true
}
func (a *AWZoom) responseSignature() (awHint, string) {
	return awPtz, "zS\x02\x02"
}
func (a *AWZoom) unpackResponse(cmd string) {
	a.Zoom = toInteractiveSpeed(cmd[2:4])
}
func (a *AWZoom) packResponse() string {
	return "zS" + a.Zoom.toWire()
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

// AWPresetPlayback is a special response that only appears in notifications
//
// This indicates that a camera has reached a specific preset location that it
// was previously commanded to.
type AWPresetPlayback struct {
	Preset Preset
}

func init() { registerResponse(func() AWResponse { return &AWPresetPlayback{} }) }
func (a *AWPresetPlayback) Ok() bool {
	return true
}
func (a *AWPresetPlayback) responseSignature() (awHint, string) {
	return awNty, "q\x02\x02"
}
func (a *AWPresetPlayback) unpackResponse(cmd string) {
	a.Preset = toPreset(cmd[1:3])
}
func (a *AWPresetPlayback) packResponse() string {
	return "q" + a.Preset.toWire()
}

// Toggle is a boolean on/off value which also have invalid values
type Toggle int

const (
	Off Toggle = 0
	On  Toggle = 1
)

func (t Toggle) toWire() string {
	if int(t) < 0 {
		// keep invalid values when represented on-wire
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
func (a *AWLensInformation) Ok() bool {
	return true
}
func (a *AWLensInformation) responseSignature() (awHint, string) {
	return awNty | awPtz, "lPI\x01\x01\x01\x01\x01\x01\x01\x01\x01"
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
func (a *AWLensInformationQuery) requestSignature() (awHint, string) {
	return awPtz, "#LPI"
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
func (a *AWLensInformationNotify) Ok() bool {
	return true
}
func (a *AWLensInformationNotify) requestSignature() (awHint, string) {
	return awPtz, "#LPC\x02"
}
func (a *AWLensInformationNotify) unpackRequest(cmd string) {
	a.Enabled = toToggle(cmd[4:5])
}
func (a *AWLensInformationNotify) packRequest() string {
	return "#LPC" + a.Enabled.toWire()
}
func (a *AWLensInformationNotify) responseSignature() (awHint, string) {
	return awNty, "lPC\x02"
}
func (a *AWLensInformationNotify) unpackResponse(cmd string) {
	a.Enabled = toToggle(cmd[3:4])
}
func (a *AWLensInformationNotify) packResponse() string {
	return "lPC" + a.Enabled.toWire()
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
func (a *AWSoftwareVersion) Ok() bool {
	return true
}
func (a *AWSoftwareVersion) responseSignature() (awHint, string) {
	return awNty | awPtz, "qSV\x02V\x02\x02.\x02\x02\x00\x02\x02\x02"
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
func (a *AWSoftwareVersionQuery) requestSignature() (awHint, string) {
	return awPtz, "#QSV\x02"
}
func (a *AWSoftwareVersionQuery) unpackRequest(cmd string) {
	a.Component = dec2int(cmd[4:5])
}
func (a *AWSoftwareVersionQuery) packRequest() string {
	return "#QSV" + int2dec(a.Component, 1)
}
