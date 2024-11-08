package panasonic

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// notifyUnpack retrieves a string response from the notification container
//
// This function ignores the undocumented metadata within the container.
// Format is:
// - 22 bytes reserved
// -  4 bytes of length, encoded as (l + 8) in uint16 big endian
// -  l bytes of string data (surrounded by CRLF)
// - 24 bytes reserved
//
// It is observed that string data sometimes have trailing null bytes. This
// function trims the string to printable ASCII characters.
func notifyUnpack(b []byte) (string, error) {
	if len(b) < 24 {
		return "", fmt.Errorf("panasonic.notifyUnpack: data too short %d (expected >=24)", len(b))
	}
	var l uint16
	_, err := binary.Decode(b[22:24], binary.BigEndian, &l)
	if err != nil {
		return "", fmt.Errorf("panasonic.tcpUnwrap: invalid length data: %w", err)
	}
	l -= 8
	dl := (22 + 2 + 4 + int(l) + 24)
	if len(b) != dl {
		return "", fmt.Errorf("panasonic.tcpUnwrap: data length mismatch %d (expected %d)", len(b), dl)
	}

	return string(trim(b[30 : 30+l])), nil
}

// notifyPack packages a string response into the notification container
//
// In addition to the documented format, this function guesses the undocumented
// bytes. Format is:
// -  4 bytes source IP
// -  2 bytes counter, which starts at 1, incremented for each packet
// -  6 bytes date, encoded as uint8 each: YEAR (last two decimal digits), MONTH, DAY, HOUR, MINUTE, SECOND
// - 10 bytes constant, hex 00 01 00 80 00 00 00 00 00 01
// -  2 bytes length, encoded as (l + 8) in uint16 big endian
// -  4 bytes constant, hex 01 00 00 00
// -  l bytes string data (surrounded by 2-2 bytes of CRLF)
// -  4 bytes constant, hex 00 02 00 18
// -  6 bytes source MAC
// -  2 bytes constant, hex 00 01
// -  6 bytes date, same as above
// -  6 bytes constant 00 00 00 00 00 00
// Date is in local time of the camera, not UTC.
func notifyPack(response string, session *NotifySession, date time.Time) []byte {
	offset := 30 + len(response)
	length := offset + 26

	b := make([]byte, length)
	_ = b[offset+25]
	_ = b[30]

	srcip := session.srcIP
	if srcip == (netip.Addr{}) {
		// avoid panicing on uninitialized sessions
		srcip = netip.IPv4Unspecified()
	}
	ipv4 := srcip.As4()
	copy(b, ipv4[:])

	counter := session.Counter.next()
	binary.BigEndian.PutUint16(b[4:], counter)

	b[6] = uint8(date.Year() % 100)
	b[7] = uint8(date.Month())
	b[8] = uint8(date.Day())
	b[9] = uint8(date.Hour())
	b[10] = uint8(date.Minute())
	b[11] = uint8(date.Second())

	copy(b[12:], "\x00\x01\x00\x80\x00\x00\x00\x00\x00\x01")

	replen := len(response) + 4 + 8 // reported length includes CRLFs + 8 bytes
	binary.BigEndian.PutUint16(b[22:], uint16(replen))

	copy(b[24:], "\x01\x00\x00\x00\r\n")
	copy(b[30:], response)
	copy(b[offset:], "\r\n\x00\x02\x00\x18")

	copy(b[offset+6:], session.srcMac)

	copy(b[offset+12:], "\x00\x01")
	copy(b[offset+14:], b[6:12]) // copy date instead of recalculating it
	copy(b[offset+20:], "\x00\x00\x00\x00\x00\x00")

	return b
}

// NotifyListener is the listener object for the AW notification protocol
// Use the CameraRemote.NotificationListener() method to obtain a new instance.
type NotifyListener struct {
	once sync.Once
	lis  *net.TCPListener
	cam  *CameraClient
}

// Start requests the camera to start sending notifications
func (l *NotifyListener) Start() error {
	err := l.start()
	l.once.Do(func() {}) // mark start as done
	return err
}

