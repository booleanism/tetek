package usecases

import (
	"context"
	"strings"

	"github.com/booleanism/tetek/feeds/internal/domain/entities"
	"github.com/booleanism/tetek/feeds/internal/domain/model"
	"github.com/booleanism/tetek/feeds/internal/domain/model/pools"
	"github.com/booleanism/tetek/feeds/internal/usecases/dto"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/loggr"
)

type GetFeedUseCase interface {
	// GetFeeds [repo.FeedsRepo]'s FeedsEntityRepo wrapper with some logic behind.
	// Return how many rows inside and nil'd error.
	// If error occur returning either [errro.ErrFeedsMissingRequiredField] [errro.ErrFeedsNoFeeds] or [errro.ErrFeedsDBError]
	GetFeeds(ctx context.Context, gfr dto.GetFeedsRequest, fp **pools.Feeds) (n int, err errro.Error)
}

// If gfr.ID provided it's mutate fp with a feeds corresponding ID and ignores all [GetFeedsRequest] fields.
// If gfr.Uname was given then it will returning visible (not hidden) feeds with corresponding gfr,Type.
// Lastly, if above condition does not meets. The routine tooks gfr.Type and returns feeds corresponding type.
// If gfr.Type not provided, return [errro.ErrFeedsMissingRequiredField].
func (uc usecases) GetFeeds(ctx context.Context, gfr dto.GetFeedsRequest, fp **pools.Feeds) (int, errro.Error) {
	ctx, _ = loggr.GetLogger(ctx, "getFeeds-usecases")

	if gfr.ID.String() != "00000000-0000-0000-0000-000000000000" {
		buf := &entities.Feed{}
		err := uc.getFeedsByIDExec(ctx, gfr.ID, &buf)
		if err != nil {
			return 0, err
		}
		(*fp).Value = append((*fp).Value, *buf)
		return 1, nil
	}

	if gfr.Type == "" {
		e := errro.New(errro.ErrFeedsMissingRequiredField, "missing feeds type")
		return 0, e
	}

	// set default rows limit
	limit := gfr.Limit
	if limit == 0 {
		limit = model.DefaultLimit
	}

	fType := model.GetFeedsByType{
		Offset: gfr.Offset,
		Type:   strings.ToUpper(gfr.Type),
		Limit:  limit,
	}

	if gfr.Uname != "" {
		return uc.getFeedsExec(func() (int, error) {
			return uc.repo.GetFeedsNotHiddenBy(ctx, fType, gfr.Uname, &(*fp).Value)
		})
	}

	return uc.getFeedsExec(func() (int, error) {
		return uc.repo.GetFeedsByType(ctx, fType, &(*fp).Value)
	})
}
