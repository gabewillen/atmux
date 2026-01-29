package process

import (
	"context"
	"fmt"
	"strings"

	"github.com/agentflare-ai/amux/internal/inference"
)

// Gater uses an LLM to decide if an event should trigger a notification.
type Gater struct {
	Engine inference.LiquidgenEngine
}

// ShouldNotify returns true if the LLM thinks the event is noteworthy.
func (g *Gater) ShouldNotify(ctx context.Context, event Event) bool {
	if g.Engine == nil {
		return true // Default to notify
	}

	prompt := fmt.Sprintf("Process Event: %s\nPayload: %v\nShould the user be notified? Answer only YES or NO.", event.Type, event.Payload)
	
	stream, err := g.Engine.Generate(ctx, inference.LiquidgenRequest{
		Model:       "lfm2.5-thinking",
		Prompt:      prompt,
		MaxTokens:   5,
		Temperature: 0.0,
	})
	if err != nil {
		return true // Fallback
	}
	defer stream.Close()

	token, err := stream.Next()
	if err != nil {
		return true
	}

	clean := strings.ToUpper(strings.TrimSpace(token))
	return strings.Contains(clean, "YES")
}
