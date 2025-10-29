package recipes

import (
	"github.com/booleanism/tetek/account/amqp"
	"github.com/booleanism/tetek/auth/internal/contract"
	"github.com/booleanism/tetek/auth/internal/jwt"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/go-logr/logr"
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
		return "", loggr.Log.Error(3, func(z logr.LogSink) errro.Error {
			e := errro.New(errro.EAUTH_INVALID_LOGIN_PARAM, "either email or uname and passwd should non empty value")
			z.Error(e, "missing required user login field", "user", user)
			return e
		})
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
		return "", loggr.Log.Error(0, func(z logr.LogSink) errro.Error {
			e := errro.New(errro.EACCOUNT_SERVICE_UNAVAILABLE, "failed to publish account task")
			z.Error(err, e.Error(), "id", id, "task", task)
			return e
		})
	}

	res, err := r.l.Consume(id)
	if err != nil {
		return "", loggr.Log.Error(0, func(z logr.LogSink) errro.Error {
			e := errro.New(errro.EACCOUNT_SERVICE_UNAVAILABLE, "failed consuming account result")
			z.Error(err, e.Error(), "id", id, "task", task)
			return e
		})
	}

	if res.Code == errro.SUCCESS {
		if err := bcrypt.CompareHashAndPassword([]byte(res.Detail.Passwd), []byte(user.Passwd)); err != nil {
			return "", loggr.Log.Error(4, func(z logr.LogSink) errro.Error {
				e := errro.New(errro.EAUTH_INVALID_CREDS, "invalid credentials")
				z.Error(e, "hashed passwd does not match", "passwd", user.Passwd, "actual", res.Detail.Passwd)
				return e
			})
		}
		jwt, err := r.jwt.Generate(res.Detail)
		if err != nil {
			return "", loggr.Log.Error(0, func(z logr.LogSink) errro.Error {
				e := errro.New(errro.EAUTH_JWT_GENERATAION_FAIL, "failed to generate jwt")
				z.Error(err, e.Error(), "user", user)
				return e
			})
		}

		return jwt, nil
	}

	if res.Code == errro.EACCOUNT_NO_USER {
		return "", loggr.Log.Error(4, func(z logr.LogSink) errro.Error {
			e := errro.New(errro.EACCOUNT_NO_USER, "user not found")
			z.Error(e, "could not find user", "user", user)
			return e
		})
	}

	return "", loggr.Log.Error(1, func(z logr.LogSink) errro.Error {
		e := errro.New(errro.EAUTH_RETRIEVE_USER_FAIL, "failed to retrieve user information")
		z.Error(e, "something happen", "user", user)
		return e
	})
}
