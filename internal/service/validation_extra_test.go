package service

import (
	"careerprogression/internal/domain"
	"testing"
)

func TestValidateRole(t *testing.T) {
	for _, r := range []domain.Role{domain.RoleGraduate, domain.RoleEmployer, domain.RoleCounselor, domain.RoleMajorAdmin} {
		if ValidateRole(r) != nil {
			t.Fatal(r)
		}
	}
	if ValidateRole("invalid") == nil {
		t.Fatal("invalid role")
	}
}
func TestNormalizeMajor(t *testing.T) {
	if NormalizeMajor("  NETWORK ") != "network" {
		t.Fatal("normalize")
	}
	if NormalizeMajor("") != "" {
		t.Fatal("empty")
	}
}
func TestTrustedSources(t *testing.T) {
	for _, s := range []string{"graduate", "employer", "counselor", "imported_verified"} {
		if !IsTrustedSource(s) {
			t.Fatal(s)
		}
	}
	for _, s := range []string{"unknown", "", "manual"} {
		if IsTrustedSource(s) {
			t.Fatal(s)
		}
	}
}
func TestTrustedSourceCase(t *testing.T) {
	if !IsTrustedSource(" EMPLOYER ") {
		t.Fatal("case")
	}
}
func TestRoleStrings(t *testing.T) {
	values := []domain.Role{domain.RoleGraduate, domain.RoleEmployer, domain.RoleCounselor, domain.RoleMajorAdmin}
	for _, v := range values {
		if string(v) == "" {
			t.Fatal("empty")
		}
	}
}
func TestStatusStrings(t *testing.T) {
	values := []domain.EmploymentStatus{domain.EmploymentDraft, domain.EmploymentActive, domain.EmploymentEnded, domain.EmploymentDisputed}
	for _, v := range values {
		if string(v) == "" {
			t.Fatal("empty")
		}
	}
}
func TestAppealStrings(t *testing.T) {
	values := []domain.AppealStatus{domain.AppealOpen, domain.AppealReview, domain.AppealResolved, domain.AppealRejected}
	for _, v := range values {
		if string(v) == "" {
			t.Fatal("empty")
		}
	}
}
func TestValidationComposition(t *testing.T) {
	if ValidateRole(domain.RoleGraduate) != nil || NormalizeMajor(" m ") != "m" || !IsTrustedSource("employer") {
		t.Fatal("composition")
	}
}
