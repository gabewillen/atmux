// Package tui provides terminal screen decoding and TUI XML encoding for amux.
//
// The TUI decoder incrementally processes PTY output to build a terminal
// screen model, which can be serialized to XML for LLM ingestion.
//
// See spec §7.7 and §11.2.1 for TUI decoding requirements.
package tui

import (
	"encoding/xml"
	"strings"
)

// Screen represents the current terminal screen state.
type Screen struct {
	Rows    int
	Cols    int
	Cells   [][]Cell
	CursorX int
	CursorY int
}

// Cell represents a single terminal cell.
type Cell struct {
	Char  rune
	Style Style
}

// Style represents cell styling.
type Style struct {
	Bold       bool
	Dim        bool
	Italic     bool
	Underline  bool
	Foreground string
	Background string
}

// NewScreen creates a new screen with the given dimensions.
func NewScreen(rows, cols int) *Screen {
	cells := make([][]Cell, rows)
	for i := range cells {
		cells[i] = make([]Cell, cols)
		for j := range cells[i] {
			cells[i][j] = Cell{Char: ' '}
		}
	}

	return &Screen{
		Rows:  rows,
		Cols:  cols,
		Cells: cells,
	}
}

// Resize resizes the screen.
func (s *Screen) Resize(rows, cols int) {
	newCells := make([][]Cell, rows)
	for i := range newCells {
		newCells[i] = make([]Cell, cols)
		for j := range newCells[i] {
			if i < len(s.Cells) && j < len(s.Cells[i]) {
				newCells[i][j] = s.Cells[i][j]
			} else {
				newCells[i][j] = Cell{Char: ' '}
			}
		}
	}

	s.Rows = rows
	s.Cols = cols
	s.Cells = newCells
}

// Clear clears the screen.
func (s *Screen) Clear() {
	for i := range s.Cells {
		for j := range s.Cells[i] {
			s.Cells[i][j] = Cell{Char: ' '}
		}
	}
	s.CursorX = 0
	s.CursorY = 0
}

// xmlLine represents a line in XML output.
type xmlLine struct {
	Row     int    `xml:"row,attr"`
	Content string `xml:",chardata"`
}

// xmlScreen represents the screen in XML output.
type xmlScreen struct {
	XMLName xml.Name  `xml:"screen"`
	Rows    int       `xml:"rows,attr"`
	Cols    int       `xml:"cols,attr"`
	Lines   []xmlLine `xml:"line"`
}

// ToXML serializes the screen to XML.
func (s *Screen) ToXML() ([]byte, error) {
	lines := make([]xmlLine, s.Rows)
	for i := range s.Cells {
		var sb strings.Builder
		for _, cell := range s.Cells[i] {
			sb.WriteRune(cell.Char)
		}
		lines[i] = xmlLine{
			Row:     i,
			Content: strings.TrimRight(sb.String(), " "),
		}
	}

	screen := xmlScreen{
		Rows:  s.Rows,
		Cols:  s.Cols,
		Lines: lines,
	}

	return xml.MarshalIndent(screen, "", "  ")
}

// Decoder decodes PTY output into a screen model.
type Decoder struct {
	screen *Screen
}

// NewDecoder creates a new decoder.
func NewDecoder(rows, cols int) *Decoder {
	return &Decoder{
		screen: NewScreen(rows, cols),
	}
}

// Write processes PTY output.
func (d *Decoder) Write(data []byte) (int, error) {
	// Simplified decoder - just writes characters
	// A full implementation would handle ANSI escape sequences
	for _, b := range data {
		if b == '\n' {
			d.screen.CursorY++
			d.screen.CursorX = 0
			if d.screen.CursorY >= d.screen.Rows {
				d.screen.CursorY = d.screen.Rows - 1
			}
		} else if b == '\r' {
			d.screen.CursorX = 0
		} else if b >= 32 && b < 127 {
			if d.screen.CursorY < len(d.screen.Cells) && d.screen.CursorX < len(d.screen.Cells[d.screen.CursorY]) {
				d.screen.Cells[d.screen.CursorY][d.screen.CursorX] = Cell{Char: rune(b)}
			}
			d.screen.CursorX++
			if d.screen.CursorX >= d.screen.Cols {
				d.screen.CursorX = 0
				d.screen.CursorY++
				if d.screen.CursorY >= d.screen.Rows {
					d.screen.CursorY = d.screen.Rows - 1
				}
			}
		}
	}
	return len(data), nil
}

// Screen returns the current screen state.
func (d *Decoder) Screen() *Screen {
	return d.screen
}

// Reset resets the decoder.
func (d *Decoder) Reset() {
	d.screen.Clear()
}
