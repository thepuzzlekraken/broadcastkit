package blackmagicdesign

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
)

// VideohubSocket is a wrapper for the Blackmagic Videohub protocol.
// Use VideohubDial to create a new connection.
type VideohubSocket struct {
	Conn  io.ReadWriteCloser
	rlock sync.Mutex
	scan  *bufio.Scanner
}

// VideohubBlock is the basic unit of data exchanged with a Videohub device.
// To meaningfully process VideohubBlock, type-assert it to the specific block
// type, for example ProtocolPreambleBlock. Blocks that are not supported
// by this library will be represented as UnknownBlock. For the best
// compatibility, applications should ignore blocks they do not recognize,
// including the UnknownBlock.
// This interface intentionally can't be implemented by other packages.
type VideohubBlock interface {
	header() string
	parse([]byte) error
	dump(*bytes.Buffer)
}

// DialVideohub connects to a Videohub device at the given address.
// If port is not specified, the default 9990 is assumed.
// This call supports TCP over IPv4 only. This is intended until Blackmagic
// introduces IPv6 support in it's hardware.
func DialVideohub(Addr string) (*VideohubSocket, error) {
	if !strings.Contains(Addr, ":") {
		Addr = Addr + ":9990"
	}
	conn, err := net.Dial("tcp4", Addr)
	if err != nil {
		return nil, err
	}
	v := &VideohubSocket{
		Conn: conn,
	}
	return v, nil
}

// VideohubListener is a net.Listener like struct for Videohub devices.
type VideohubListener struct {
	listener net.Listener
}

// ListenVideohub returns a listener for Videohub clients like net.Listen
func ListenVideohub(Addr string) (*VideohubListener, error) {
	if Addr == "" {
		Addr = "0.0.0.0:9990"
	}
	if !strings.Contains(Addr, ":") {
		Addr = Addr + ":9990"
	}
	listener, err := net.Listen("tcp4", Addr)
	if err != nil {
		return nil, err
	}
	v := &VideohubListener{
		listener: listener,
	}
	return v, nil
}

// Accept waits for and returns the next Videohub connection to the listener.
func (l *VideohubListener) Accept() (*VideohubSocket, error) {
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}
	v := &VideohubSocket{
		Conn: conn,
	}
	return v, nil
}

// Close closes the VideohubListener, including the underlying net.Listener.
func (l *VideohubListener) Close() error {
	return l.listener.Close()
}

// Addr returns the listener's network address.
func (l *VideohubListener) Addr() net.Addr {
	return l.listener.Addr()
}

// Write writes a VideohubBlock to the connection.
// Write can be made to time-out by SetDeadline or SetWriteDeadline.
// See also: net.Conn.Write()
func (c *VideohubSocket) Write(m VideohubBlock) error {
	var buf bytes.Buffer
	buf.WriteString(m.header())
	buf.WriteByte('\n')
	m.dump(&buf)
	buf.WriteByte('\n')
	_, err := c.Conn.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("broadcastkit/blackmagicdesign: socket write: %w", err)
	}
	return nil
}

// Read reads a VideohubBlock from the connection.
//
// Read maintains an internal buffer, and will block until a complete block is
// available. For unrecognized blocks, an UnknownBlock{} is returned.
// ACK and NAK messages are treated as complete blocks and returned immediately.
// Concurrent calls to Read are safe, but the order of blocks is not guaranteed.
// Read can be made to time-out by SetDeadline or SetReadDeadline.
func (c *VideohubSocket) Read() (VideohubBlock, error) {
	c.rlock.Lock()
	defer c.rlock.Unlock()
	if c.scan == nil {
		c.scan = bufio.NewScanner(c.Conn)
		c.scan.Split(blankSplitter)
	}
	for {
		if !c.scan.Scan() {
			err := c.scan.Err()
			if err != nil {
				// non-EOF scanner errors are not recoverable
				c.Conn.Close()
			} else {
				err = io.EOF
			}
			return nil, fmt.Errorf("broadcastkit/blackmagicdesign: videohub scan: %w", err)
		}

		r := c.scan.Bytes()
		header, body, ok := bytes.Cut(r, []byte("\n"))
		if !ok {
			// Specification states clients should ignore blocks that they
			// do not recognize. We assume empty blocks fall within this case.
			continue
		}

		// Case-sensitivity is unspecified, but devices behave case-insensitive
		// according to our testing. This cleanup should improve compatibility.
		header = trim(header)
		uppercase(header)

		msg := newBlock(string(header))
		if msg == nil {
			// Ignoring unrecognised blocks according to specification.
			continue
		}

		err := msg.parse(body)
		if err != nil {
			return msg, fmt.Errorf("broadcastkit/blackmagicdesign: videohub parse: %w", err)
		}

		return msg, nil
	}
}

