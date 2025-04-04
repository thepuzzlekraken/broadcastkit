package panasonic

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SwitcherError struct {
	code int
}

func (e SwitcherError) Error() string {
	return "Switcher error: " + e.message()
}

func (e SwitcherError) message() string {
	switch e.code {
	case 1:
		return "1 Out of the parameter range"
	case 2:
		return "2 Syntax error"
	default:
		return strconv.Itoa(e.code)
	}
}

type SwitcherClient struct {
	Remote    netip.AddrPort
	KeepAlive bool
	lock      sync.Mutex
	thread    sync.Once
	tcp       *net.TCPConn
	tcpBuff   *bufio.Reader
}

const hsPeriod = 15 * time.Second

func (s *SwitcherClient) keepAliveThread() {
	tick := time.NewTicker(hsPeriod)
	defer tick.Stop()
	for range tick.C {
		s.command("QBSC:01")
	}
}

func (s *SwitcherClient) dial() error {
	s.thread.Do(func() {
		go s.keepAliveThread()
	})
	dialer := net.Dialer{Timeout: networkTimeout}
	conn, err := dialer.Dial("tcp4", s.Remote.String())
	if err != nil {
		return err
	}
	tcp, ok := conn.(*net.TCPConn)
	if !ok {
		return errors.New("connection handle casting to tcp4 failed")
	}
	s.tcp = tcp
	s.tcpBuff = bufio.NewReader(tcp)
	return nil
}

func (s *SwitcherClient) close() {
	if s.tcp != nil {
		s.tcp.Close()
		s.tcp = nil
		s.tcpBuff = nil
	}
}

func (s *SwitcherClient) command(send string) (string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	var syserr error
	for retry := 0; retry < 3; retry++ {
		syserr = nil // ignore earlier errors

		if s.tcp == nil { // connect if necesary
			syserr = s.dial()
			if syserr != nil {
				continue
			}
		}

		s.tcp.SetDeadline(time.Now().Add(networkTimeout))
		buf := make([]byte, len(send)+2)
		buf[0] = '\x02' // STX
		copy(buf[1:], send)
		buf[len(buf)-1] = '\x03' // ETX
		_, syserr = s.tcp.Write(buf)
		if syserr != nil {
			s.close()
			continue
		}

		s.tcp.SetDeadline(time.Now().Add(networkTimeout))
		var recv string

		recv, syserr = s.tcpBuff.ReadString('\x03') // messages end in ETX

		// Panasonic sometimes closes errors with \x00 instead of \x03
		// we check for these "hidden" errors before checking the error
		// This workaround only works because the bad errors are in the same
		// tcp message as the ACK above
		if strings.HasPrefix(recv, "\x02EROR:") {
			c, err := strconv.Atoi(recv[5:])
			if err != nil {
				return "", SwitcherError{code: -1}
			}
			return "", SwitcherError{code: c}
		}

		if syserr != nil {
			s.close()
			continue
		}

		if recv[0] != '\x02' {
			return "", errors.New("corrupt message from switcher, missing STX")
		}

		return recv[1 : len(recv)-1], nil
	}
	return "", &SystemError{syserr}
}

func (s *SwitcherClient) SwitchBus(bus Bus, src Source) error {
	params := fmt.Sprintf("%02d:%02d", bus, src)
	res, err := s.command("SBUS:" + params)
	if err != nil {
		return err
	}
	if res != "ABUS:"+params {
		return fmt.Errorf("unexpected switcher response to SBUS: %s", res)
	}
	return nil
}

func (s *SwitcherClient) QueryBus(bus Bus) (Source, error) {
	param := fmt.Sprintf("%02d", bus)
	res, err := s.command("QBSC:" + param)
	if err != nil {
		return 0, err
	}
	if !strings.HasPrefix(res, "ABSC:"+param+":") {
		return 0, fmt.Errorf("unexpected switcher response to QBUS: %s", res)
	}
	src, err := strconv.Atoi(res[len(param)+6:])
	if err != nil {
		return 0, err
	}
	return Source(src), nil
}
