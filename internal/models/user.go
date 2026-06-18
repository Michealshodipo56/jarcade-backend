package models

import (
	"strings"
	"time"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
}

type PublicUser struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Initials  string    `json:"initials"`
	Display   string    `json:"display"`
	CreatedAt time.Time `json:"createdAt"`
}

func ToPublicUser(u User) PublicUser {
	local := u.Email
	if at := strings.Index(u.Email, "@"); at > 0 {
		local = u.Email[:at]
	}

	initials := "JA"
	switch {
	case len(local) >= 2:
		initials = strings.ToUpper(local[:2])
	case len(local) == 1:
		initials = strings.ToUpper(local) + "U"
	}

	return PublicUser{
		ID:        u.ID,
		Email:     u.Email,
		Initials:  initials,
		Display:   local,
		CreatedAt: u.CreatedAt,
	}
}
