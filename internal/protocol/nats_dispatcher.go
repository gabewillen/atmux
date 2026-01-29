package protocol

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// ErrNATSNotConnected is returned when the dispatcher is not connected.
	ErrNATSNotConnected = errors.New("nats not connected")
	// ErrNATSProtocol is returned for malformed NATS protocol frames.
	ErrNATSProtocol = errors.New("nats protocol error")
)

// NATSOptions configures NATS connection metadata and auth.
type NATSOptions struct {
	// Name sets the client name.
	Name string
	// User sets the username for auth.
	User string
	// Password sets the password for auth.
	Password string
	// Token sets the auth token.
	Token string
}

type subscriptionHandler struct {
	onEvent func(Event)
	onRaw   func(Message)
}

// NATSDispatcher publishes and subscribes to events over NATS.
type NATSDispatcher struct {
	conn       net.Conn
	reader     *bufio.Reader
	writer     *bufio.Writer
	mu         sync.Mutex
	subs       map[string]subscriptionHandler
	closed     bool
	closedCh   chan struct{}
	nextSID    uint64
	nextInbox  uint64
	maxPayload int
	options    NATSOptions
}

// NewNATSDispatcher connects to a NATS server and returns a dispatcher.
func NewNATSDispatcher(ctx context.Context, rawURL string, options NATSOptions) (*NATSDispatcher, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}
	if parsed.Scheme == "" {
		parsed.Scheme = "nats"
	}
	if parsed.Scheme != "nats" {
		return nil, fmt.Errorf("nats connect: %w", ErrNATSProtocol)
	}
	host := parsed.Host
	if host == "" {
		return nil, fmt.Errorf("nats connect: %w", ErrNATSProtocol)
	}
	if !strings.Contains(host, ":") {
		host = host + ":4222"
	}
	if options.User == "" && options.Password == "" && parsed.User != nil {
		options.User = parsed.User.Username()
		if pass, ok := parsed.User.Password(); ok {
			options.Password = pass
		}
	}
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", host)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}
	dispatcher := &NATSDispatcher{
		conn:     conn,
		reader:   bufio.NewReader(conn),
		writer:   bufio.NewWriter(conn),
		subs:     make(map[string]subscriptionHandler),
		closedCh: make(chan struct{}),
		options:  options,
	}
	if err := dispatcher.readInfo(ctx); err != nil {
		_ = dispatcher.Close(ctx)
		return nil, err
	}
	if err := dispatcher.sendConnect(ctx); err != nil {
		_ = dispatcher.Close(ctx)
		return nil, err
	}
	go dispatcher.readLoop()
	return dispatcher, nil
}

// Close closes the underlying NATS connection.
func (d *NATSDispatcher) Close(ctx context.Context) error {
	_ = ctx
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return nil
	}
	d.closed = true
	select {
	case <-d.closedCh:
	default:
		close(d.closedCh)
	}
	if d.conn == nil {
		return nil
	}
	if err := d.conn.Close(); err != nil {
		return fmt.Errorf("nats close: %w", err)
	}
	return nil
}

// Publish publishes an event to a subject.
func (d *NATSDispatcher) Publish(ctx context.Context, subject string, event Event) error {
	if ctx.Err() != nil {
		return fmt.Errorf("nats publish: %w", ctx.Err())
	}
	if d == nil {
		return fmt.Errorf("nats publish: %w", ErrNATSNotConnected)
	}
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return fmt.Errorf("nats publish: %w", ErrNATSProtocol)
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("nats publish: %w", err)
	}
	return d.PublishRaw(ctx, subject, payload, "")
}

// Subscribe subscribes to a subject.
func (d *NATSDispatcher) Subscribe(ctx context.Context, subject string, handler func(Event)) (Subscription, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("nats subscribe: %w", ctx.Err())
	}
	if d == nil {
		return nil, fmt.Errorf("nats subscribe: %w", ErrNATSNotConnected)
	}
	subject = strings.TrimSpace(subject)
	if subject == "" || handler == nil {
		return nil, fmt.Errorf("nats subscribe: %w", ErrNATSProtocol)
	}
	return d.SubscribeRaw(ctx, subject, func(msg Message) {
		var event Event
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			return
		}
		handler(event)
	})
}

