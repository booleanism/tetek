package usecases

import (
	"context"
	"database/sql"

	"github.com/booleanism/tetek/account/internal/usecases/dto"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/loggr"
)

type GetUserProfileUseCase interface {
	GetProfile(ctx context.Context, gpr dto.ProfileRequest, user **dto.User) errro.Error
}

func (uc usecases) GetProfile(ctx context.Context, gpr dto.ProfileRequest, user **dto.User) errro.Error {
	ctx, log := loggr.GetLogger(ctx, "getProfile-usecases")

	if gpr.Uname == "" {
		e := errro.New(errro.ErrAccountEmptyParam, "Uname should not empty")
		log.V(2).Info(e.Msg())
		return e
	}

	_, err := uc.repo.GetUser(ctx, gpr.Uname, user)
	if err == sql.ErrNoRows {
		e := errro.FromError(errro.ErrAccountNoUser, "no user corresponding Uname", err)
		return e
	}

	if err != nil {
		e := errro.FromError(errro.ErrAccountDBError, "failed to fetch user information", err)
		return e
	}

	return nil
}