// Close closes the VideohubSocket, including the underlying net.Conn.
func (c *VideohubSocket) Close() error {
	return c.Conn.Close()
}

// newBlock returns the specific block struct for a given header
// this is a statically maintained list of supported blocks
// any new implementations of VideohubBlock must be added here to fully work
func newBlock(header string) VideohubBlock {
	switch header {
	case ProtocolPreambleBlock{}.header():
		return &ProtocolPreambleBlock{}
	case VideohubDeviceBlock{}.header():
		return &VideohubDeviceBlock{}
	case AckBlock{}.header():
		return &AckBlock{}
	case NakBlock{}.header():
		return &NakBlock{}
	case PingBlock{}.header():
		return &PingBlock{}
	case InputLabelsBlock{}.header():
		return &InputLabelsBlock{}
	case OutputLabelsBlock{}.header():
		return &OutputLabelsBlock{}
	case MonitoringOutputLabelsBlock{}.header():
		return &MonitoringOutputLabelsBlock{}
	case SerialPortLabelsBlock{}.header():
		return &SerialPortLabelsBlock{}
	case FrameLabelsBlock{}.header():
		return &FrameLabelsBlock{}
	case VideoOutputRoutingBlock{}.header():
		return &VideoOutputRoutingBlock{}
	case VideoMonitoringOutputRoutingBlock{}.header():
		return &VideoMonitoringOutputRoutingBlock{}
	case SerialPortRoutingBlock{}.header():
		return &SerialPortRoutingBlock{}
	case ProcessingUnitRoutingBlock{}.header():
		return &ProcessingUnitRoutingBlock{}
	case FrameBufferRoutingBlock{}.header():
		return &FrameBufferRoutingBlock{}
	case VideoOutputLocksBlock{}.header():
		return &VideoOutputLocksBlock{}
	case MonitoringOutputLocksBlock{}.header():
		return &MonitoringOutputLocksBlock{}
	case SerialPortLocksBlock{}.header():
		return &SerialPortLocksBlock{}
	case ProcessingUnitLocksBlock{}.header():
		return &ProcessingUnitLocksBlock{}
	case FrameBufferLocksBlock{}.header():
		return &FrameBufferLocksBlock{}
	case ConfigurationBlock{}.header():
		return &ConfigurationBlock{}
	case EndPreludeBlock{}.header():
		return &EndPreludeBlock{}
	default:
		return &UnknownBlock{}
	}
}

type VersionNumber struct {
	Major int
	Minor int
}

// ProtocolPreambleBlock is the first block sent by the Videohub device.
type ProtocolPreambleBlock struct {
	Empty   bool          // Indicates an empty block to request information
	Version VersionNumber // Version of protocol spoken by the Videohub device
}

func (ProtocolPreambleBlock) header() string {
	return "PROTOCOL PREAMBLE:"
}

