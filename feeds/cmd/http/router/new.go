package router

import (
	"context"
	"time"

	"github.com/booleanism/tetek/feeds/recipes"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/gofiber/fiber/v3"
)

func (fr FeedsRouter) NewFeed(ctx fiber.Ctx) error {
	req := recipes.NewFeedRequest{}
	gRes := helper.GenericResponse{}
	if err := helper.BindRequest(ctx, &req); err != nil {
		return err.SendError(ctx, fiber.StatusBadRequest)
	}

	cto, cancel := context.WithTimeout(
		ctx.Context(),
		TIMEOUT*time.Second)
	defer cancel()

	err := fr.rec.New(cto, req)
	if err == nil {
		gRes.Code = errro.SUCCESS
		gRes.Message = "success add new feed"
		res := recipes.NewFeedResponse{
			GenericResponse: gRes,
			Detail:          req,
		}
		return ctx.Status(fiber.StatusCreated).JSON(&res)
	}

	gRes.Code = errro.EFEEDS_NEW_FAIL
	gRes.Message = "failed to create new feed"
	res := recipes.NewFeedResponse{
		GenericResponse: gRes,
		Detail:          req,
	}
	e := errro.New(res.Code, res.Message).WithDetail(res.Json(), errro.TDETAIL_JSON)
	return e.SendError(ctx, fiber.StatusInternalServerError)
}
