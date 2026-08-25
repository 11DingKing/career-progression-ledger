package service

import (
	"careerprogression/internal/domain"
	"fmt"
)

type Visibility string

const (
	VisibilityPublic     Visibility = "public"
	VisibilityRestricted Visibility = "restricted"
	VisibilityPrivate    Visibility = "private"
)

func VisibleRole(role domain.Role, consent bool) Visibility {
	if !consent {
		return VisibilityPrivate
	}
	switch role {
	case domain.RoleMajorAdmin, domain.RoleCounselor:
		return VisibilityRestricted
	case domain.RoleEmployer:
		return VisibilityPublic
	default:
		return VisibilityPrivate
	}
}
func RedactGraduate(g domain.Graduate, v Visibility) domain.Graduate {
	if v == VisibilityPrivate {
		g.Contact = ""
	}
	if v == VisibilityPublic {
		g.Contact = "masked"
		g.StudentNo = "masked"
	}
	return g
}
func CanEdit(role domain.Role, owner bool, field string) bool {
	if owner && field != "verification" {
		return true
	}
	if role == domain.RoleMajorAdmin {
		return true
	}
	if role == domain.RoleEmployer && field == "feedback" {
		return true
	}
	return false
}
func CheckAccess(role domain.Role, resource string) error {
	if role == "" {
		return fmt.Errorf("role required")
	}
	if resource == "salary" && role == domain.RoleEmployer {
		return fmt.Errorf("salary restricted")
	}
	return nil
}