func (k *ProtocolPreambleBlock) parse(b []byte) error {
	if len(trim(b)) == 0 {
		k.Empty = true
		return nil
	}
	for l, d := range colonLines(b) {

		lowercase(l)
		switch string(l) {
		case "version":
			// trying to avoid depending on regexes
			ma, mi, ok := bytes.Cut(d, []byte("."))
			if !ok {
				return fmt.Errorf("malformed version in preamble")
			}
			var err error
			k.Version.Major, err = strconv.Atoi(string(ma))
			if err != nil {
				return fmt.Errorf("malformed major version in preamble: %w", err)
			}
			k.Version.Minor, err = strconv.Atoi(string(mi))
			if err != nil {
				return fmt.Errorf("malformed minor version in preamble: %w", err)
			}
		}
	}
	return nil
}

func (k *ProtocolPreambleBlock) dump(b *bytes.Buffer) {
	if k.Empty {
		return
	}
	b.WriteString("Version: ")
	b.WriteString(strconv.Itoa(k.Version.Major))
	b.WriteByte('.')
	b.WriteString(strconv.Itoa(k.Version.Minor))
	b.WriteString("\n")
}

// DevicePresent is ENUM type for device presence indicator.
// This is always DevicePresentTrue on modern devices. For any client-sent
// commands, this should be left as the zero-value: DevicePresentUnknown.
type DevicePresent int

const (
	DevicePresentUnknown = iota
	DevicePresentTrue
	DevicePresentFalse
	DevicePresentNeedsUpdate
)

func toDevicePresent(s string) DevicePresent {
	switch s {
	case "true":
		return DevicePresentTrue
	case "false":
		return DevicePresentFalse
	case "needs_update":
		return DevicePresentNeedsUpdate
	default:
		return DevicePresentUnknown
	}
}

// VideohubDeviceBlock describes the hardware capabilities and basic identifiers
// of the Videohub device.
// Sent by all devices upon connection. Clients may send this block to set the
// device's friendly name.
type VideohubDeviceBlock struct {
	Empty                  bool
	DevicePresent          DevicePresent
	ModelName              string
	FriendlyName           string
	UniqueID               string
	VideoInputs            int
	VideoProcessingUnits   int
	VideoOutputs           int
	VideoMonitoringOutputs int
	SerialPorts            int
}

func (VideohubDeviceBlock) header() string {
	return "VIDEOHUB DEVICE:"
}

func (k *VideohubDeviceBlock) parse(b []byte) error {
	if len(trim(b)) == 0 {
		k.Empty = true
		return nil
	}
	for key, val := range colonLines(b) {
		lowercase(key)
		switch string(key) {
		case "device present":
			lowercase(val)
			k.DevicePresent = toDevicePresent(string(val))
		case "model name":
			k.ModelName = string(val)
		case "friendly name":
			k.FriendlyName = string(val)
		case "unique id":
			k.UniqueID = string(val)
		case "video inputs":
			num, err := strconv.Atoi(string(val))
			if err != nil {
				continue
			}
			k.VideoInputs = num
		case "video processing units":
			num, err := strconv.Atoi(string(val))
			if err != nil {
				continue
			}
			k.VideoProcessingUnits = num
		case "video outputs":
			num, err := strconv.Atoi(string(val))
			if err != nil {
				continue
			}
			k.VideoOutputs = num
		case "video monitoring outputs":
			num, err := strconv.Atoi(string(val))
			if err != nil {
				continue
			}
			k.VideoMonitoringOutputs = num
		case "serial ports":
			num, err := strconv.Atoi(string(val))
			if err != nil {
				continue
			}
			k.SerialPorts = num
		}
	}
	return nil
}

