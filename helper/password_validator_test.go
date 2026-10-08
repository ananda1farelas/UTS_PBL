package helper

import "testing"

func TestValidatePassword(t *testing.T) {
	cases := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"too short", "abc123", true},
		{"exactly 8 chars", "abcd1234", false},
		{"long password", "superSecret123!", false},
		{"empty", "", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePassword(tc.password)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidatePassword(%q) error = %v, wantErr %v", tc.password, err, tc.wantErr)
			}
		})
	}
}

func TestHashAndCheckPassword(t *testing.T) {
	plain := "mySecret123"

	hashed, err := HashPassword(plain)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hashed == plain {
		t.Fatal("HashPassword did not hash the password")
	}

	if !CheckPassword(hashed, plain) {
		t.Fatal("CheckPassword should return true for the correct password")
	}
	if CheckPassword(hashed, "wrong-password") {
		t.Fatal("CheckPassword should return false for an incorrect password")
	}
}
