package tui

// Cell represents a single character cell on the terminal screen.
type Cell struct {
	Rune  rune
	Style Style
}

// Style represents the visual style of a cell.
type Style struct {
	FgColor int // ANSI color code or -1
	BgColor int // ANSI color code or -1
	Bold    bool
}

// Screen represents the state of the terminal screen.
type Screen struct {
	Rows    int
	Cols    int
	Cells   [][]Cell
	CursorX int
	CursorY int
}

// NewScreen creates a new Screen with the given dimensions.
func NewScreen(rows, cols int) *Screen {
	cells := make([][]Cell, rows)
	for i := range cells {
		cells[i] = make([]Cell, cols)
		for j := range cells[i] {
			cells[i][j] = Cell{Rune: ' '}
		}
	}
	return &Screen{
		Rows:  rows,
		Cols:  cols,
		Cells: cells,
	}
}
