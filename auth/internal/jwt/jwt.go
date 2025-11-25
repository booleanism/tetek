package jwt

import (
	"context"
	"errors"
	"time"

	"github.com/booleanism/tetek/account/amqp"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/golang-jwt/jwt/v5"
)

type JwtRecipes interface {
	Verify(context.Context, string, **JwtClaims) error
	Generate(amqp.User) (string, error)
}

// ExpJwt expired token in hours
const ExpJwt = 24 * time.Hour

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

func (r *jwtRecipe) Verify(ctx context.Context, j string, claims **JwtClaims) error {
	corrID := ctx.Value(keystore.RequestID{}).(string)
	_, log := loggr.GetLogger(ctx, "verifier")
	token, err := jwt.ParseWithClaims(j, &JwtClaims{}, func(token *jwt.Token) (any, error) {
		return r.secrt, nil
	})
	if err != nil {
		log.V(2).Info("failed to parse JWT", "requestID", corrID, "error", err)
		return err
	}

	c, ok := token.Claims.(*JwtClaims)
	if !token.Valid || !ok {
		e := errors.New("invalid claims")
		log.V(2).Info("missmatch claims type", "requestID", corrID, "error", e, "claims", token.Claims)
		return e
	}

	if c.Subject != "auth" {
		e := errors.New("invalid subject claims")
		log.V(2).Info("expected auth subject", "requestID", corrID, "error", e, "subject", c.Subject)
		return e
	}

	if time.Now().After(c.ExpiresAt.Time) {
		e := errors.New("token expired")
		log.V(2).Info("expiration time exceeded", "requestID", corrID, "error", e, "exp", token.Claims, "now", time.Now())
		return e
	}

	(*claims) = c

	return nil
}

func (r *jwtRecipe) Generate(user amqp.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, JwtClaims{
		Uname: user.Uname,
		Role:  string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "auth",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ExpJwt)),
		},
	})

	return token.SignedString(r.secrt)
}
