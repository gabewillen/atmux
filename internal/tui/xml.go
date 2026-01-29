package tui

import (
	"encoding/xml"
)

// ToXML serializes the screen to the compact XML format.
func (s *Screen) ToXML() ([]byte, error) {
	type XCell struct {
		Char string `xml:",chardata"`
	}

	type XRow struct {
		Cells []XCell `xml:"c"`
	}

	type XScreen struct {
		XMLName xml.Name `xml:"screen"`
		Rows    int      `xml:"rows,attr"`
		Cols    int      `xml:"cols,attr"`
		CursorX int      `xml:"cursor_x,attr"`
		CursorY int      `xml:"cursor_y,attr"`
		Lines   []XRow   `xml:"l"`
	}

	xLines := make([]XRow, 0, s.Rows)
	for _, row := range s.Cells {
		xRow := XRow{Cells: make([]XCell, 0, len(row))}
		for _, cell := range row {
			xRow.Cells = append(xRow.Cells, XCell{Char: string(cell.Rune)})
		}
		xLines = append(xLines, xRow)
	}

	xScreen := XScreen{
		Rows:    s.Rows,
		Cols:    s.Cols,
		CursorX: s.CursorX,
		CursorY: s.CursorY,
		Lines:   xLines,
	}

	return xml.Marshal(xScreen)
}
