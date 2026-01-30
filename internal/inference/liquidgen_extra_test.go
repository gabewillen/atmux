package inference

import "testing"

func TestLiquidgenVersion(t *testing.T) {
	engine := &LiquidgenEngine{version: "v1.2.3"}
	if engine.Version() != "v1.2.3" {
		t.Fatalf("unexpected version")
	}
}
