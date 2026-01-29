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
)

var (
	// ErrNATSNotConnected is returned when the dispatcher is not connected.
	ErrNATSNotConnected = errors.New("nats not connected")
	// ErrNATSProtocol is returned for malformed NATS protocol frames.
	ErrNATSProtocol = errors.New("nats protocol error")
)

// NATSDispatcher publishes and subscribes to events over NATS.
type NATSDispatcher struct {
	conn    net.Conn
	reader  *bufio.Reader
	writer  *bufio.Writer
	mu      sync.Mutex
	subs    map[string]func(Event)
	closed  bool
	nextSID uint64
}

// NewNATSDispatcher connects to a NATS server and returns a dispatcher.
func NewNATSDispatcher(ctx context.Context, rawURL string) (*NATSDispatcher, error) {
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
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", host)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}
	dispatcher := &NATSDispatcher{
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
		subs:   make(map[string]func(Event)),
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
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed || d.conn == nil {
		return fmt.Errorf("nats publish: %w", ErrNATSNotConnected)
	}
	if _, err := fmt.Fprintf(d.writer, "PUB %s %d\r\n", subject, len(payload)); err != nil {
		return fmt.Errorf("nats publish: %w", err)
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
	d.subs[sid] = handler
	d.mu.Unlock()
	return &natsSubscription{dispatcher: d, sid: sid}, nil
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
	return nil
}

func (d *NATSDispatcher) sendConnect(ctx context.Context) error {
	if ctx.Err() != nil {
		return fmt.Errorf("nats connect: %w", ctx.Err())
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	connect := `{"verbose":false,"pedantic":false,"tls_required":false,"name":"amux","lang":"go","version":"1.0","protocol":1}`
	if _, err := d.writer.WriteString("CONNECT " + connect + "\r\n"); err != nil {
		return fmt.Errorf("nats connect: %w", err)
	}
	if err := d.writer.Flush(); err != nil {
		return fmt.Errorf("nats connect: %w", err)
	}
	return nil
}

func (d *NATSDispatcher) readLoop() {
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
			d.writePong()
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
			_ = subject
			payload := make([]byte, length)
			if _, err := io.ReadFull(d.reader, payload); err != nil {
				return
			}
			trailer := make([]byte, 2)
			if _, err := io.ReadFull(d.reader, trailer); err != nil {
				return
			}
			var event Event
			if err := json.Unmarshal(payload, &event); err != nil {
				continue
			}
			handler := d.handlerForSID(sid)
			if handler != nil {
				handler(event)
			}
		}
	}
}

func (d *NATSDispatcher) handlerForSID(sid string) func(Event) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.subs[sid]
}

func (d *NATSDispatcher) writePong() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed || d.conn == nil {
		return
	}
	_, _ = d.writer.WriteString("PONG\r\n")
	_ = d.writer.Flush()
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
