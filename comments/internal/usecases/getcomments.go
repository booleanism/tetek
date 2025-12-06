package usecases

import (
	"context"
	"database/sql"

	"github.com/booleanism/tetek/comments/internal/internal/domain/model"
	"github.com/booleanism/tetek/comments/internal/usecases/dto"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/loggr"
)

type GetCommentsUseCase interface {
	GetComments(ctx context.Context, gcr dto.GetCommentsRequest, buf *[]dto.Comment) (n int, err errro.Error)
}

func (uc usecases) GetComments(ctx context.Context, gcr dto.GetCommentsRequest, buf *[]dto.Comment) (int, errro.Error) {
	ctx, log := loggr.GetLogger(ctx, "getComments-usecases")

	if gcr.Head.String() == "00000000-0000-0000-0000-000000000000" {
		e := errro.New(errro.ErrCommMissingRequiredField, "invalid Head ID")
		log.V(2).Info(e.Msg())
		return 0, e
	}

	lim := gcr.Limit
	if lim == 0 {
		lim = model.DefaultLimit
	}

	n, err := uc.repo.GetComments(ctx, gcr.Head, lim, gcr.Offset, buf)
	if err == sql.ErrNoRows {
		e := errro.FromError(errro.ErrCommNoComments, "no such comments", err)
		return 0, e
	}

	if err != nil {
		e := errro.FromError(errro.ErrCommDBError, "failed to fetch comments", err)
		return 0, e
	}

	return n, nil
}
