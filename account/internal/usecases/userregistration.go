package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/booleanism/tetek/account/internal/domain/entities"
	"github.com/booleanism/tetek/account/internal/domain/model"
	"github.com/booleanism/tetek/account/internal/usecases/dto"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/loggr"
)

type UserRegistrationUseCase interface {
	RegistUser(ctx context.Context, rur dto.RegistRequest) errro.Error
}

func (uc usecases) RegistUser(ctx context.Context, rur dto.RegistRequest) errro.Error {
	ctx, log := loggr.GetLogger(ctx, "registUser-usecases")

	if rur.Uname == "" || rur.Email == "" || rur.Passwd == "" || rur.ID.String() == "00000000-0000-0000-0000-000000000000" {
		e := errro.New(errro.ErrAccountInvalidRegistParam, "missing required field")
		log.V(2).Info(e.Msg())
		return e
	}

	if len(rur.Passwd) < model.MinPasswdLength {
		e := errro.New(errro.ErrAccountInvalidRegistParam, fmt.Sprintf("insufficient passwd length, minimal: %d", model.MinPasswdLength))
		log.V(2).Info(e.Msg())
		return e
	}

	now := time.Now()
	u := &entities.User{
		ID:        rur.ID,
		Uname:     rur.Uname,
		Email:     rur.Email,
		Passwd:    rur.Passwd,
		Role:      "N",
		CreatedAt: &now,
	}
	err := uc.repo.PutUser(ctx, &u)
	if err != nil {
		e := errro.FromError(errro.ErrAccountRegistFail, "registration failed", err)
		log.Error(err, e.Msg())
		return e
	}

	return nil
}
