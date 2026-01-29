package protocol

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestNATSDispatcherPublishSubscribe(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() {
		if err := listener.Close(); err != nil {
			t.Errorf("close listener: %v", err)
		}
	})
	serverErr := make(chan error, 1)
	pubSeen := make(chan Event, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			serverErr <- err
			return
		}
		defer func() {
			if err := conn.Close(); err != nil {
				serverErr <- err
			}
		}()
		reader := bufio.NewReader(conn)
		writer := bufio.NewWriter(conn)
		if _, err := writer.WriteString("INFO {}\r\n"); err != nil {
			serverErr <- err
			return
		}
		if err := writer.Flush(); err != nil {
			serverErr <- err
			return
		}
		if _, err := reader.ReadString('\n'); err != nil {
			serverErr <- err
			return
		}
		subLine, err := reader.ReadString('\n')
		if err != nil {
			serverErr <- err
			return
		}
		fields := strings.Fields(subLine)
		if len(fields) < 3 {
			serverErr <- fmt.Errorf("bad sub line: %s", subLine)
			return
		}
		subject := fields[1]
		sid := fields[2]
		event := Event{Name: "server.push"}
		payload, err := json.Marshal(event)
		if err != nil {
			serverErr <- err
			return
		}
		if _, err := fmt.Fprintf(writer, "MSG %s %s %d\r\n", subject, sid, len(payload)); err != nil {
			serverErr <- err
			return
		}
		if _, err := writer.Write(payload); err != nil {
			serverErr <- err
			return
		}
		if _, err := writer.WriteString("\r\n"); err != nil {
			serverErr <- err
			return
		}
		if err := writer.Flush(); err != nil {
			serverErr <- err
			return
		}
		pubLine, err := reader.ReadString('\n')
		if err != nil {
			serverErr <- err
			return
		}
		pubFields := strings.Fields(pubLine)
		if len(pubFields) < 3 {
			serverErr <- fmt.Errorf("bad pub line: %s", pubLine)
			return
		}
		length, err := strconv.Atoi(pubFields[len(pubFields)-1])
		if err != nil {
			serverErr <- err
			return
		}
		payloadBuf := make([]byte, length)
		if _, err := io.ReadFull(reader, payloadBuf); err != nil {
			serverErr <- err
			return
		}
		trailer := make([]byte, 2)
		if _, err := io.ReadFull(reader, trailer); err != nil {
			serverErr <- err
			return
		}
		var published Event
		if err := json.Unmarshal(payloadBuf, &published); err != nil {
			serverErr <- err
			return
		}
		pubSeen <- published
		serverErr <- nil
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	dispatcher, err := NewNATSDispatcher(ctx, "nats://"+listener.Addr().String())
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	received := make(chan Event, 1)
	sub, err := dispatcher.Subscribe(ctx, "events.test", func(e Event) {
		received <- e
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	select {
	case event := <-received:
		if event.Name != "server.push" {
			t.Fatalf("unexpected event name: %s", event.Name)
		}
	case <-ctx.Done():
		t.Fatalf("timeout waiting for server event")
	}
	if err := dispatcher.Publish(ctx, "events.out", Event{Name: "client.push"}); err != nil {
		t.Fatalf("publish: %v", err)
	}
	select {
	case event := <-pubSeen:
		if event.Name != "client.push" {
			t.Fatalf("unexpected published event name: %s", event.Name)
		}
	case <-ctx.Done():
		t.Fatalf("timeout waiting for publish")
	}
	if err := sub.Unsubscribe(); err != nil {
		t.Fatalf("unsubscribe: %v", err)
	}
	if err := dispatcher.Close(ctx); err != nil {
		t.Fatalf("close: %v", err)
	}
	if err := <-serverErr; err != nil {
		t.Fatalf("server error: %v", err)
	}
}