// start is the actual implementation of the Start() method
func (l *NotifyListener) start() error {
	port := netip.MustParseAddrPort(l.lis.Addr().String()).Port()
	res, err := l.cam.httpGet("/cgi-bin/event", "connect=start&my_port="+strconv.Itoa(int(port))+"&uid=0", nil)
	if err != nil {
		return &SystemError{err}
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		return &SystemError{fmt.Errorf("http status code: %d (expected %d)", res.StatusCode, http.StatusNoContent)}
	}
	return nil
}

// Stop requests the camera to stop sending notifications
func (l *NotifyListener) Stop() error {
	port := netip.MustParseAddrPort(l.lis.Addr().String()).Port()
	res, err := l.cam.httpGet("/cgi-bin/event", "connect=start&my_port="+strconv.Itoa(int(port))+"&uid=0", nil)
	if err != nil {
		return &SystemError{err}
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		return &SystemError{fmt.Errorf("http status code: %d (expected %d)", res.StatusCode, http.StatusNoContent)}
	}
	return nil
}

// Addr returns the local address where this listener awaits notifications
func (l *NotifyListener) Addr() netip.AddrPort {
	return netip.MustParseAddrPort(l.lis.Addr().String())
}

// acceptTCP accepts tcp connections from the camera only
func (l *NotifyListener) acceptTCP() (*net.TCPConn, error) {
	camaddr := l.cam.Remote.Addr()
	for {
		conn, err := l.lis.AcceptTCP()
		if err != nil {
			return nil, err
		}

		remaddr := netip.MustParseAddrPort(conn.RemoteAddr().String()).Addr()
		if remaddr == camaddr {
			return conn, nil
		}

		conn.Close()
	}
}

// SetDeadline sets the deadline for Accept to return by.
func (l *NotifyListener) SetDeadline(t time.Time) error {
	return l.lis.SetDeadline(t)
}

// Accept blocks until the next notification is received and returns it.
//
// If Start() has not been called, this function will call it automatically
// once. This does not include handling network issues or reconnections. It is
// the responsibility of the caller to re-call Start() if expected notifications
// are not received.
func (l *NotifyListener) Accept() (AWResponse, error) {
	l.once.Do(func() { l.start() })

	conn, err := l.acceptTCP()
	if err != nil {
		return nil, &SystemError{err}
	}

	conn.SetDeadline(time.Now().Add(networkTimeout))
	b, err := io.ReadAll(conn)
	conn.Close()

	if err != nil {
		return nil, &SystemError{err}
	}

	cmd, err := notifyUnpack(b)
	if err != nil {
		return nil, &SystemError{err}
	}

	return newResponse(cmd, quirkNotify), nil
}

// Close closes the listener.
//
// Any currently blocked Accept() calls will be unblocked and return an error.
// Stop() will be called automatically before closing.
func (l *NotifyListener) Close() error {
	_ = l.Stop()
	return l.lis.Close()
}

// NotifyServer is a thread-safe locked list of NotifySessions
type NotifyServer struct {
	lock sync.Mutex
	// It has been observed that Panasonic allows only one notification session
	// per IP address, stopping any previous ones when a new port is registered,
	// effectively keying the map on IP instead of IP:PORT. Uncertain if it is
	// intentional, we stick to IP:PORT mapping.
	list map[netip.AddrPort]*NotifySession
}

// SendAll sends a notification to all active sessions.
//
// Sessions which have observed 3 or more errors (equal to one dropped message)
// are automatically removed from the active sessions list.
func (l *NotifyServer) SendAll(res AWResponse) {
	l.lock.Lock()
	defer l.lock.Unlock()
	for dst, s := range l.list {
		if s.Errors.Load() > 2 {
			delete(l.list, dst)
			continue
		}
		s.Send(res)
	}
}

// Add adds a new peer to the session list
func (l *NotifyServer) Add(peer netip.AddrPort) {
	l.lock.Lock()
	defer l.lock.Unlock()
	if l.list == nil {
		l.list = make(map[netip.AddrPort]*NotifySession)
	}
	l.list[peer] = NewNotifySession(peer)
}

