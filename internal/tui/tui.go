// Package tui handles terminal screen decoding and XML encoding.
package tui

import (
	"encoding/xml"
	"strings"
)

// Screen represents the decoded state of a terminal screen.
type Screen struct {
	Rows []string `xml:"row"`
}

// DecodeScreen decodes raw PTY output into a Screen model.
// For Phase 5, this is a basic implementation that splits by newline
// and strips some common ANSI sequences (simplified).
func DecodeScreen(data []byte) *Screen {
	s := string(data)
	// Simplified: remove ANSI codes (very basic)
	// In a real implementation, we'd use a VT100 parser.
	// We'll strip CSI sequences for now.
	clean := stripANSI(s)
	

rows := strings.Split(clean, "\n")
	return &Screen{
		Rows: rows,
	}
}

// EncodeXML encodes the screen state to XML format.
func (s *Screen) EncodeXML() ([]byte, error) {
	type ScreenXML struct {
		XMLName xml.Name `xml:"screen"`
		Rows    []string `xml:"row"`
	}
	
	wrapper := ScreenXML{
		Rows: s.Rows,
	}
	return xml.Marshal(wrapper)
}

func stripANSI(str string) string {
	// Simple state machine to strip CSI (ESC [ ... final byte)
	var sb strings.Builder
	inEsc := false
	inCSI := false
	
	for i := 0; i < len(str); i++ {
		c := str[i]
		if !inEsc && !inCSI {
			if c == 0x1B { // ESC
				inEsc = true
			} else {
				sb.WriteByte(c)
			}
			continue
		}
		
		if inEsc {
			if c == '[' {
				inCSI = true
				inEsc = false
			} else {
				// Not a CSI, just a regular ESC sequence (e.g. ESC M), skip current char and reset
				// Simplified: we just drop the sequence char
				inEsc = false
			}
			continue
		}
		
		if inCSI {
			// End of CSI is byte 0x40-0x7E
			if c >= 0x40 && c <= 0x7E {
				inCSI = false
			}
			// Else continue skipping (parameter bytes)
		}
	}
	return sb.String()
}