func (k *VideohubDeviceBlock) dump(b *bytes.Buffer) {
	if k.Empty {
		return
	}
	if k.DevicePresent == DevicePresentFalse {
		b.WriteString("Device present: false\n")
		return
	}
	if k.DevicePresent == DevicePresentNeedsUpdate {
		b.WriteString("Device present: needs_update\n")
		return
	}
	if k.DevicePresent == DevicePresentTrue {
		b.WriteString("Device present: true\n")
	}
	if k.ModelName != "" {
		b.WriteString("Model name: ")
		b.WriteString(strings.ReplaceAll(k.ModelName, "\n", ""))
		b.WriteByte('\n')
	}
	if k.FriendlyName != "" {
		b.WriteString("Friendly name: ")
		b.WriteString(strings.ReplaceAll(k.FriendlyName, "\n", ""))
		b.WriteByte('\n')
	}
	if k.UniqueID != "" {
		b.WriteString("Unique id: ")
		b.WriteString(strings.ReplaceAll(k.UniqueID, "\n", ""))
		b.WriteByte('\n')
	}
	if k.DevicePresent == DevicePresentTrue {
		b.WriteString("Video inputs: ")
		b.WriteString(strconv.Itoa(k.VideoInputs))
		b.WriteString("\nVideo processing units: ")
		b.WriteString(strconv.Itoa(k.VideoProcessingUnits))
		b.WriteString("\nVideo outputs: ")
		b.WriteString(strconv.Itoa(k.VideoOutputs))
		b.WriteString("\nVideo monitoring outputs: ")
		b.WriteString(strconv.Itoa(k.VideoMonitoringOutputs))
		b.WriteString("\nSerial ports: ")
		b.WriteString(strconv.Itoa(k.SerialPorts))
		b.WriteByte('\n')
	}
}

// ConfigurationBlock describes the extra configuration of the Videohub device,
// excluding video routing and labels.
// Note: This block is not documented by Blackmagic Design. We've observed that
// changes to TakeMode may not be honored immediately by the behavior of the
// device front panel.
type ConfigurationBlock struct {
	Empty    bool
	TakeMode bool
}

func (ConfigurationBlock) header() string {
	return "CONFIGURATION:"
}

func (k *ConfigurationBlock) parse(b []byte) error {
	if len(trim(b)) == 0 {
		k.Empty = true
		return nil
	}
	for key, val := range colonLines(b) {
		lowercase(key)
		switch string(key) {
		case "take mode":
			lowercase(val)
			switch string(val) {
			case "true":
				k.TakeMode = true
			case "false":
				k.TakeMode = false
			}
		}
	}
	return nil
}

func (k *ConfigurationBlock) dump(b *bytes.Buffer) {
	if k.Empty {
		return
	}
	if k.TakeMode {
		b.WriteString("Take mode: true\n")
	} else {
		b.WriteString("Take mode: false\n")
	}
}

// Single-line block sent as acceptance confirmation by Videohub devices.
type AckBlock struct{}

func (AckBlock) header() string {
	return "ACK"
}

func (*AckBlock) parse([]byte) error {
	return nil
}

func (*AckBlock) dump(*bytes.Buffer) {

}

// Single-line block sent as rejection confirmation by Videohub devices.
type NakBlock struct{}

func (NakBlock) header() string {
	return "NAK"
}

func (*NakBlock) parse([]byte) error {
	return nil
}

func (*NakBlock) dump(*bytes.Buffer) {

}

// An empty block that may be used to check the connection is still alive.
// Videohub devices should immediately respond with an AckBlock.
type PingBlock struct{}

func (PingBlock) header() string {
	return "PING:"
}

func (*PingBlock) parse([]byte) error {
	return nil
}

func (*PingBlock) dump(*bytes.Buffer) {
	//nop
}

// An empty block signaling that the Videohub has finished dumping it's initial
// state and any further blocks are change notices.
type EndPreludeBlock struct{}

func (EndPreludeBlock) header() string {
	return "END PRELUDE:"
}

func (*EndPreludeBlock) parse([]byte) error {
	return nil
}

func (*EndPreludeBlock) dump(*bytes.Buffer) {
	//nop
}

// This block is used to indicate that the protocol wrapper library did not
// recognize the block which was sent. Clients can usually safely ignore these.
// Videohub servers should usually respond with a NakBlock.
type UnknownBlock struct{}

