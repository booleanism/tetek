package recipes

import (
	"github.com/booleanism/tetek/account/amqp"
	"github.com/booleanism/tetek/auth/internal/contract"
	"github.com/booleanism/tetek/auth/internal/jwt"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type LoginRecipe interface {
	Login(amqp.User) (string, errro.Error)
}

type loginRecipe struct {
	l   *contract.LocalAccContr
	jwt jwt.JwtRecipes
}

func NewLogin(contr *contract.LocalAccContr, jwt jwt.JwtRecipes) *loginRecipe {
	return &loginRecipe{contr, jwt}
}

func (r *loginRecipe) Login(user amqp.User) (string, errro.Error) {
	if (user.Uname == "" && user.Email == "") || user.Passwd == "" {
		e := errro.New(errro.EAUTH_INVALID_LOGIN_PARAM, "either email or uname and passwd should non empty value")
		return "", e
	}

	id := uuid.NewString()
	task := amqp.AccountTask{
		Cmd: 0,
		User: amqp.User{
			Uname:  user.Uname,
			Email:  user.Email,
			Passwd: user.Passwd,
		},
	}

	if err := r.l.Publish(id, task); err != nil {
		e := errro.New(errro.EACCOUNT_SERVICE_UNAVAILABLE, "failed to publish account task")
		return "", e
	}

	res, err := r.l.Consume(id)
	if err != nil {
		e := errro.New(errro.EACCOUNT_SERVICE_UNAVAILABLE, "failed consuming account result")
		return "", e
	}

	if res.Code == errro.SUCCESS {
		if err := bcrypt.CompareHashAndPassword([]byte(res.Detail.Passwd), []byte(user.Passwd)); err != nil {
			e := errro.New(errro.EAUTH_INVALID_CREDS, "invalid credentials")
			return "", e
		}
		jwt, err := r.jwt.Generate(res.Detail)
		if err != nil {
			e := errro.New(errro.EAUTH_JWT_GENERATAION_FAIL, "failed to generate jwt")
			return "", e
		}

		return jwt, nil
	}

	if res.Code == errro.EACCOUNT_NO_USER {
		e := errro.New(errro.EACCOUNT_NO_USER, "user not found")
		return "", e
	}

	e := errro.New(errro.EAUTH_RETRIEVE_USER_FAIL, "failed to retrieve user information")
	return "", e
}
