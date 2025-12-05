package usecases

import (
	"context"

	messaging "github.com/booleanism/tetek/account/infra/messaging/rabbitmq"
	"github.com/booleanism/tetek/auth/internal/usecases/dto"
	"github.com/booleanism/tetek/pkg/contracts/adapter"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/go-logr/logr"
	"golang.org/x/crypto/bcrypt"
)

type LoginUseCase interface {
	Login(context.Context, dto.LoginRequest) (string, errro.Error)
}

func (r usecases) Login(ctx context.Context, req dto.LoginRequest) (string, errro.Error) {
	ctx, log := loggr.GetLogger(ctx, "login-recipes")

	user := &messaging.User{}
	(*user).Uname = req.Uname
	(*user).Passwd = req.Passwd
	(*user).Email = req.Email

	if err := checkLoginProperty(log, &user); err != nil {
		return "", err
	}

	res := &messaging.AccountResult{}
	task := messaging.AccountTask{Cmd: 0, User: *user}
	if err := adapter.AccountAdapter(ctx, r.l, task, &res); err != nil {
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

func checkLoginProperty(log logr.Logger, user **messaging.User) errro.Error {
	if ((*user).Uname == "" && (*user).Email == "") || (*user).Passwd == "" {
		e := errro.New(errro.ErrAuthInvalidLoginParam, "either email or uname and passwd should non empty value")
		log.V(1).Info("missing user property", "error", e)
		return e
	}
	return nil
}
