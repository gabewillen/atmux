package tui

import (
	"testing"
)

func TestDecodeScreen(t *testing.T) {
	raw := "Hello\x1b[31mWorld\x1b[0m\nLine2"
	screen := DecodeScreen([]byte(raw))
	
	if len(screen.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(screen.Rows))
	}
	if screen.Rows[0] != "HelloWorld" {
		t.Errorf("Expected row 0 'HelloWorld', got %q", screen.Rows[0])
	}
	if screen.Rows[1] != "Line2" {
		t.Errorf("Expected row 1 'Line2', got %q", screen.Rows[1])
	}
}

func TestEncodeXML(t *testing.T) {
	screen := &Screen{
		Rows: []string{"Row1", "Row2"},
	}
	
	data, err := screen.EncodeXML()
	if err != nil {
		t.Fatalf("EncodeXML failed: %v", err)
	}
	
	expected := `<screen><row>Row1</row><row>Row2</row></screen>`
	if string(data) != expected {
		t.Errorf("XML mismatch.\nWant: %s\nGot:  %s", expected, string(data))
	}
}
