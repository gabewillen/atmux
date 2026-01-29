package rpc

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

// Handler processes a JSON-RPC request.
type Handler func(context.Context, json.RawMessage) (any, *Error)

// Server hosts JSON-RPC handlers over a stream transport.
type Server struct {
	mu       sync.RWMutex
	handlers map[string]Handler
	logger   *log.Logger
}

// NewServer constructs a JSON-RPC server.
func NewServer(logger *log.Logger) *Server {
	return &Server{handlers: make(map[string]Handler), logger: logger}
}

// Register registers a handler for a method.
func (s *Server) Register(method string, handler Handler) {
	if s == nil || method == "" || handler == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[method] = handler
}

// Serve accepts connections and handles requests.
func (s *Server) Serve(ctx context.Context, listener net.Listener) error {
	if s == nil {
		return fmt.Errorf("rpc serve: server is nil")
	}
	if listener == nil {
		return fmt.Errorf("rpc serve: listener is nil")
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
			}
			return fmt.Errorf("rpc accept: %w", err)
		}
		go s.serveConn(ctx, conn)
	}
}

func (s *Server) serveConn(ctx context.Context, conn net.Conn) {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	defer func() {
		if err := conn.Close(); err != nil && s.logger != nil {
			s.logger.Printf("rpc close: %v", err)
		}
	}()
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var req Request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			s.writeError(writer, nil, CodeParseError, "parse error")
			continue
		}
		if req.JSONRPC != "2.0" || req.Method == "" {
			s.writeError(writer, req.ID, CodeInvalidRequest, "invalid request")
			continue
		}
		handler := s.handler(req.Method)
		if handler == nil {
			s.writeError(writer, req.ID, CodeMethodNotFound, "method not found")
			continue
		}
		result, rpcErr := handler(ctx, req.Params)
		if len(req.ID) == 0 {
			continue
		}
		resp := Response{JSONRPC: "2.0", ID: req.ID}
		if rpcErr != nil {
			resp.Error = rpcErr
		} else {
			resp.Result = result
		}
		if err := writeJSON(writer, resp); err != nil {
			return
		}
	}
}

func (s *Server) handler(method string) Handler {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.handlers[method]
}

func (s *Server) writeError(writer *bufio.Writer, id json.RawMessage, code int, message string) {
	resp := Response{JSONRPC: "2.0", ID: id, Error: &Error{Code: code, Message: message}}
	_ = writeJSON(writer, resp)
}

func writeJSON(writer *bufio.Writer, value any) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if _, err := writer.Write(payload); err != nil {
		return err
	}
	if _, err := writer.WriteString("\n"); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	return nil
}
