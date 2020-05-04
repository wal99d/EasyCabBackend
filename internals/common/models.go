package common

import "github.com/dgrijalva/jwt-go"

type Claims struct {
	Mobile string `json:"mobile"`
	jwt.StandardClaims
}
