package metus

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"sync"
)

type MetusSocket struct {
	Conn io.ReadWriter
	lock sync.Mutex
	once sync.Once
	buf  *bufio.Reader
}

func Connect(address netip.AddrPort) (*MetusSocket, error) {
	if !address.IsValid() {
		return nil, errors.New("invalid address")
	}
	if address.Port() == 0 {
		address = netip.AddrPortFrom(address.Addr(), 32106)
	}
	conn, err := net.Dial("tcp", address.String())
	if err != nil {
		return nil, err
	}
	return &MetusSocket{
		Conn: conn,
	}, nil
}

func (m *MetusSocket) command(cmd []byte, emtpyEnd int) ([][]byte, error) {
	m.once.Do(func() {
		m.buf = bufio.NewReader(m.Conn)
	})
	if m.Conn == nil {
		return nil, errors.New("not connected")
	}
	m.Conn.Write(cmd)
	var firstLine []byte
	var err error
	for {
		firstLine, err = m.buf.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		firstLine = bytes.TrimSpace(firstLine)
		if len(firstLine) == 0 {
			continue
		}
		break
	}
	if !bytes.HasPrefix(firstLine, []byte("OK: ")) {
		return nil, fmt.Errorf("unexpected response: %v", firstLine)
	}
	var lines [][]byte
	var empties int
	lines = append(lines, firstLine[4:])
	for {
		line, err := m.buf.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			empties++
			if empties >= emtpyEnd {
				return lines, nil
			}
			continue
		}
		lines = append(lines, line)
		empties = 0
	}
}

func (m *MetusSocket) StartAll() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	_, err := m.command([]byte("Start\r\n"), 2)
	return err
}

func (m *MetusSocket) Start(name string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Start %q\r\n", name)
	_, err := m.command(buf.Bytes(), 1)
	return err
}

func (m *MetusSocket) StopAll() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	_, err := m.command([]byte("Stop\r\n"), 2)
	return err
}

func (m *MetusSocket) Stop(name string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Stop %q\r\n", name)
	_, err := m.command(buf.Bytes(), 1)
	return err
}

type Status int

/*
None: indicates that the encoder has been disposed.
Running: shows that recording has just been started.
Runned: shows that Ingest has been capturing the media source.
Stopping: indicates that recording has just been stopped.
Stopped: indicates that the current encoder stopped capturing.
Pausing: shows that recording has just been paused.
Paused: shows that the encoder paused capturing.
Preparing: indicates that the user has just defined either encoder or encoder and its profile.
Prepared: indicates that the user has defined either encoder or encoder and its profile
Splitting: indicates that the user has clicked <split> button.
Splitted: shows that recording has been splitted.
*/
const (
	StatusNone Status = iota
	StatusRunning
	StatusRunned
	StatusStopping
	StatusStopped
	StatusPausing
	StatusPaused
	StatusPreparing
	StatusPrepared
	StatusSplitting
	StatusSplitted
)

func toStatus(status []byte) (Status, error) {
	switch string(status) {
	case "None":
		return StatusNone, nil
	case "Running":
		return StatusRunning, nil
	case "Runned":
		return StatusRunned, nil
	case "Stopping":
		return StatusStopping, nil
	case "Stopped":
		return StatusStopped, nil
	case "Pausing":
		return StatusPausing, nil
	case "Paused":
		return StatusPaused, nil
	case "Preparing":
		return StatusPreparing, nil
	case "Prepared":
		return StatusPrepared, nil
	case "Splitting":
		return StatusSplitting, nil
	case "Splitted":
		return StatusSplitted, nil
	default:
		return -1, fmt.Errorf("unknown status: %s", status)
	}
}

func (m *MetusSocket) Status(name string) (Status, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "EncStatus %q\r\n", name)
	reply, err := m.command(buf.Bytes(), 1)
	if err != nil {
		return -1, err
	}
	if !bytes.HasPrefix(reply[0], []byte(name+":")) {
		return -1, fmt.Errorf("unexpected reply: %v", reply)
	}
	status := reply[0][len(name)+1:]
	return toStatus(status)
}

func (m *MetusSocket) StatusAll() (map[string]Status, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	reply, err := m.command([]byte("EncStatus\r\n"), 1)
	if err != nil {
		return nil, err
	}
	r := make(map[string]Status)
	for _, line := range reply {
		i := bytes.LastIndex(line, []byte(":"))
		if i == -1 {
			return nil, fmt.Errorf("unexpected reply: %s", reply)
		}
		name := string(line[:i])
		status, err := toStatus(line[i+1:])
		if err != nil {
			return nil, err
		}
		r[name] = status
	}
	return r, nil
}
