package remote

import (
	"context"

	"github.com/agentflare-ai/amux/internal/protocol"
)

type staticFormatter struct {
	prefix string
}

func (s staticFormatter) Format(ctx context.Context, input string) (string, error) {
	_ = ctx
	return s.prefix + input, nil
}

func protocolMessage(subject string, data []byte) protocol.Message {
	return protocol.Message{Subject: subject, Data: data}
}