// PublishRaw publishes a raw payload to a subject with optional reply.
func (d *NATSDispatcher) PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error {
	if ctx.Err() != nil {
		return fmt.Errorf("nats publish: %w", ctx.Err())
	}
	if d == nil {
		return fmt.Errorf("nats publish: %w", ErrNATSNotConnected)
	}
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return fmt.Errorf("nats publish: %w", ErrNATSProtocol)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed || d.conn == nil {
		return fmt.Errorf("nats publish: %w", ErrNATSNotConnected)
	}
	reply = strings.TrimSpace(reply)
	if reply == "" {
		if _, err := fmt.Fprintf(d.writer, "PUB %s %d\r\n", subject, len(payload)); err != nil {
			return fmt.Errorf("nats publish: %w", err)
		}
	} else {
		if _, err := fmt.Fprintf(d.writer, "PUB %s %s %d\r\n", subject, reply, len(payload)); err != nil {
			return fmt.Errorf("nats publish: %w", err)
		}
	}
	if _, err := d.writer.Write(payload); err != nil {
		return fmt.Errorf("nats publish: %w", err)
	}
	if _, err := d.writer.WriteString("\r\n"); err != nil {
		return fmt.Errorf("nats publish: %w", err)
	}
	if err := d.writer.Flush(); err != nil {
		return fmt.Errorf("nats publish: %w", err)
	}
	return nil
}

// SubscribeRaw subscribes to a subject and receives raw NATS messages.
func (d *NATSDispatcher) SubscribeRaw(ctx context.Context, subject string, handler func(Message)) (Subscription, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("nats subscribe: %w", ctx.Err())
	}
	if d == nil {
		return nil, fmt.Errorf("nats subscribe: %w", ErrNATSNotConnected)
	}
	subject = strings.TrimSpace(subject)
	if subject == "" || handler == nil {
		return nil, fmt.Errorf("nats subscribe: %w", ErrNATSProtocol)
	}
	sid := strconv.FormatUint(atomic.AddUint64(&d.nextSID, 1), 10)
	d.mu.Lock()
	if d.closed || d.conn == nil {
		d.mu.Unlock()
		return nil, fmt.Errorf("nats subscribe: %w", ErrNATSNotConnected)
	}
	if _, err := fmt.Fprintf(d.writer, "SUB %s %s\r\n", subject, sid); err != nil {
		d.mu.Unlock()
		return nil, fmt.Errorf("nats subscribe: %w", err)
	}
	if err := d.writer.Flush(); err != nil {
		d.mu.Unlock()
		return nil, fmt.Errorf("nats subscribe: %w", err)
	}
	d.subs[sid] = subscriptionHandler{onRaw: handler}
	d.mu.Unlock()
	return &natsSubscription{dispatcher: d, sid: sid}, nil
}

// Request sends a request and waits for a single reply.
func (d *NATSDispatcher) Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (Message, error) {
	if ctx.Err() != nil {
		return Message{}, fmt.Errorf("nats request: %w", ctx.Err())
	}
	if d == nil {
		return Message{}, fmt.Errorf("nats request: %w", ErrNATSNotConnected)
	}
	inbox := d.nextInboxSubject()
	response := make(chan Message, 1)
	sub, err := d.SubscribeRaw(ctx, inbox, func(msg Message) {
		select {
		case response <- msg:
		default:
		}
	})
	if err != nil {
		return Message{}, fmt.Errorf("nats request: %w", err)
	}
	publishErr := d.PublishRaw(ctx, subject, payload, inbox)
	if publishErr != nil {
		_ = sub.Unsubscribe()
		return Message{}, fmt.Errorf("nats request: %w", publishErr)
	}
	wait := timeout
	if wait <= 0 {
		wait = 5 * time.Second
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case msg := <-response:
		_ = sub.Unsubscribe()
		return msg, nil
	case <-timer.C:
		_ = sub.Unsubscribe()
		return Message{}, fmt.Errorf("nats request: %w", context.DeadlineExceeded)
	case <-ctx.Done():
		_ = sub.Unsubscribe()
		return Message{}, fmt.Errorf("nats request: %w", ctx.Err())
	}
}

// MaxPayload returns the server-advertised maximum payload size.
func (d *NATSDispatcher) MaxPayload() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.maxPayload <= 0 {
		return 1024 * 1024
	}
	return d.maxPayload
}

// Closed returns a channel closed when the dispatcher connection ends.
func (d *NATSDispatcher) Closed() <-chan struct{} {
	if d == nil {
		ch := make(chan struct{})
		close(ch)
		return ch
	}
	return d.closedCh
}

func (d *NATSDispatcher) unsubscribe(sid string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed || d.conn == nil {
		return fmt.Errorf("nats unsubscribe: %w", ErrNATSNotConnected)
	}
	if _, err := fmt.Fprintf(d.writer, "UNSUB %s\r\n", sid); err != nil {
		return fmt.Errorf("nats unsubscribe: %w", err)
	}
	if err := d.writer.Flush(); err != nil {
		return fmt.Errorf("nats unsubscribe: %w", err)
	}
	delete(d.subs, sid)
	return nil
}

