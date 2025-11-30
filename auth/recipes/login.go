package recipes

import (
	"context"

	"github.com/booleanism/tetek/account/amqp"
	"github.com/booleanism/tetek/auth/internal/jwt"
	"github.com/booleanism/tetek/pkg/contracts"
	"github.com/booleanism/tetek/pkg/contracts/adapter"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/go-logr/logr"
	"golang.org/x/crypto/bcrypt"
)

type LoginRecipe interface {
	Login(context.Context, LoginRequest) (string, errro.Error)
}

type loginRecipe struct {
	l   contracts.AccountDealer
	jwt jwt.JwtRecipes
}

func NewLogin(contr contracts.AccountDealer, jwt jwt.JwtRecipes) *loginRecipe {
	return &loginRecipe{contr, jwt}
}

func (r *loginRecipe) Login(ctx context.Context, req LoginRequest) (string, errro.Error) {
	ctx, log := loggr.GetLogger(ctx, "login-recipes")

	user := &amqp.User{}
	req.toUser(&user)
	if err := checkLoginProperty(log, &user); err != nil {
		return "", err
	}

	res := &amqp.AccountResult{}
	task := amqp.AccountTask{Cmd: 0, User: *user}
	if err := adapter.AccAdapter(ctx, r.l, task, &res); err != nil {
		return "", err
	}

	if res.Code == errro.Success {
		if err := bcrypt.CompareHashAndPassword([]byte(res.Detail.Passwd), []byte(req.Passwd)); err != nil {
			e := errro.FromError(errro.ErrAuthInvalidCreds, "invalid credentials", err)
			log.V(1).Info(e.Msg())
			return "", e
		}
		jwt, err := r.jwt.Generate(res.Detail)
		if err != nil {
			e := errro.FromError(errro.ErrAuthJWTGenerationFail, "failed to generate jwt", err)
			log.Error(err, e.Msg())
			return "", e
		}

		return jwt, nil
	}

	if res.Code == errro.ErrAccountNoUser {
		e := errro.New(errro.ErrAccountNoUser, "user not found")
		log.V(2).Info(e.Msg(), "user", user)
		return "", e
	}

	e := errro.New(errro.ErrAuthFailRetrieveUser, "failed to retrieve user information")
	log.V(2).Info(e.Msg(), "user", user)
	return "", e
}

func checkLoginProperty(log logr.Logger, user **amqp.User) errro.Error {
	if ((*user).Uname == "" && (*user).Email == "") || (*user).Passwd == "" {
		e := errro.New(errro.ErrAuthInvalidLoginParam, "either email or uname and passwd should non empty value")
		log.V(1).Info("missing user property", "error", e)
		return e
	}
	return nil
}
