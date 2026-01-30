package tui

import (
	"strings"
	"testing"
)

func TestDecoderPlainText(t *testing.T) {
	dec := NewDecoder(Config{Rows: 2, Cols: 10})
	if err := dec.Write([]byte("hello")); err != nil {
		t.Fatalf("write: %v", err)
	}
	xml := dec.EncodeXML()
	if !strings.Contains(xml, ">h</r>") || !strings.Contains(xml, `ch="l"`) || !strings.Contains(xml, ">o</r>") {
		t.Fatalf("expected xml runs, got %s", xml)
	}
}

func TestDecoderAltScreen(t *testing.T) {
	dec := NewDecoder(Config{Rows: 1, Cols: 5})
	_ = dec.Write([]byte("\x1b[?1049h"))
	xml := dec.EncodeXML()
	if !strings.Contains(xml, `alt="1"`) {
		t.Fatalf("expected alt screen on, got %s", xml)
	}
	_ = dec.Write([]byte("\x1b[?1049l"))
	xml = dec.EncodeXML()
	if !strings.Contains(xml, `alt="0"`) {
		t.Fatalf("expected alt screen off, got %s", xml)
	}
}

func TestDecoderResizePreserves(t *testing.T) {
	dec := NewDecoder(Config{Rows: 1, Cols: 5})
	_ = dec.Write([]byte("abcde"))
	dec.Resize(1, 3)
	xml := dec.EncodeXML()
	if !strings.Contains(xml, ">a</r>") || !strings.Contains(xml, ">b</r>") {
		t.Fatalf("expected preserved content, got %s", xml)
	}
}

func TestDecoderControlsAndClear(t *testing.T) {
	dec := NewDecoder(Config{Rows: 1, Cols: 6})
	_ = dec.Write([]byte("abc"))
	_ = dec.Write([]byte("\x1b[1;2H"))
	_ = dec.Write([]byte("\x1b[K"))
	xml := dec.EncodeXML()
	if !strings.Contains(xml, ">a</r>") {
		t.Fatalf("expected remaining text, got %s", xml)
	}
	if strings.Contains(xml, ">b</r>") || strings.Contains(xml, ">c</r>") {
		t.Fatalf("expected cleared line, got %s", xml)
	}
	_ = dec.Write([]byte("\x1b[J"))
	xml = dec.EncodeXML()
	if strings.Contains(xml, "<row") {
		t.Fatalf("expected cleared screen, got %s", xml)
	}
}

func TestDecoderCursorVisibility(t *testing.T) {
	dec := NewDecoder(Config{Rows: 1, Cols: 4})
	_ = dec.Write([]byte("\x1b[?25l"))
	xml := dec.EncodeXML()
	if !strings.Contains(xml, `cur_vis="0"`) {
		t.Fatalf("expected cursor hidden, got %s", xml)
	}
	_ = dec.Write([]byte("\x1b[?25h"))
	xml = dec.EncodeXML()
	if !strings.Contains(xml, `cur_vis="1"`) {
		t.Fatalf("expected cursor visible, got %s", xml)
	}
}

func TestDecoderBackspaceTabWrap(t *testing.T) {
	dec := NewDecoder(Config{Rows: 1, Cols: 12})
	_ = dec.Write([]byte("ab\bC"))
	xml := dec.EncodeXML()
	if !strings.Contains(xml, ">a</r>") || !strings.Contains(xml, ">C</r>") {
		t.Fatalf("expected backspace replace, got %s", xml)
	}
	if strings.Contains(xml, ">b</r>") {
		t.Fatalf("expected backspace to overwrite b, got %s", xml)
	}
	_ = dec.Write([]byte("\ra\tb"))
	xml = dec.EncodeXML()
	if !strings.Contains(xml, `x="8"`) {
		t.Fatalf("expected tab expansion, got %s", xml)
	}
}

func TestDecoderScrollWrap(t *testing.T) {
	dec := NewDecoder(Config{Rows: 1, Cols: 3})
	_ = dec.Write([]byte("abcd"))
	xml := dec.EncodeXML()
	if !strings.Contains(xml, ">d</r>") {
		t.Fatalf("expected scroll wrap, got %s", xml)
	}
	if strings.Contains(xml, ">a</r>") || strings.Contains(xml, ">b</r>") || strings.Contains(xml, ">c</r>") {
		t.Fatalf("expected old content cleared, got %s", xml)
	}
}

func TestDecoderNilReceiver(t *testing.T) {
	var dec *Decoder
	if err := dec.Write([]byte("hi")); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if dec.EncodeXML() != "" {
		t.Fatalf("expected empty xml for nil decoder")
	}
	if dec.String() != "" {
		t.Fatalf("expected empty string for nil decoder")
	}
}
