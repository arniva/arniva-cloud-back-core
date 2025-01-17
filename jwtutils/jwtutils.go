package jwtutils

import (
	"github.com/golang-jwt/jwt"
)

// Claims JWT token içeriği
type Claims struct {
	UserID    string `json:"user_id"`
	CompanyID string `json:"company_id"`
	jwt.StandardClaims
}

func ParseClaims(token string) (*Claims, error) {
	var claims Claims
	_, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})
	if err != nil {
		return nil, err
	}
	return &claims, nil
}
