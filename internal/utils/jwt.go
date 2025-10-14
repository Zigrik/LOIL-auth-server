package utils

import (
	"time"
	
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID     int    `json:"userId"`
	Login      string `json:"login"`
	GameSurname string `json:"gameSurname"`
	jwt.RegisteredClaims
}

func GenerateJWT(userID int, login, gameSurname, secret string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	
	claims := &Claims{
		UserID:     userID,
		Login:      login,
		GameSurname: gameSurname,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateJWT(tokenString, secret string) (*Claims, error) {
	claims := &Claims{}
	
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	
	if err != nil {
		return nil, err
	}
	
	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	
	return claims, nil
}