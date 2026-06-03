package entity

import "time"

type UserStatus string
type UserGender string
type UserRole string

const (
	StatusActive  UserStatus = "ACTIVE"
	StatusBanned  UserStatus = "BANNED"
	StatusDeleted UserStatus = "DELETED"

	GenderMale   UserGender = "MALE"
	GenderFemale UserGender = "FEMALE"
	GenderOther  UserGender = "OTHER"

	RoleTenant   UserRole = "TENANT"
	RoleLandlord UserRole = "LANDLORD"
	RoleAdmin    UserRole = "ADMIN"
)

type User struct {
	ID               string
	Email            string
	FullName         string
	PhoneNumber      string
	Gender           UserGender
	DateOfBirth      *time.Time
	AvatarURL        string
	AvatarPublicID   string
	ProfileCompleted bool
	Status           UserStatus
	Roles            []UserRole
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

func (u *User) HasRole(role UserRole) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

func (u *User) IsProfileComplete() bool {
	return u.FullName != "" &&
		u.PhoneNumber != "" &&
		u.Gender != "" &&
		u.DateOfBirth != nil
}
