package recipes

import (
	"context"

	"github.com/booleanism/tetek/account/amqp"
	"github.com/booleanism/tetek/auth/internal/jwt"
	"github.com/booleanism/tetek/pkg/contracts"
	"github.com/booleanism/tetek/pkg/errro"
	"golang.org/x/crypto/bcrypt"
)

type LoginRecipe interface {
	Login(context.Context, amqp.User) (string, errro.Error)
}

type loginRecipe struct {
	l   contracts.AccountSubscribe
	jwt jwt.JwtRecipes
}

func NewLogin(contr contracts.AccountSubscribe, jwt jwt.JwtRecipes) *loginRecipe {
	return &loginRecipe{contr, jwt}
}

func (r *loginRecipe) Login(ctx context.Context, user amqp.User) (string, errro.Error) {
	if (user.Uname == "" && user.Email == "") || user.Passwd == "" {
		e := errro.New(errro.ErrAuthInvalidLoginParam, "either email or uname and passwd should non empty value")
		return "", e
	}

	task := amqp.AccountTask{
		Cmd: 0,
		User: amqp.User{
			Uname:  user.Uname,
			Email:  user.Email,
			Passwd: user.Passwd,
		},
	}

	if err := r.l.Publish(ctx, task); err != nil {
		e := errro.New(errro.ErrAccountServiceUnavailable, "failed to publish account task")
		return "", e
	}

	var res *amqp.AccountRes
	err := r.l.Consume(ctx, &res)
	if err != nil {
		e := errro.New(errro.ErrAccountServiceUnavailable, "failed consuming account result")
		return "", e
	}

	if res.Code == errro.Success {
		if err := bcrypt.CompareHashAndPassword([]byte(res.Detail.Passwd), []byte(user.Passwd)); err != nil {
			e := errro.New(errro.ErrAuthInvalidCreds, "invalid credentials")
			return "", e
		}
		jwt, err := r.jwt.Generate(res.Detail)
		if err != nil {
			e := errro.New(errro.ErrAuthJWTGenerationFail, "failed to generate jwt")
			return "", e
		}

		return jwt, nil
	}

	if res.Code == errro.ErrAccountNoUser {
		e := errro.New(errro.ErrAccountNoUser, "user not found")
		return "", e
	}

	e := errro.New(errro.ErrAuthFailRetrieveUser, "failed to retrieve user information")
	return "", e
}
