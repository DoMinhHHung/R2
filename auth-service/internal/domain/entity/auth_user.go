package entity

import "time"

type Role string

const (
	RoleAdmin    Role = "ADMIN"
	RoleLandlord Role = "LANDLORD"
	RoleTenant   Role = "TENANT"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleLandlord, RoleTenant:
		return true
	}
	return false
}

type AuthUser struct {
	ID           string
	Email        string
	PasswordHash string
	Role         Role
	IsVerified   bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
