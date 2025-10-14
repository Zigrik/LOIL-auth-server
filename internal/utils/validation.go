package utils

import (
	"regexp"
	"strings"
)

// ValidateLogin проверяет формат логина: буквы, цифры, подчеркивание
func ValidateLogin(login string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]{3,20}$`, login)
	return matched
}

// ValidateGameSurname проверяет игровую фамилию: только латинские буквы
func ValidateGameSurname(surname string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z]{2,20}$`, surname)
	return matched
}

// NormalizeGameSurname нормализует фамилию: Первая заглавная, остальные строчные
func NormalizeGameSurname(surname string) string {
	if len(surname) == 0 {
		return surname
	}

	return strings.ToUpper(surname[:1]) + strings.ToLower(surname[1:])
}

// ValidateEmail проверяет формат email
func ValidateEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(emailRegex, email)
	return matched
}
