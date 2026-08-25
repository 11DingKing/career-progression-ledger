package service

import (
	"careerprogression/internal/domain"
	"fmt"
	"strings"
)

func ValidateRole(role domain.Role) error {
	switch role {
	case domain.RoleGraduate, domain.RoleEmployer, domain.RoleCounselor, domain.RoleMajorAdmin:
		return nil
	}
	return fmt.Errorf("unknown role %q", role)
}
func NormalizeMajor(v string) string { return strings.ToLower(strings.TrimSpace(v)) }
func IsTrustedSource(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "graduate", "employer", "counselor", "imported_verified":
		return true
	}
	return false
}
