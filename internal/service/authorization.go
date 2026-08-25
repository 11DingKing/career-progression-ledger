package service

import (
	"careerprogression/internal/domain"
	"fmt"
)

type Permission string

const (
	ReadProfile    Permission = "read_profile"
	WriteProfile   Permission = "write_profile"
	VerifyEmployer Permission = "verify_employer"
	FreezeStats    Permission = "freeze_stats"
	ResolveAppeal  Permission = "resolve_appeal"
)

func Allowed(role domain.Role, p Permission) bool {
	switch role {
	case domain.RoleGraduate:
		return p == ReadProfile || p == WriteProfile
	case domain.RoleEmployer:
		return p == ReadProfile
	case domain.RoleCounselor:
		return p == ReadProfile || p == WriteProfile || p == VerifyEmployer
	case domain.RoleMajorAdmin:
		return true
	}
	return false
}
func Require(role domain.Role, p Permission) error {
	if !Allowed(role, p) {
		return fmt.Errorf("permission %s denied for %s", p, role)
	}
	return nil
}
