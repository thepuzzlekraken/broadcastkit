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

type MessageCause int

const (
	ErrorCause  MessageCause = -1
	Reply       MessageCause = 0
	Unsolicited MessageCause = 1
)

func (c MessageCause) String() string {
	switch c {
	case ErrorCause:
		return "ERROR"
	case Reply:
		return "OK"
	case Unsolicited:
		return "NOTIFY"
	default:
		return "UNKNOWN"
	}
}

type Message interface {
	_send()
	_recv()
}

type ReceiveMessage interface {
	_recv()
}

type ErrorMessage struct {
	Details string
}

func (m *ErrorMessage) _recv() {}

type HeartbeatMessage struct{}

func (m *HeartbeatMessage) _recv() {}
func (m *HeartbeatMessage) _send() {}

func parseLine(line []byte) (MessageCause, ReceiveMessage, error) {
	l := trimSpace(line)

	var reason MessageCause
	switch {
	case bytes.HasPrefix(l, []byte("OK")):
		reason = Reply
		l = l[2:]
	case bytes.HasPrefix(l, []byte("NOTIFY")):
		reason = Unsolicited
		l = l[6:]
	case bytes.HasPrefix(l, []byte("ERROR")):
		reason = ErrorCause
		l = l[5:]
	default:
		return 0, nil, fmt.Errorf("syntax error: %s, unknown prefix", line)
	}

	if len(l) == 0 || !isSpace(l[0]) {
		return 0, nil, fmt.Errorf("syntax error: %s, no prefix separator", line)
	}
	l = trimSpace(l)

	if reason == ErrorCause {
		// Errors will not be processed further.
		return ErrorCause, &ErrorMessage{string(l)}, nil
	}

	switch {
	case bytes.HasPrefix(l, []byte("get")):
		fallthrough
	case bytes.HasPrefix(l, []byte("set")):
		msg, err := parseParam(l)
		return reason, msg, err
	default:
		msg, err := parseInfo(l)
		return reason, msg, err
	}
}

type ScpSocket struct {
	Conn net.Conn

	rlock sync.Mutex
	scan  *bufio.Scanner
}

func autoquote(s string) string {
	if strings.ContainsAny(s, " \"") {
		return fmt.Sprintf("%q", s)
	}
	return s
}

func (c *ScpSocket) Write(msg Message) error {
	if c.Conn == nil {
		return errors.New("connection not established")
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
		panic(fmt.Sprintf("invalid ScpSocket.Write message type: %T", msg))
	}
	_, err := c.Conn.Write(buf.Bytes())
	return err
}

func (c *ScpSocket) Read() (MessageCause, ReceiveMessage, error) {
	c.rlock.Lock()
	defer c.rlock.Unlock()
	if c.scan == nil {
		c.scan = bufio.NewScanner(c.Conn)
	}

	if !c.scan.Scan() {
		err := c.scan.Err()
		if err == nil {
			err = io.EOF
		}
		return 0, nil, fmt.Errorf("ScpSocket.Read failed: %w", err)
	}

	l := c.scan.Bytes()
	bytes.Trim(l, whitespaces)
	if len(l) == 0 {
		return Reply, &HeartbeatMessage{}, nil
	}

	return parseLine(l)
}

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
