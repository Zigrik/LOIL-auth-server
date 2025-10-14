package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword создает bcrypt хэш пароля
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash проверяет пароль против хэша
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePassword проверяет минимальные требования к паролю
func ValidatePassword(password string) bool {
	return len(password) >= 6
}
