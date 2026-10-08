package model

import "time"

const (
	RoleAdmin     = "admin"
	RoleMahasiswa = "mahasiswa"
)

// User represents a row in the users table.
type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
