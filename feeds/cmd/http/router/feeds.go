package router

import (
	"context"
	"time"

	"github.com/booleanism/tetek/feeds/recipes"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/gofiber/fiber/v3"
)

const TIMEOUT = 10

type FeedsRouter struct {
	rec recipes.FeedRecipes
}

func NewFeedRouter(rec recipes.FeedRecipes) FeedsRouter {
	return FeedsRouter{rec}
}

func (fr FeedsRouter) GetFeeds(ctx fiber.Ctx) error {
	req := recipes.GetFeedsRequest{}
	gRes := helper.GenericResponse{}
	if err := helper.BindRequest(ctx, &req); err != nil {
		return err.SendError(ctx, fiber.StatusBadRequest)
	}

	cto, cancel := context.WithTimeout(
		ctx.Context(),
		TIMEOUT*time.Second)
	defer cancel()

	f, err := fr.rec.Feeds(cto, req)
	if err == nil {
		gRes.Code = errro.SUCCESS
		gRes.Message = "fetch feeds success"
		res := recipes.GetFeedsResponse{
			GenericResponse: gRes,
			Detail:          f,
		}
		return ctx.Status(fiber.StatusOK).JSON(&res)
	}

	if err.Code() == errro.EFEEDS_NO_FEEDS {
		gRes.Code = err.Code()
		gRes.Message = err.Error()
		return ctx.Status(fiber.StatusNotFound).JSON(&gRes)
	}

	gRes.Code = errro.EFEEDS_DB_ERR
	gRes.Message = "fail to fetch feeds"
	return err.WithDetail(gRes.Json(), errro.TDETAIL_JSON).SendError(ctx, fiber.StatusInternalServerError)
}
