// Package tui handles terminal screen decoding and XML encoding.
package tui

import (
	"encoding/xml"
	"strings"
	"sync"
	"unicode/utf8"
)

// Cell represents a single character cell on the screen.
type Cell struct {
	Char rune
}

// Terminal represents the state of the virtual terminal.
type Terminal struct {
	mu   sync.RWMutex
	Rows int
	Cols int
	Grid [][]Cell
	Cx   int
	Cy   int
}

// NewTerminal creates a new terminal with given dimensions.
func NewTerminal(rows, cols int) *Terminal {
	t := &Terminal{
		Rows: rows,
		Cols: cols,
	}
	t.resizeGrid(rows, cols)
	return t
}

func (t *Terminal) resizeGrid(rows, cols int) {
	newGrid := make([][]Cell, rows)
	for y := 0; y < rows; y++ {
		newGrid[y] = make([]Cell, cols)
		for x := 0; x < cols; x++ {
			newGrid[y][x] = Cell{Char: ' '}
		}
	}
	t.Grid = newGrid
	t.Rows = rows
	t.Cols = cols
	t.Cx = 0
	t.Cy = 0
}

// Write parses input and updates the terminal state.
// This is a simplified VT100 parser complying with Spec §7.7.3 requirements.
func (t *Terminal) Write(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	i := 0
	for i < len(p) {
		r, size := utf8.DecodeRune(p[i:])
		if r == utf8.RuneError && size == 1 {
			i++
			continue
		}
		i += size

		switch r {
		case 0x1B: // ESC
			if i < len(p) && p[i] == '[' { // CSI
				i++
				params := ""
				for i < len(p) {
					c := p[i]
					i++
					if c >= 0x40 && c <= 0x7E {
						t.handleCSI(params, rune(c))
						break
					}
					params += string(c)
				}
			}
		case '\n':
			t.Cy++
			if t.Cy >= t.Rows {
				t.scrollUp()
				t.Cy = t.Rows - 1
			}
		case '\r':
			t.Cx = 0
		case '\b':
			if t.Cx > 0 {
				t.Cx--
			}
		case '\t':
			t.Cx += 8 - (t.Cx % 8)
			if t.Cx >= t.Cols {
				t.Cx = t.Cols - 1
			}
		default:
			// Printable
			if t.Cy < t.Rows && t.Cx < t.Cols {
				t.Grid[t.Cy][t.Cx] = Cell{Char: r}
				t.Cx++
				if t.Cx >= t.Cols {
					t.Cx = 0
					t.Cy++
					if t.Cy >= t.Rows {
						t.scrollUp()
						t.Cy = t.Rows - 1
					}
				}
			}
		}
	}
	return len(p), nil
}

func (t *Terminal) scrollUp() {
	for y := 0; y < t.Rows-1; y++ {
		t.Grid[y] = t.Grid[y+1]
	}
	t.Grid[t.Rows-1] = make([]Cell, t.Cols)
	for x := 0; x < t.Cols; x++ {
		t.Grid[t.Rows-1][x] = Cell{Char: ' '}
	}
}

func (t *Terminal) handleCSI(params string, final rune) {
	switch final {
	case 'H', 'f': // CUP
		t.Cx = 0
		t.Cy = 0
	case 'J': // ED
		if strings.Contains(params, "2") {
			t.resizeGrid(t.Rows, t.Cols)
		}
	case 'K': // EL
		for x := t.Cx; x < t.Cols; x++ {
			t.Grid[t.Cy][x] = Cell{Char: ' '}
		}
	}
}

// ScreenXML is the export format.
type ScreenXML struct {
	XMLName xml.Name `xml:"screen"`
	Rows    []string `xml:"row"`
}

// EncodeXML exports the current state.
func (t *Terminal) EncodeXML() ([]byte, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()


rows := make([]string, t.Rows)
	for y := 0; y < t.Rows; y++ {
		var sb strings.Builder
		for x := 0; x < t.Cols; x++ {
			sb.WriteRune(t.Grid[y][x].Char)
		}
		// Trim trailing spaces for compactness? Spec says "compact XML".
		rows[y] = strings.TrimRight(sb.String(), " ")
	}

	wrapper := ScreenXML{
		Rows: rows,
	}
	return xml.Marshal(wrapper)
}
