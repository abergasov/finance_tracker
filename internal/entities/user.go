package entities

import "github.com/google/uuid"

type User struct {
	ID    uuid.UUID `db:"u_id" json:"id"`
	Email string    `db:"email" json:"email"`
	Name  string    `db:"user_name" json:"name"`
}

type AuthUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type AuthSession struct {
	Token string   `json:"token"`
	User  AuthUser `json:"user"`
}
