package api

import (
	"encoding/json"
	"testing"
)

func TestLocationTypeJSONExtra(t *testing.T) {
	loc, err := ParseLocationType("local")
	if err != nil {
		t.Fatalf("parse location: %v", err)
	}
	data, err := json.Marshal(loc)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(data) != "\"local\"" {
		t.Fatalf("unexpected marshal: %s", data)
	}
	var decoded LocationType
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded != LocationLocal {
		t.Fatalf("unexpected decode: %v", decoded)
	}
	if _, err := ParseLocationType("unknown"); err == nil {
		t.Fatalf("expected invalid location error")
	}
	if err := json.Unmarshal([]byte("\"nope\""), &decoded); err == nil {
		t.Fatalf("expected invalid location json error")
	}
}
