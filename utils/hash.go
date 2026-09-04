package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword hash password by given password.
//
// It returns string, and nil error when successful.
// Otherwise, empty string, and error will be returned.
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}

// CheckPasswordHash check password hash by given password.
//
// It returns bool, and nil error when successful.
// Otherwise, empty bool, and error will be returned.
func CheckPasswordHash(password, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(password), []byte(hash))
	if err != nil {
		return false, err
	}

	return true, nil
}
