package protocol

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
)

// EmbeddedServer provides a minimal NATS-compatible server for local use.
type EmbeddedServer struct {
	listener net.Listener
	mu       sync.Mutex
	closed   bool
	subs     map[string]*subscription
}

type subscription struct {
	connState *connState
	subject   string
	sid       string
}

type connState struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
	mu     sync.Mutex
	sids   map[string]struct{}
}

// StartEmbeddedServer starts a local NATS-compatible server.
func StartEmbeddedServer(ctx context.Context, addr string) (*EmbeddedServer, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("embedded nats: %w", err)
	}
	server := &EmbeddedServer{
		listener: listener,
		subs:     make(map[string]*subscription),
	}
	go server.acceptLoop(ctx)
	return server, nil
}

// URL returns the nats:// URL for the embedded server.
func (s *EmbeddedServer) URL() string {
	if s == nil || s.listener == nil {
		return ""
	}
	host, port, err := net.SplitHostPort(s.listener.Addr().String())
	if err != nil {
		return "nats://" + s.listener.Addr().String()
	}
	if host == "0.0.0.0" || host == "::" || host == "" {
		host = "127.0.0.1"
	}
	return "nats://" + net.JoinHostPort(host, port)
}

// Close stops the embedded server.
func (s *EmbeddedServer) Close() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	listener := s.listener
	s.mu.Unlock()
	if listener == nil {
		return nil
	}
	if err := listener.Close(); err != nil {
		return fmt.Errorf("embedded nats close: %w", err)
	}
	return nil
}

func (s *EmbeddedServer) acceptLoop(ctx context.Context) {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
			}
			return
		}
		state := &connState{
			conn:   conn,
			reader: bufio.NewReader(conn),
			writer: bufio.NewWriter(conn),
			sids:   make(map[string]struct{}),
		}
		go s.handleConn(ctx, state)
	}
}

func (s *EmbeddedServer) handleConn(ctx context.Context, state *connState) {
	if state == nil {
		return
	}
	_ = state.writeLine("INFO {}")
	for {
		line, err := state.reader.ReadString('\n')
		if err != nil {
			s.removeConn(state)
			_ = state.conn.Close()
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if line == "PING" {
			_ = state.writeLine("PONG")
			continue
		}
		if strings.HasPrefix(line, "CONNECT") || line == "PONG" || strings.HasPrefix(line, "-ERR") {
			continue
		}
		if strings.HasPrefix(line, "SUB ") {
			s.handleSub(state, line)
			continue
		}
		if strings.HasPrefix(line, "UNSUB ") {
			s.handleUnsub(state, line)
			continue
		}
		if strings.HasPrefix(line, "PUB ") {
			s.handlePub(state, line)
			continue
		}
		select {
		case <-ctx.Done():
			s.removeConn(state)
			_ = state.conn.Close()
			return
		default:
		}
	}
}

func (s *EmbeddedServer) handleSub(state *connState, line string) {
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return
	}
	subject := fields[1]
	sid := fields[2]
	s.mu.Lock()
	s.subs[sid] = &subscription{connState: state, subject: subject, sid: sid}
	s.mu.Unlock()
	state.mu.Lock()
	state.sids[sid] = struct{}{}
	state.mu.Unlock()
}

func (s *EmbeddedServer) handleUnsub(state *connState, line string) {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return
	}
	sid := fields[1]
	s.mu.Lock()
	delete(s.subs, sid)
	s.mu.Unlock()
	state.mu.Lock()
	delete(state.sids, sid)
	state.mu.Unlock()
}

func (s *EmbeddedServer) handlePub(state *connState, line string) {
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return
	}
	subject := fields[1]
	lengthRaw := fields[len(fields)-1]
	length, err := parseLength(lengthRaw)
	if err != nil {
		return
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(state.reader, payload); err != nil {
		return
	}
	trailer := make([]byte, 2)
	if _, err := io.ReadFull(state.reader, trailer); err != nil {
		return
	}
	s.publish(subject, payload)
}

func parseLength(raw string) (int, error) {
	var length int
	for _, ch := range raw {
		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("invalid length")
		}
		length = length*10 + int(ch-'0')
	}
	return length, nil
}

func (s *EmbeddedServer) publish(subject string, payload []byte) {
	s.mu.Lock()
	subs := make([]*subscription, 0, len(s.subs))
	for _, sub := range s.subs {
		if sub.subject == subject {
			subs = append(subs, sub)
		}
	}
	s.mu.Unlock()
	for _, sub := range subs {
		sub.connState.sendMessage(subject, sub.sid, payload)
	}
}

func (s *EmbeddedServer) removeConn(state *connState) {
	s.mu.Lock()
	for sid := range state.sids {
		delete(s.subs, sid)
	}
	s.mu.Unlock()
}

func (c *connState) writeLine(line string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := c.writer.WriteString(line + "\r\n"); err != nil {
		return err
	}
	if err := c.writer.Flush(); err != nil {
		return err
	}
	return nil
}

func (c *connState) sendMessage(subject, sid string, payload []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, _ = fmt.Fprintf(c.writer, "MSG %s %s %d\r\n", subject, sid, len(payload))
	_, _ = c.writer.Write(payload)
	_, _ = c.writer.WriteString("\r\n")
	_ = c.writer.Flush()
}