func (UnknownBlock) header() string {
	return ""
}
func (*UnknownBlock) parse([]byte) error {
	return nil
}

func (*UnknownBlock) dump(*bytes.Buffer) {
	//nop
}

// Labels are used to describe the string labels of connectors
// A nil set of labels indicates a request to receive the current label list.
// A partial set of labels from a client indicates a change request.
// A partial set of labels from a device indicates a change notification.
// A full set of labels is sent by devices upon connection and upon request.
type Labels map[int]string

func (s *Labels) parse(b []byte) error {
	if len(trim(b)) == 0 {
		*s = nil
	}
	c := make(Labels)
	for n, l := range numberedLines(b) {
		c[n] = string(l)
	}
	*s = c
	return nil
}

func (s *Labels) dump(b *bytes.Buffer) {
	for n, l := range orderedIter(*s) {
		b.WriteString(strconv.Itoa(n))
		b.WriteByte(' ')
		b.WriteString(strings.ReplaceAll(l, "\n", ""))
		b.WriteByte('\n')
	}
}

// InputLabelsBlock contains the string labels of input connectors.
type InputLabelsBlock struct {
	Labels
}

func (InputLabelsBlock) header() string {
	return "INPUT LABELS:"
}

// OutputLabelsBlock contains the string labels of output connectors.
type OutputLabelsBlock struct {
	Labels
}

func (OutputLabelsBlock) header() string {
	return "OUTPUT LABELS:"
}

// MonitoringOutputLabelsBlock contains the string labels of monitoring outputs
// Not used by modern Videohub devices.
type MonitoringOutputLabelsBlock struct {
	Labels
}

func (MonitoringOutputLabelsBlock) header() string {
	return "MONITORING OUTPUT LABELS:"
}

// SerialPortLabelsBlock contains the string labels of serial ports.
// Not used by modern Videohub devices.
type SerialPortLabelsBlock struct {
	Labels
}

func (SerialPortLabelsBlock) header() string {
	return "SERIAL PORT LABELS:"
}

// FrameLabelsBlock contains the string labels of frame connectors.
// Not used by modern Videohub devices.
type FrameLabelsBlock struct {
	Labels
}

func (FrameLabelsBlock) header() string {
	return "FRAME LABELS:"
}

// Routing describes the which input is connected to the specified output.
// The map is keyed by outputs containing values of inputs (map[OUTPUT] = INPUT)
// A nil set of routing indicates a request to receive the current routing.
// A partial routing from a client indicates a change request.
// A partial routing from a device indicates a change notification.
// A full routing is sent by devices upon connection and upon request.
type Routing map[int]int

func (r *Routing) parse(b []byte) error {
	if len(trim(b)) == 0 {
		*r = nil
	}
	c := make(Routing)
	for n, l := range numberedLines(b) {
		t, err := strconv.Atoi(string(l))
		if err != nil {
			continue
		}
		c[n] = t
	}
	*r = c
	if bytes.Count(b, []byte("\n")) != len(c) {
		return fmt.Errorf("partially-valid routing block")
	}
	return nil
}

func (r *Routing) dump(b *bytes.Buffer) {
	for n, l := range orderedIter(*r) {
		b.WriteString(strconv.Itoa(n))
		b.WriteByte(' ')
		b.WriteString(strconv.Itoa(l))
		b.WriteByte('\n')
	}
}

// VideoOutputRoutingBlock contains the primary video routing setup.
type VideoOutputRoutingBlock struct {
	Routing
}

func (VideoOutputRoutingBlock) header() string {
	return "VIDEO OUTPUT ROUTING:"
}

// VideoMonitoringOutputRoutingBlock contains routing for monitoring outputs.
// Not used by modern Videohub devices.
type VideoMonitoringOutputRoutingBlock struct {
	Routing
}

func (VideoMonitoringOutputRoutingBlock) header() string {
	return "VIDEO MONITORING OUTPUT ROUTING:"
}

