package config

import (
	"testing"
	"time"
)

func TestParseTOMLNumbersAndDates(t *testing.T) {
	doc := `
int_dec = 1_234
int_hex = 0x10
int_oct = 0o10
int_bin = 0b10
float_val = 3.14
float_inf = inf
dt = 2024-01-01T12:30:00Z
date = 2024-01-01
time = 12:30:00
`
	parsed, err := ParseTOML([]byte(doc))
	if err != nil {
		t.Fatalf("parse toml: %v", err)
	}
	if parsed["int_dec"].(int64) != 1234 {
		t.Fatalf("unexpected int_dec")
	}
	if parsed["int_hex"].(int64) != 16 {
		t.Fatalf("unexpected int_hex")
	}
	if parsed["int_oct"].(int64) != 8 {
		t.Fatalf("unexpected int_oct")
	}
	if parsed["int_bin"].(int64) != 2 {
		t.Fatalf("unexpected int_bin")
	}
	if parsed["float_val"].(float64) != 3.14 {
		t.Fatalf("unexpected float_val")
	}
	if parsed["float_inf"].(float64) <= 0 {
		t.Fatalf("expected inf")
	}
	if _, ok := parsed["dt"].(time.Time); !ok {
		t.Fatalf("expected datetime")
	}
	if _, ok := parsed["date"].(time.Time); !ok {
		t.Fatalf("expected date")
	}
	if _, ok := parsed["time"].(time.Time); !ok {
		t.Fatalf("expected time")
	}
}

func TestParseTOMLInvalidNumber(t *testing.T) {
	doc := "bad = 0x\n"
	if _, err := ParseTOML([]byte(doc)); err == nil {
		t.Fatalf("expected parse error")
	}
}
