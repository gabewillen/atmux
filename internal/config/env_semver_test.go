package config

import "testing"

func TestEnvMapIncludesSet(t *testing.T) {
	t.Setenv("AMUX__FOO", "bar")
	env := EnvMap()
	if env["AMUX__FOO"] != "bar" {
		t.Fatalf("expected env map to include AMUX__FOO")
	}
}

func TestValidateSemverConstraint(t *testing.T) {
	if err := ValidateSemverConstraint(">=1.2.3 <2.0.0"); err != nil {
		t.Fatalf("expected valid constraint: %v", err)
	}
	if err := ValidateSemverConstraint(""); err == nil {
		t.Fatalf("expected empty constraint error")
	}
	if err := ValidateSemverConstraint("1.2.3"); err == nil {
		t.Fatalf("expected invalid constraint error")
	}
}

