package remote

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
)

// NewPTYConn returns a net.Conn that bridges PTY I/O over NATS.
func NewPTYConn(ctx context.Context, dispatcher protocol.Dispatcher, prefix string, hostID api.HostID, sessionID api.SessionID) (net.Conn, error) {
	if dispatcher == nil {
		return nil, fmt.Errorf("pty conn: dispatcher is nil")
	}
	if hostID == "" || sessionID.IsZero() {
		return nil, fmt.Errorf("pty conn: %w", ErrInvalidMessage)
	}
	local, remote := net.Pipe()
	outSubject := PtyOutSubject(prefix, hostID, sessionID)
	inSubject := PtyInSubject(prefix, hostID, sessionID)
	sub, err := dispatcher.SubscribeRaw(ctx, outSubject, func(msg protocol.Message) {
		if len(msg.Data) == 0 {
			return
		}
		_, _ = local.Write(msg.Data)
	})
	if err != nil {
		_ = local.Close()
		_ = remote.Close()
		return nil, fmt.Errorf("pty conn: %w", err)
	}
	go func() {
		buf := make([]byte, 4096)
		for {
			n, readErr := local.Read(buf)
			if n > 0 {
				chunks := chunkBytes(dispatcher.MaxPayload(), buf[:n])
				for _, chunk := range chunks {
					_ = dispatcher.PublishRaw(context.Background(), inSubject, chunk, "")
				}
			}
			if readErr != nil {
				if readErr != io.EOF {
					_ = local.Close()
				}
				_ = sub.Unsubscribe()
				return
			}
		}
	}()
	return remote, nil
}

func chunkBytes(maxPayload int, data []byte) [][]byte {
	if len(data) == 0 {
		return nil
	}
	max := maxPayload
	if max <= 0 {
		max = 1024 * 1024
	}
	if len(data) <= max {
		return [][]byte{append([]byte(nil), data...)}
	}
	chunks := make([][]byte, 0, (len(data)/max)+1)
	for len(data) > 0 {
		end := max
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, append([]byte(nil), data[:end]...))
		data = data[end:]
	}
	return chunks
}
