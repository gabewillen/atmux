package protocol

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
)

// EmbeddedServerConfig configures the embedded NATS-compatible server.
type EmbeddedServerConfig struct {
	// MaxPayload sets the advertised max payload size.
	MaxPayload int
	// Auth configures per-connection authorization.
	Auth AuthConfig
}

// Permissions defines publish and subscribe authorizations.
type Permissions struct {
	Publish   []string
	Subscribe []string
}

// AuthConfig maps auth credentials to subject permissions.
type AuthConfig struct {
	Tokens map[string]Permissions
	Users  map[string]UserAuth
}

// UserAuth defines a username/password and permissions pair.
type UserAuth struct {
	Password    string
	Permissions Permissions
}

// ConnectInfo captures CONNECT payload fields used for auth.
type ConnectInfo struct {
	Token    string
	User     string
	Password string
	Name     string
}

var errAuthRequired = errors.New("auth required")

func (a AuthConfig) Authorize(info ConnectInfo) (Permissions, error) {
	if len(a.Tokens) == 0 && len(a.Users) == 0 {
		return Permissions{}, nil
	}
	if info.Token != "" {
		perms, ok := a.Tokens[info.Token]
		if ok {
			return perms, nil
		}
	}
	if info.User != "" {
		entry, ok := a.Users[info.User]
		if ok && entry.Password == info.Password {
			return entry.Permissions, nil
		}
	}
	return Permissions{}, errAuthRequired
}

// EmbeddedServer provides a minimal NATS-compatible server for local use.
type EmbeddedServer struct {
	listener   net.Listener
	mu         sync.Mutex
	closed     bool
	subs       map[*connState]map[string]*subscription
	maxPayload int
	auth       AuthConfig
}

type subscription struct {
	subject string
	sid     string
}

type connState struct {
	conn       net.Conn
	reader     *bufio.Reader
	writer     *bufio.Writer
	mu         sync.Mutex
	perms      Permissions
	authorized bool
}

// StartEmbeddedServer starts a local NATS-compatible server.
func StartEmbeddedServer(ctx context.Context, addr string, cfg EmbeddedServerConfig) (*EmbeddedServer, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("embedded nats: %w", err)
	}
	maxPayload := cfg.MaxPayload
	if maxPayload <= 0 {
		maxPayload = 1024 * 1024
	}
	server := &EmbeddedServer{
		listener:   listener,
		subs:       make(map[*connState]map[string]*subscription),
		maxPayload: maxPayload,
		auth:       cfg.Auth,
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
		}
		go s.handleConn(ctx, state)
	}
}

