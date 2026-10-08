package helper

import "errors"

// ValidatePassword enforces the minimum password rule from the spec:
// at least 8 characters.
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password minimal 8 karakter")
	}
	return nil
}
