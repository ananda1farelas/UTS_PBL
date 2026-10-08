package helper

import "github.com/ananda1farelas/latihan-fiber/UTS/app/model"

// IsAdmin reports whether the role string is the admin role.
func IsAdmin(role string) bool {
	return role == model.RoleAdmin
}

// IsMahasiswa reports whether the role string is the mahasiswa role.
func IsMahasiswa(role string) bool {
	return role == model.RoleMahasiswa
}
