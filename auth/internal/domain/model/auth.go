package model

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ExpJwt expired token in hours
const ExpJwt = 24 * time.Hour

type JwtClaims struct {
	Uname string `json:"uname"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}
