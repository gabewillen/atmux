package tui

import (
	"testing"
)

func TestDecodeScreen(t *testing.T) {
	input := "Hello\x1b[31mWorld\x1b[0m\nLine 2"
	screen := DecodeScreen([]byte(input))
	
	if len(screen.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(screen.Rows))
	}
	if screen.Rows[0] != "HelloWorld" {
		t.Errorf("Expected 'HelloWorld' (stripped), got %q", screen.Rows[0])
	}
	if screen.Rows[1] != "Line 2" {
		t.Errorf("Expected 'Line 2', got %q", screen.Rows[1])
	}
}

func TestEncodeXML(t *testing.T) {
	screen := &Screen{
		Rows: []string{"Row 1", "Row 2"},
	}
	data, err := screen.EncodeXML()
	if err != nil {
		t.Fatalf("EncodeXML failed: %v", err)
	}
	
	expected := `<screen><row>Row 1</row><row>Row 2</row></screen>`
	if string(data) != expected {
		t.Errorf("XML mismatch.\nWant: %s\nGot:  %s", expected, string(data))
	}
}