func (s *EmbeddedServer) handleConn(ctx context.Context, state *connState) {
	if state == nil {
		return
	}
	authRequired := len(s.auth.Tokens) > 0 || len(s.auth.Users) > 0
	info := fmt.Sprintf("INFO {\"max_payload\":%d,\"auth_required\":%v}", s.maxPayload, authRequired)
	_ = state.writeLine(info)
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
		if strings.HasPrefix(line, "CONNECT") {
			s.handleConnect(state, line)
			if authRequired && !state.authorized {
				return
			}
			continue
		}
		if !state.authorized && authRequired {
			_ = state.writeLine("-ERR 'Authorization Violation'")
			_ = state.conn.Close()
			return
		}
		if line == "PONG" || strings.HasPrefix(line, "-ERR") {
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

func (s *EmbeddedServer) handleConnect(state *connState, line string) {
	payload := strings.TrimSpace(strings.TrimPrefix(line, "CONNECT"))
	var raw map[string]any
	if err := decodeConnect(payload, &raw); err != nil {
		_ = state.writeLine("-ERR 'Protocol Error'")
		_ = state.conn.Close()
		return
	}
	info := ConnectInfo{}
	if value, ok := raw["auth_token"].(string); ok {
		info.Token = value
	}
	if value, ok := raw["user"].(string); ok {
		info.User = value
	}
	if value, ok := raw["pass"].(string); ok {
		info.Password = value
	}
	if value, ok := raw["name"].(string); ok {
		info.Name = value
	}
	perms, err := s.auth.Authorize(info)
	if err != nil {
		_ = state.writeLine("-ERR 'Authorization Violation'")
		_ = state.conn.Close()
		return
	}
	state.mu.Lock()
	state.perms = perms
	state.authorized = true
	state.mu.Unlock()
}

func decodeConnect(payload string, dest *map[string]any) error {
	if payload == "" {
		return fmt.Errorf("connect payload empty")
	}
	if dest == nil {
		return fmt.Errorf("connect payload nil")
	}
	return jsonUnmarshalStrict([]byte(payload), dest)
}

func jsonUnmarshalStrict(data []byte, dest *map[string]any) error {
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.UseNumber()
	if err := dec.Decode(dest); err != nil {
		return err
	}
	return nil
}

func (s *EmbeddedServer) handleSub(state *connState, line string) {
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return
	}
	subject := fields[1]
	if !subjectAllowed(subject, state.perms.Subscribe) {
		_ = state.writeLine("-ERR 'Permissions Violation'")
		return
	}
	sid := fields[2]
	s.mu.Lock()
	if s.subs[state] == nil {
		s.subs[state] = make(map[string]*subscription)
	}
	s.subs[state][sid] = &subscription{subject: subject, sid: sid}
	s.mu.Unlock()
}

func (s *EmbeddedServer) handleUnsub(state *connState, line string) {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return
	}
	sid := fields[1]
	s.mu.Lock()
	if subs, ok := s.subs[state]; ok {
		delete(subs, sid)
		if len(subs) == 0 {
			delete(s.subs, state)
		}
	}
	s.mu.Unlock()
}

func (s *EmbeddedServer) handlePub(state *connState, line string) {
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return
	}
	subject := fields[1]
	if !subjectAllowed(subject, state.perms.Publish) {
		_ = state.writeLine("-ERR 'Permissions Violation'")
		return
	}
	reply := ""
	lengthRaw := fields[len(fields)-1]
	if len(fields) == 4 {
		reply = fields[2]
	}
	length, err := parseLength(lengthRaw)
	if err != nil {
		return
	}
	if length > s.maxPayload {
		_ = state.writeLine("-ERR 'Maximum Payload Exceeded'")
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
	s.publish(subject, reply, payload)
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

func subjectAllowed(subject string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}
	for _, pattern := range patterns {
		if matchSubject(subject, pattern) {
			return true
		}
	}
	return false
}

func matchSubject(subject string, pattern string) bool {
	subjectParts := strings.Split(subject, ".")
	patternParts := strings.Split(pattern, ".")
	for i := 0; i < len(patternParts); i++ {
		if i >= len(subjectParts) {
			return false
		}
		switch patternParts[i] {
		case ">":
			return true
		case "*":
			continue
		default:
			if patternParts[i] != subjectParts[i] {
				return false
			}
		}
	}
	return len(subjectParts) == len(patternParts)
}

func (s *EmbeddedServer) publish(subject, reply string, payload []byte) {
	type delivery struct {
		state *connState
		sub   *subscription
	}
	s.mu.Lock()
	deliveries := make([]delivery, 0, len(s.subs))
	for state, subs := range s.subs {
		for _, sub := range subs {
			if matchSubject(subject, sub.subject) {
				deliveries = append(deliveries, delivery{state: state, sub: sub})
			}
		}
	}
	s.mu.Unlock()
	for _, delivery := range deliveries {
		delivery.state.sendMessage(subject, delivery.sub.sid, reply, payload)
	}
}

func (s *EmbeddedServer) removeConn(state *connState) {
	s.mu.Lock()
	delete(s.subs, state)
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

func (c *connState) sendMessage(subject, sid, reply string, payload []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if reply == "" {
		_, _ = fmt.Fprintf(c.writer, "MSG %s %s %d\r\n", subject, sid, len(payload))
	} else {
		_, _ = fmt.Fprintf(c.writer, "MSG %s %s %s %d\r\n", subject, sid, reply, len(payload))
	}
	_, _ = c.writer.Write(payload)
	_, _ = c.writer.WriteString("\r\n")
	_ = c.writer.Flush()
}