// Remove removes a peer from the session list
func (l *NotifyServer) Remove(peer netip.AddrPort) {
	l.lock.Lock()
	defer l.lock.Unlock()
	delete(l.list, peer)
}

// Len returns the number of active sessions
func (l *NotifyServer) Len() int {
	l.lock.Lock()
	defer l.lock.Unlock()
	for dst, s := range l.list {
		if s.Errors.Load() > 2 {
			delete(l.list, dst)
			continue
		}
	}
	return len(l.list)
}

// NotifySession is the metadata for the session
//
// Peer must not be modified after creation. Counter and Errors must be accessed
// atomically. NotifySession contains unexported autodiscovered metadata.
//
// Use NewNotifySession to create a NotifySession with autodiscovery.
type NotifySession struct {
	Peer netip.AddrPort
	// Figuring out the local source IP and MAC address is suprisingly complex.
	// We store them as session data to avoid re-discovering at every packet.
	srcIP   netip.Addr
	srcMac  net.HardwareAddr
	Counter NotifyCounter
	Errors  atomic.Int32
}

// NotifyCounter is a 16-bit unsigned thread-safe counter for keeping track of
// notification packet numbers.
//
// Implementation uses atomic.Uint32 as there's no atomic.Uint16. Applications
// should never access the underlying value directly.
type NotifyCounter atomic.Uint32

// next yields the counter value for the next notification packet
func (c *NotifyCounter) next() uint16 {
	return uint16((*atomic.Uint32)(c).Add(1))
}

// Load returns the counter value of the last notification packet
func (c *NotifyCounter) Load() uint16 {
	return uint16((*atomic.Uint32)(c).Load())
}

// discoverMac returns the MAC address which would be used on the local network
// with the given source IP. This may differ from the actual src MAC if the
// packet is going through a gateway.
func discoverMac(ip netip.Addr) net.HardwareAddr {
	interfaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	for _, i := range interfaces {
		addrs, err := i.Addrs()
		if err != nil {
			panic(err)
		}
		for _, a := range addrs {
			t := netip.MustParsePrefix(a.String())
			if t.Contains(ip) {
				return i.HardwareAddr
			}
		}
	}
	return nil
}

// discoverSource returns the IP address that would be autoselected for outgoing
// connections to the provided destination.
func discoverSource(dst netip.Addr) netip.Addr {
	// Dialing incorrectly is not the most elegant thing to do, but we can avoid
	// sending actual packets by using UDP instead of TCP for address discovery.
	c, err := net.DialUDP("udp4", nil, net.UDPAddrFromAddrPort(
		netip.AddrPortFrom(dst, 1),
	))
	if err != nil {
		panic(err)
	}
	defer c.Close()
	return netip.MustParseAddrPort(c.LocalAddr().String()).Addr()
}

// NewNotifySession creates a new NotifySession with autodiscovered metadata.
func NewNotifySession(peer netip.AddrPort) *NotifySession {
	srcIP := discoverSource(peer.Addr())
	srcMac := discoverMac(srcIP)
	return &NotifySession{
		Peer:   peer,
		srcIP:  srcIP,
		srcMac: srcMac,
	}
}

// sendTCP dumps the data into a new TCP connection, than slams it closed.
func sendTCP(b []byte, dst netip.AddrPort) error {
	conn, err := net.DialTimeout("tcp4", dst.String(), networkTimeout)
	if err != nil {
		return err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(networkTimeout))
	_, err = conn.Write(b)
	if err != nil {
		return err
	}
	return nil
}

// Send sends a notification packet to the session.
//
// It retries up to 3 times and returns the error of the last try.
func (s *NotifySession) Send(res AWResponse) error {
	if q, ok := res.(awQuirkedPacking); ok {
		res = q.packingQuirk(quirkNotify)
	}
	b := notifyPack(res.packResponse(), s, time.Now())
	var err error
	for i := 0; i < 3; i++ {
		err = sendTCP(b, s.Peer)
		if err == nil {
			s.Errors.Store(0)
			return nil
		}
		s.Errors.Add(1)
		// Sleeping just 100ms is the backoff strategy of Panasonic hardware.
		time.Sleep(100 * time.Millisecond)
	}
	return &SystemError{err}
}
