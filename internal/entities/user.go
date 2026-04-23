package entities

import "github.com/google/uuid"

type User struct {
	ID    uuid.UUID `db:"u_id" json:"id"`
	Email string    `db:"email" json:"email"`
	Name  string    `db:"user_name" json:"name"`
}
