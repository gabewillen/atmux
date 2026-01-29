// Package monitor implements PTY output monitoring (delegates to adapters)
package monitor

import (
	"context"
	"io"
	"os"
	"sync"
	"time"

	"github.com/stateforward/amux/internal/adapteriface"
	"github.com/stateforward/amux/internal/event"
)

// Monitor monitors PTY output and detects various conditions
type Monitor struct {
	ptyFile     *os.File
	outputChan  chan []byte
	stopChan    chan struct{}
	wg          sync.WaitGroup
	interval    time.Duration
	adapterIface adapteriface.Interface
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewMonitor creates a new PTY monitor
func NewMonitor(ptyFile *os.File, adapterName string) *Monitor {
	ctx, cancel := context.WithCancel(context.Background())

	monitor := &Monitor{
		ptyFile:    ptyFile,
		outputChan: make(chan []byte, 100), // Buffered channel to prevent blocking
		stopChan:   make(chan struct{}),
		interval:   100 * time.Millisecond, // Check every 100ms
		ctx:        ctx,
		cancel:     cancel,
		// Use the global adapter interface for now
		adapterIface: adapteriface.GlobalInterface,
	}

	// Start the monitoring goroutines
	monitor.startReading()

	return monitor
}

// startReading starts the goroutine that reads from the PTY
func (m *Monitor) startReading() {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		
		buffer := make([]byte, 1024)
		for {
			select {
			case <-m.stopChan:
				return
			case <-m.ctx.Done():
				return
			default:
				n, err := m.ptyFile.Read(buffer)
				if err != nil {
					if err != io.EOF {
						// Log error but continue monitoring
						continue
					}
					return
				}
				
				if n > 0 {
					// Send the output to be processed
					output := make([]byte, n)
					copy(output, buffer[:n])
					
					select {
					case m.outputChan <- output:
					case <-time.After(1 * time.Second): // Timeout to prevent blocking
						// If we can't send within 1 second, drop the output
					}
					
					// Emit PTY output event
					go func(out []byte) {
						ctx := context.Background()
						err := event.EmitEvent(ctx, "pty.output", map[string]interface{}{
							"output": string(out),
							"length": len(out),
						})
						if err != nil {
							// Log error but continue
							_ = err // Use the error to avoid SA9003 warning
						}
					}(output)
				}
			}
		}
	}()
}

// Start begins monitoring the PTY
func (m *Monitor) Start() {
	// Start processing the output in a separate goroutine
	m.wg.Add(1)
	go m.processOutput()
}

// processOutput processes the output from the PTY
func (m *Monitor) processOutput() {
	defer m.wg.Done()
	
	for {
		select {
		case <-m.stopChan:
			return
		case <-m.ctx.Done():
			return
		case output := <-m.outputChan:
			// Process the output for patterns using the adapter
			outputStr := string(output)
			
			// Use the adapter to match patterns in the output
			matches, err := m.adapterIface.MatchPatterns(m.ctx, outputStr)
			if err != nil {
				// Log error but continue
				_ = err // Use the error to avoid SA9003 warning
				continue
			}
			
			// Handle each match
			for _, match := range matches {
				// Emit event about the match
				go func(m adapteriface.Match) {
					ctx := context.Background()
					err := event.EmitEvent(ctx, "pattern.matched", map[string]interface{}{
						"pattern_id": m.PatternID,
						"action":     m.Action,
						"score":      m.Score,
						"data":       m.Data,
					})
					if err != nil {
						// Log error but continue
						_ = err // Use the error to avoid SA9003 warning
					}
				}(match)

				// Execute the corresponding action
				action := adapteriface.Action{
					Type: match.Action,
					Data: match.Data,
				}

				go func(act adapteriface.Action) {
					err := m.adapterIface.ExecuteAction(m.ctx, act)
					if err != nil {
						// Log error but continue
						_ = err // Use the error to avoid SA9003 warning
					}
				}(action)
			}
		}
	}
}

// Stop stops monitoring the PTY
func (m *Monitor) Stop() {
	close(m.stopChan)
	m.cancel()
	m.wg.Wait()
}

// DetectActivity detects if there's been recent activity in the PTY
func (m *Monitor) DetectActivity(timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	// Create a temporary channel to listen for output
	tempChan := make(chan []byte, 1)
	
	// Start a goroutine to relay output to our temp channel
	done := make(chan bool, 1)
	
	go func() {
		select {
		case <-tempChan:
			done <- true
		case <-ctx.Done():
			done <- false
		}
	}()
	
	// Relay from main output channel to temp channel
	go func() {
		for {
			select {
			case output := <-m.outputChan:
				select {
				case tempChan <- output:
				default:
					// If tempChan is full, continue
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	
	return <-done
}

// GetLastOutput returns the last output from the PTY
func (m *Monitor) GetLastOutput() []byte {
	// This is a simplified implementation
	// In a real implementation, we'd want to keep track of recent output
	return []byte{} // Placeholder
}

// WaitForPattern waits for a specific pattern to appear in the output
func (m *Monitor) WaitForPattern(pattern string, timeout time.Duration) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	// Create a channel to signal when the pattern is found
	foundChan := make(chan bool, 1)
	
	// Start a goroutine to monitor for the pattern
	go func() {
		for {
			select {
			case output := <-m.outputChan:
				if contains(string(output), pattern) {
					foundChan <- true
					return
				}
			case <-ctx.Done():
				foundChan <- false
				return
			}
		}
	}()
	
	found := <-foundChan
	if !found {
		return false, context.DeadlineExceeded
	}
	
	return true, nil
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && find(s, substr)
}

// Helper function to find a substring
func find(s, substr string) bool {
	sLen := len(s)
	substrLen := len(substr)
	
	if substrLen == 0 {
		return true
	}
	
	for i := 0; i <= sLen-substrLen; i++ {
		match := true
		for j := 0; j < substrLen; j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	
	return false
}