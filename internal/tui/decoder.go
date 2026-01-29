package tui

// Decoder decodes a stream of bytes into a Screen.
type Decoder struct {
	screen *Screen
}

// NewDecoder creates a new decoder for the given screen.
func NewDecoder(screen *Screen) *Decoder {
	return &Decoder{screen: screen}
}

// Write parses input bytes and updates the screen.
// This is a simplified stub for Phase 5. In a real impl, this would parse ANSI codes.
// For now, it just writes raw chars and handles basic newlines.
func (d *Decoder) Write(p []byte) (n int, err error) {
	for _, b := range p {
		d.processByte(b)
	}
	return len(p), nil
}

func (d *Decoder) processByte(b byte) {
	// Very basic terminal emulation stub
	switch b {
	case '\n':
		d.screen.CursorY++
		d.screen.CursorX = 0
	case '\r':
		d.screen.CursorX = 0
	default:
		if d.screen.CursorY < d.screen.Rows && d.screen.CursorX < d.screen.Cols {
			d.screen.Cells[d.screen.CursorY][d.screen.CursorX] = Cell{Rune: rune(b)}
			d.screen.CursorX++
		}
	}
	// Scroll if needed (omitted for stub)
}
