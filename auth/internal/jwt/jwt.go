package jwt

import (
	"errors"
	"time"

	// "github.com/booleanism/tetek/account/contract"
	"github.com/booleanism/tetek/account/amqp"
	"github.com/golang-jwt/jwt/v5"
)

type JwtRecipes interface {
	Verify(string) (*JwtClaims, error)
	Generate(amqp.User) (string, error)
}

// in hours
const EXP_JWT = 24

type JwtClaims struct {
	Uname string `json:"uname"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

type jwtRecipe struct {
	secrt []byte
}

func NewJwt(secr []byte) *jwtRecipe {
	return &jwtRecipe{secr}
}

func (r *jwtRecipe) Verify(j string) (*JwtClaims, error) {
	token, err := jwt.ParseWithClaims(j, &JwtClaims{}, func(token *jwt.Token) (any, error) {
		return r.secrt, nil
	})
	if err != nil {
		return nil, errors.New("fail to parse JWT")
	}

	c, ok := token.Claims.(*JwtClaims)
	if !token.Valid || !ok {
		return nil, errors.New("invalid token")
	}

	if c.Subject != "auth" {
		return nil, errors.New("invalid subject")
	}

	if time.Now().After(c.ExpiresAt.Time) {
		return nil, errors.New("token expired")
	}

	return c, nil
}

func (r *jwtRecipe) Generate(user amqp.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, JwtClaims{
		Uname: user.Uname,
		Role:  string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "auth",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(EXP_JWT * time.Hour)),
		},
	})

	return token.SignedString(r.secrt)
}