// SerialPortRoutingBlock contains routing for serial ports.
// Not used by modern Videohub devices.
type SerialPortRoutingBlock struct {
	Routing
}

func (SerialPortRoutingBlock) header() string {
	return "SERIAL PORT ROUTING:"
}

// ProcessingUnitRoutingBlock contains routing for processing units.
// Not used by modern Videohub devices.
type ProcessingUnitRoutingBlock struct {
	Routing
}

func (ProcessingUnitRoutingBlock) header() string {
	return "PROCESSING UNIT ROUTING:"
}

// FrameBufferRoutingBlock contains routing for frame buffers.
// Not used by modern Videohub devices.
type FrameBufferRoutingBlock struct {
	Routing
}

func (FrameBufferRoutingBlock) header() string {
	return "FRAME BUFFER ROUTING:"
}

// A Lock is the ENUM for lock states
// Values slightly depend on the direction of message.
type Lock int

const (
	LockUnknown  = -1   // Unknown value, should never appear in a block.
	LockUnlocked = iota // Unlocked, or request to unlock owned
	LockLocked          // Locked non-owned, or request to lock
	LockOwned           // Locked owned, or request to lock
	LockForced          // Request to unlock non-owned
)

func (l Lock) String() string {
	switch l {
	case LockUnlocked:
		return "U"
	case LockLocked:
		return "L"
	case LockOwned:
		return "O"
	case LockForced:
		return "F"
	default:
		return "?"
	}
}

func toLock(s string) Lock {
	switch s {
	case "U":
		return LockUnlocked
	case "L":
		return LockLocked
	case "O":
		return LockOwned
	case "F":
		return LockForced
	default:
		return LockUnknown
	}
}

// Locks describes the locking state of items
// A nil set of locks indicates a request to receive the current lock set.
// A partial set of locks from a client indicates a change request.
// A partial set of locks from a device indicates a change notification.
// A full set of locks is sent by devices upon connection and upon request.
type Locks map[int]Lock

func (s *Locks) parse(b []byte) error {
	if len(trim(b)) == 0 {
		*s = nil
	}
	c := make(Locks)
	for n, l := range numberedLines(b) {
		k := toLock(string(l))
		if k == LockUnknown {
			continue
		}
		c[n] = k
	}
	*s = c
	return nil
}

func (s *Locks) dump(b *bytes.Buffer) {
	for n, l := range orderedIter(*s) {
		if l == LockUnknown {
			continue
		}
		b.WriteString(strconv.Itoa(n))
		b.WriteByte(' ')
		b.WriteString(l.String())
		b.WriteByte('\n')
	}
}

// VideoOutputLocksBlock contains the locking state of video outputs.
type VideoOutputLocksBlock struct {
	Locks
}

func (VideoOutputLocksBlock) header() string {
	return "VIDEO OUTPUT LOCKS:"
}

// MonitoringOutputLocksBlock contains the locking state of monitoring outputs.
// Not used by modern Videohub devices.
type MonitoringOutputLocksBlock struct {
	Locks
}

func (MonitoringOutputLocksBlock) header() string {
	return "MONITORING OUTPUT LOCKS:"
}

// SerialPortLocksBlock contains the locking state of serial ports.
// Not used by modern Videohub devices.
type SerialPortLocksBlock struct {
	Locks
}

func (SerialPortLocksBlock) header() string {
	return "SERIAL PORT LOCKS:"
}

// ProcessingUnitLocksBlock contains the locking state of processing units.
// Not used by modern Videohub devices.
type ProcessingUnitLocksBlock struct {
	Locks
}

func (ProcessingUnitLocksBlock) header() string {
	return "PROCESSING UNIT LOCKS:"
}

// FrameBufferLocksBlock contains the locking state of frame buffers.
// Not used by modern Videohub devices.
type FrameBufferLocksBlock struct {
	Locks
}

func (FrameBufferLocksBlock) header() string {
	return "FRAME BUFFER LOCKS:"
}
