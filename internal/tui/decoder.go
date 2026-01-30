package tui

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"
)

// Config controls decoder sizing.
type Config struct {
	// Rows is the terminal row count.
	Rows int
	// Cols is the terminal column count.
	Cols int
}

// Decoder maintains a terminal screen model from PTY output.
type Decoder struct {
	mu          sync.Mutex
	rows        int
	cols        int
	alt         bool
	curX        int
	curY        int
	curVisible  bool
	screen      [][]rune
	wrapPending bool
	state       parseState
	paramsBuf   bytes.Buffer
}

type parseState int

const (
	stateNormal parseState = iota
	stateESC
	stateCSI
)

// NewDecoder constructs a new decoder with the provided size.
func NewDecoder(cfg Config) *Decoder {
	rows := cfg.Rows
	cols := cfg.Cols
	if rows <= 0 {
		rows = 24
	}
	if cols <= 0 {
		cols = 80
	}
	decoder := &Decoder{
		rows:       rows,
		cols:       cols,
		curVisible: true,
	}
	decoder.screen = makeScreen(rows, cols)
	return decoder
}

// Write feeds PTY output into the decoder.
func (d *Decoder) Write(data []byte) error {
	if d == nil || len(data) == 0 {
		return nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	for i := 0; i < len(data); {
		b := data[i]
		switch d.state {
		case stateNormal:
			if b == 0x1b {
				d.state = stateESC
				i++
				continue
			}
			if b == '\n' {
				d.newline()
				i++
				continue
			}
			if b == '\r' {
				d.curX = 0
				d.wrapPending = false
				i++
				continue
			}
			if b == '\b' {
				if d.curX > 0 {
					d.curX--
				}
				i++
				continue
			}
			if b == '\t' {
				if d.wrapPending {
					d.newline()
					d.wrapPending = false
				}
				d.curX = ((d.curX / 8) + 1) * 8
				if d.curX >= d.cols {
					d.newline()
				}
				i++
				continue
			}
			r, size := utf8.DecodeRune(data[i:])
			if r == utf8.RuneError && size == 1 {
				r = rune(b)
				size = 1
			}
			d.putRune(r)
			i += size
		case stateESC:
			if b == '[' {
				d.state = stateCSI
				d.paramsBuf.Reset()
				i++
				continue
			}
			d.state = stateNormal
		case stateCSI:
			if b >= 0x40 && b <= 0x7e {
				d.handleCSI(d.paramsBuf.String(), b)
				d.state = stateNormal
				i++
				continue
			}
			d.paramsBuf.WriteByte(b)
			i++
		}
	}
	return nil
}

// Resize updates the terminal geometry, preserving overlapping content.
func (d *Decoder) Resize(rows, cols int) {
	if d == nil {
		return
	}
	if rows <= 0 || cols <= 0 {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if rows == d.rows && cols == d.cols {
		return
	}
	next := makeScreen(rows, cols)
	minRows := d.rows
	if rows < minRows {
		minRows = rows
	}
	minCols := d.cols
	if cols < minCols {
		minCols = cols
	}
	for y := 0; y < minRows; y++ {
		copy(next[y][:minCols], d.screen[y][:minCols])
	}
	d.rows = rows
	d.cols = cols
	d.screen = next
	if d.curY >= rows {
		d.curY = rows - 1
	}
	if d.curX >= cols {
		d.curX = cols - 1
	}
}

// EncodeXML returns the current screen as TUI XML.
func (d *Decoder) EncodeXML() string {
	if d == nil {
		return ""
	}
	d.mu.Lock()
	rows := d.rows
	cols := d.cols
	alt := d.alt
	curX := d.curX
	curY := d.curY
	curVis := d.curVisible
	screen := cloneScreen(d.screen)
	d.mu.Unlock()
	var b strings.Builder
	b.WriteString(`<tui v="1" cols="`)
	b.WriteString(strconv.Itoa(cols))
	b.WriteString(`" rows="`)
	b.WriteString(strconv.Itoa(rows))
	b.WriteString(`" alt="`)
	if alt {
		b.WriteString("1")
	} else {
		b.WriteString("0")
	}
	b.WriteString(`" cur_x="`)
	b.WriteString(strconv.Itoa(curX))
	b.WriteString(`" cur_y="`)
	b.WriteString(strconv.Itoa(curY))
	b.WriteString(`" cur_vis="`)
	if curVis {
		b.WriteString("1")
	} else {
		b.WriteString("0")
	}
	b.WriteString(`">`)
	for y := 0; y < rows; y++ {
		row := screen[y]
		last := lastNonSpace(row)
		if last < 0 {
			continue
		}
		b.WriteString(`<row y="`)
		b.WriteString(strconv.Itoa(y))
		b.WriteString(`">`)
		x := 0
		for x <= last {
			ch := row[x]
			runStart := x
			runLen := 1
			for runStart+runLen <= last && row[runStart+runLen] == ch {
				runLen++
			}
			if ch == ' ' {
				x += runLen
				continue
			}
			b.WriteString(`<r`)
			if runStart > 0 {
				b.WriteString(` x="`)
				b.WriteString(strconv.Itoa(runStart))
				b.WriteString(`"`)
			}
			b.WriteString(` c="`)
			b.WriteString(strconv.Itoa(runLen))
			b.WriteString(`"`)
			if runLen > 1 {
				b.WriteString(` ch="`)
				if err := xml.EscapeText(&b, []byte(string(ch))); err != nil {
					b.WriteString("?")
				}
				b.WriteString(`"`)
				b.WriteString(`/>`)
			} else {
				b.WriteString(`>`)
				if err := xml.EscapeText(&b, []byte(string(ch))); err != nil {
					b.WriteString("?")
				}
				b.WriteString(`</r>`)
			}
			x += runLen
		}
		b.WriteString(`</row>`)
	}
	b.WriteString(`</tui>`)
	return b.String()
}

func (d *Decoder) putRune(r rune) {
	if d.curY < 0 || d.curY >= d.rows {
		return
	}
	if d.curX < 0 {
		d.curX = 0
	}
	if d.wrapPending {
		d.newline()
		d.wrapPending = false
	}
	if d.curY < 0 || d.curY >= d.rows || d.curX < 0 || d.curX >= d.cols {
		return
	}
	d.screen[d.curY][d.curX] = r
	if d.curX == d.cols-1 {
		d.wrapPending = true
		return
	}
	d.curX++
}

func (d *Decoder) newline() {
	d.curX = 0
	d.curY++
	d.wrapPending = false
	if d.curY >= d.rows {
		d.scrollUp()
		d.curY = d.rows - 1
	}
}

func (d *Decoder) scrollUp() {
	if d.rows <= 1 {
		clearRow(d.screen[0])
		return
	}
	copy(d.screen[0:], d.screen[1:])
	d.screen[d.rows-1] = makeRow(d.cols)
}

func (d *Decoder) handleCSI(params string, final byte) {
	switch final {
	case 'H', 'f':
		row, col := parseRowCol(params)
		d.curY = clamp(row-1, 0, d.rows-1)
		d.curX = clamp(col-1, 0, d.cols-1)
	case 'J':
		d.clearScreen()
	case 'K':
		d.clearLine()
	case 'm':
		// SGR attributes ignored in this minimal decoder.
	case 'h', 'l':
		d.handleMode(params, final)
	}
}

func (d *Decoder) handleMode(params string, final byte) {
	if !strings.HasPrefix(params, "?") {
		return
	}
	mode := strings.TrimPrefix(params, "?")
	switch mode {
	case "1049", "47", "1047":
		d.alt = final == 'h'
		d.clearScreen()
	case "25":
		d.curVisible = final == 'h'
	}
}

func (d *Decoder) clearScreen() {
	for _, row := range d.screen {
		clearRow(row)
	}
	d.curX = 0
	d.curY = 0
	d.wrapPending = false
}

func (d *Decoder) clearLine() {
	if d.curY < 0 || d.curY >= d.rows {
		return
	}
	row := d.screen[d.curY]
	for i := d.curX; i < len(row); i++ {
		row[i] = ' '
	}
	d.wrapPending = false
}

func parseRowCol(params string) (int, int) {
	if params == "" {
		return 1, 1
	}
	parts := strings.Split(params, ";")
	row := parseCSIInt(parts, 0, 1)
	col := parseCSIInt(parts, 1, 1)
	return row, col
}

func parseCSIInt(parts []string, index int, fallback int) int {
	if index >= len(parts) {
		return fallback
	}
	value, err := strconv.Atoi(parts[index])
	if err != nil || value == 0 {
		return fallback
	}
	return value
}

func makeScreen(rows, cols int) [][]rune {
	screen := make([][]rune, rows)
	for y := 0; y < rows; y++ {
		screen[y] = makeRow(cols)
	}
	return screen
}

func makeRow(cols int) []rune {
	row := make([]rune, cols)
	for i := range row {
		row[i] = ' '
	}
	return row
}

func clearRow(row []rune) {
	for i := range row {
		row[i] = ' '
	}
}

func cloneScreen(screen [][]rune) [][]rune {
	out := make([][]rune, len(screen))
	for i := range screen {
		out[i] = append([]rune(nil), screen[i]...)
	}
	return out
}

func lastNonSpace(row []rune) int {
	for i := len(row) - 1; i >= 0; i-- {
		if row[i] != ' ' {
			return i
		}
	}
	return -1
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func (d *Decoder) String() string {
	if d == nil {
		return ""
	}
	return fmt.Sprintf("tui(%dx%d)", d.rows, d.cols)
}
