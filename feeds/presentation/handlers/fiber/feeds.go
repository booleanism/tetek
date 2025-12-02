package router

import (
	"context"

	"github.com/booleanism/tetek/feeds/internal/model/pools"
	"github.com/booleanism/tetek/feeds/recipes"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/gofiber/fiber/v3"
)

type FeedsRouter struct {
	rec recipes.FeedRecipes
}

func NewFeedRouter(rec recipes.FeedRecipes) FeedsRouter {
	return FeedsRouter{rec}
}

func (fr FeedsRouter) GetFeeds(ctx fiber.Ctx) error {
	c, log := loggr.GetLogger(ctx.Context(), ctx.Route().Name)
	log.V(1).Info("new get feed request")

	req := recipes.GetFeedsRequest{}
	gRes := helper.GenericResponse{}
	if err := helper.BindRequest(ctx, &req); err != nil {
		return err.SendError(ctx, fiber.StatusBadRequest)
	}

	cto, cancel := context.WithTimeout(
		c,
		helper.Timeout)
	defer cancel()

	fBuf, ok := pools.FeedsPool.Get().(*pools.Feeds)
	if !ok {
		gRes.Code = errro.ErrAcqPool
		gRes.Message = "failed to acquire pool"
		e := errro.New(gRes.Code, gRes.Message)
		log.Error(e, e.Msg())
		return e.WithDetail(gRes.JSON(), errro.TDetailJSON).SendError(ctx, fiber.StatusInternalServerError)
	}
	defer pools.FeedsPool.Put(fBuf)
	defer fBuf.Reset()

	err := fr.rec.Feeds(cto, req, fBuf)
	if err == nil {
		gRes.Code = errro.Success
		gRes.Message = "fetch feeds success"
		res := recipes.GetFeedsResponse{GenericResponse: gRes, Details: fBuf.Value}
		return ctx.Status(fiber.StatusOK).JSON(&res)
	}

	if err.Code() == errro.ErrFeedsNoFeeds {
		gRes.Code = err.Code()
		gRes.Message = err.Error()
		return ctx.Status(fiber.StatusNotFound).JSON(&gRes)
	}

	gRes.Code = errro.ErrFeedsDBError
	gRes.Message = "fail to fetch feeds"
	return err.WithDetail(gRes.JSON(), errro.TDetailJSON).SendError(ctx, fiber.StatusInternalServerError)
}
