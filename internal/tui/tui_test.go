package tui

import (
	"testing"
)

func TestTerminal(t *testing.T) {
	term := NewTerminal(2, 20)
	// Input with ANSI colors (SGR) which should be ignored/stripped by the parser
	input := "Hello\x1b[31mWorld\x1b[0m\r\nLine 2"
	term.Write([]byte(input))
	
	data, err := term.EncodeXML()
	if err != nil {
		t.Fatalf("EncodeXML failed: %v", err)
	}
	
	// The encoder trims trailing spaces.
	expected := `<screen><row>HelloWorld</row><row>Line 2</row></screen>`
	if string(data) != expected {
		t.Errorf("XML mismatch.\nWant: %s\nGot:  %s", expected, string(data))
	}
}
