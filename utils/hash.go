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

// CheckPasswordHash 는 저장된 bcrypt 해시와 평문 비밀번호가 일치하는지 확인한다.
//
// bcrypt.CompareHashAndPassword 는 첫 인자가 해시, 둘째가 평문이다.
// 비밀번호 불일치는 정상적인 경우이므로 error 없이 false 로 구분해 반환한다.
func CheckPasswordHash(hash, password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
