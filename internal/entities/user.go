package entities

import "github.com/google/uuid"

type User struct {
	ID              uuid.UUID `db:"u_id" json:"id"`
	Email           string    `db:"email" json:"email"`
	Name            string    `db:"user_name" json:"name"`
	DefaultCurrency string    `db:"default_currency" json:"default_currency"`
}

type AuthUser struct {
	ID              string `json:"id"`
	Email           string `json:"email"`
	Name            string `json:"name"`
	DefaultCurrency string `json:"default_currency"`
}

type AuthSession struct {
	Token string   `json:"token"`
	User  AuthUser `json:"user"`
}
