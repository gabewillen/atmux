package config

import (
	"fmt"
	"regexp"
	"strings"
)

var semverConstraintPattern = regexp.MustCompile(`^(=|==|!=|>=|<=|>|<)(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:[-+][0-9A-Za-z.-]+)?$`)

// ValidateSemverConstraint validates a conjunction of semver comparisons.
func ValidateSemverConstraint(expr string) error {
	fields := strings.Fields(expr)
	if len(fields) == 0 {
		return fmt.Errorf("empty constraint")
	}
	for _, field := range fields {
		if !semverConstraintPattern.MatchString(field) {
			return fmt.Errorf("invalid constraint: %s", field)
		}
	}
	return nil
}
