package yamaha

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
)

// Message interface is implemented by all messages sendable to a socket.
type Message interface {
	_msg()
}

type HeartbeatMessage struct{}

func (m *HeartbeatMessage) _msg() {}

func parseLine(line []byte) (bool, Message, error) {
	l := trimSpace(line)

	var unsolicited bool
	switch {
	case bytes.HasPrefix(l, []byte("OK")):
		unsolicited = false
		l = l[2:]
	case bytes.HasPrefix(l, []byte("NOTIFY")):
		unsolicited = true
		l = l[6:]
	case bytes.HasPrefix(l, []byte("ERROR")):
		return false, nil, fmt.Errorf("broadcastkit/yamaha: protocol error: %s", string(l[5:]))
	default:
		return false, nil, fmt.Errorf("broadcastkit/yamaha: syntax: invalid prefix: %s", line)
	}

	if len(l) == 0 || !isSpace(l[0]) {
		return false, nil, fmt.Errorf("broadcastkit/yamaha: syntax: missing prefix separator: %s", line)
	}
	l = trimSpace(l)

	action, _ := cutSpace(l) // lookahead to decide param or info
	switch {
	case bytes.Equal(action, []byte("get")):
		fallthrough
	case bytes.Equal(action, []byte("set")):
		msg, err := parseParam(l)
		return unsolicited, msg, err
	default:
		msg, err := parseInfo(l)
		return unsolicited, msg, err
	}
}

// ScpSocket is a connection to a Yamaha mixer via Simple Control Protocol.
//
// Yamaha SCP can be communicated over any io.ReadWriter, but typically used
// over TCP via DialSCP
//
// ScpSocket must not be copied after first use.
// ScpSocket is safe to use from multiple goroutines.
// Conn must not be used directly after the first call to ScpSocket.
type ScpSocket struct {
	Conn io.ReadWriteCloser

	rlock sync.Mutex
	scan  *bufio.Scanner
}

func (c *ScpSocket) Write(msg Message) error {
	if c.Conn == nil {
		return errors.New("broadcastkit/yamaha: connection not established")
	}
	var buf bytes.Buffer
	switch msg := msg.(type) {
	case *HeartbeatMessage:
		fmt.Fprintf(&buf, "\n")
	case *StringParam:
		if msg.Set {
			fmt.Fprintf(&buf, "set %s %d %d %q\n", autoquote(msg.Address), msg.AddressX, msg.AddressY, msg.Value)
		} else {
			fmt.Fprintf(&buf, "get %s %d %d\n", autoquote(msg.Address), msg.AddressX, msg.AddressY)
		}
	case *IntParam:
		if msg.Set {
			fmt.Fprintf(&buf, "set %s %d %d %d\n", autoquote(msg.Address), msg.AddressX, msg.AddressY, msg.Value)
		} else {
			fmt.Fprintf(&buf, "get %s %d %d\n", autoquote(msg.Address), msg.AddressX, msg.AddressY)
		}
	case *InfoMessage:
		if msg.Value == "" {
			fmt.Fprintf(&buf, "%s %s\n", msg.Action, autoquote(msg.Address))
		} else {
			fmt.Fprintf(&buf, "%s %s %q\n", msg.Action, autoquote(msg.Address), msg.Value)
		}
	default:
		// This should be impossible due to the interface constraints.
		panic(fmt.Sprintf("broadcastkit/yamaha: invalid Message type: %T", msg))
	}
	_, err := c.Conn.Write(buf.Bytes())
	return err
}

// Read reads a single Message from the socket.
//
// bool indicates if the message is an unsolicited nofication (true) or a reply (false)
// Message is the message content received
// error indicates any errors, including error messages sent by the mixer
func (c *ScpSocket) Read() (bool, Message, error) {
	c.rlock.Lock()
	defer c.rlock.Unlock()
	if c.scan == nil {
		c.scan = bufio.NewScanner(c.Conn)
	}

	if !c.scan.Scan() {
		err := c.scan.Err()
		if err != nil {
			// non-EOF scanner errors are non-recoverable
			c.Conn.Close()
		} else {
			err = io.EOF
		}
		return false, nil, fmt.Errorf("broadcastkit/yamaha: scan: %w", err)
	}

	l := c.scan.Bytes()
	bytes.Trim(l, whitespaces)
	if len(l) == 0 {
		return false, &HeartbeatMessage{}, nil
	}

	return parseLine(l)
}

func (c *ScpSocket) Close() error {
	return c.Conn.Close()
}

// DialSCP connects to a Yamaha mixer via TCP and returns an ScpSocket.
func DialSCP(addr string) (*ScpSocket, error) {
	if !strings.Contains(addr, ":") {
		addr = addr + ":49280"
	}
	conn, err := net.Dial("tcp4", addr)
	if err != nil {
		return nil, err
	}
	return &ScpSocket{
		Conn: conn,
	}, nil
}
