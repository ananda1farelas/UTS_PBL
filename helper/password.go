package helper

import "golang.org/x/crypto/bcrypt"

// HashPassword hashes a plain-text password using bcrypt.
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// CheckPassword compares a bcrypt hash against a plain-text candidate.
// It returns true when they match.
func CheckPassword(hashedPassword, candidate string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(candidate))
	return err == nil
}