func (d *NATSDispatcher) readInfo(ctx context.Context) error {
	if ctx.Err() != nil {
		return fmt.Errorf("nats connect: %w", ctx.Err())
	}
	line, err := d.reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("nats connect: %w", err)
	}
	if !strings.HasPrefix(line, "INFO") {
		return fmt.Errorf("nats connect: %w", ErrNATSProtocol)
	}
	info := strings.TrimSpace(strings.TrimPrefix(line, "INFO"))
	var payload struct {
		MaxPayload int `json:"max_payload"`
	}
	if err := json.Unmarshal([]byte(info), &payload); err == nil {
		if payload.MaxPayload > 0 {
			d.maxPayload = payload.MaxPayload
		}
	}
	return nil
}

func (d *NATSDispatcher) sendConnect(ctx context.Context) error {
	if ctx.Err() != nil {
		return fmt.Errorf("nats connect: %w", ctx.Err())
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	connect := map[string]any{
		"verbose":      false,
		"pedantic":     false,
		"tls_required": false,
		"name":         "amux",
		"lang":         "go",
		"version":      "1.0",
		"protocol":     1,
	}
	if strings.TrimSpace(d.options.Name) != "" {
		connect["name"] = d.options.Name
	}
	if strings.TrimSpace(d.options.Token) != "" {
		connect["auth_token"] = d.options.Token
	}
	if strings.TrimSpace(d.options.User) != "" {
		connect["user"] = d.options.User
	}
	if strings.TrimSpace(d.options.Password) != "" {
		connect["pass"] = d.options.Password
	}
	encoded, err := json.Marshal(connect)
	if err != nil {
		return fmt.Errorf("nats connect: %w", err)
	}
	if _, err := d.writer.WriteString("CONNECT " + string(encoded) + "\r\n"); err != nil {
		return fmt.Errorf("nats connect: %w", err)
	}
	if err := d.writer.Flush(); err != nil {
		return fmt.Errorf("nats connect: %w", err)
	}
	return nil
}

func (d *NATSDispatcher) readLoop() {
	defer func() {
		d.mu.Lock()
		if !d.closed {
			d.closed = true
			select {
			case <-d.closedCh:
			default:
				close(d.closedCh)
			}
		}
		d.mu.Unlock()
	}()
	for {
		line, err := d.reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if line == "PING" {
			_ = d.writePong()
			continue
		}
		if line == "PONG" || strings.HasPrefix(line, "INFO ") || strings.HasPrefix(line, "-ERR") {
			continue
		}
		if strings.HasPrefix(line, "MSG ") {
			subject, sid, length, err := parseMsgLine(line)
			if err != nil {
				continue
			}
			reply := parseReplyFromHeader(line)
			payload := make([]byte, length)
			if _, err := io.ReadFull(d.reader, payload); err != nil {
				return
			}
			trailer := make([]byte, 2)
			if _, err := io.ReadFull(d.reader, trailer); err != nil {
				return
			}
			handler := d.handlerForSID(sid)
			if handler.onRaw == nil && handler.onEvent == nil {
				continue
			}
			msg := Message{Subject: subject, Reply: reply, Data: payload}
			if handler.onRaw != nil {
				handler.onRaw(msg)
				continue
			}
			var event Event
			if err := json.Unmarshal(payload, &event); err != nil {
				continue
			}
			handler.onEvent(event)
		}
	}
}

func (d *NATSDispatcher) handlerForSID(sid string) subscriptionHandler {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.subs[sid]
}

func (d *NATSDispatcher) writePong() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed || d.conn == nil {
		return nil
	}
	if _, err := d.writer.WriteString("PONG\r\n"); err != nil {
		return err
	}
	if err := d.writer.Flush(); err != nil {
		return err
	}
	return nil
}

func (d *NATSDispatcher) nextInboxSubject() string {
	seq := atomic.AddUint64(&d.nextInbox, 1)
	return fmt.Sprintf("_INBOX.amux.%d", seq)
}

func parseMsgLine(line string) (string, string, int, error) {
	parts := strings.Fields(line)
	if len(parts) < 4 || parts[0] != "MSG" {
		return "", "", 0, ErrNATSProtocol
	}
	subject := parts[1]
	sid := parts[2]
	lengthStr := parts[len(parts)-1]
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", "", 0, ErrNATSProtocol
	}
	return subject, sid, length, nil
}

func parseReplyFromHeader(line string) string {
	parts := strings.Fields(line)
	if len(parts) == 5 {
		return parts[3]
	}
	return ""
}

type natsSubscription struct {
	dispatcher *NATSDispatcher
	sid        string
}

func (n *natsSubscription) Unsubscribe() error {
	if n == nil || n.dispatcher == nil {
		return fmt.Errorf("nats unsubscribe: %w", ErrNATSNotConnected)
	}
	return n.dispatcher.unsubscribe(n.sid)
}
