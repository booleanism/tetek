package router

import (
	"context"

	"github.com/booleanism/tetek/feeds/recipes"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/gofiber/fiber/v3"
)

func (fr FeedsRouter) NewFeed(ctx fiber.Ctx) error {
	c, log := loggr.GetLogger(ctx.Context(), ctx.Route().Name)
	log.V(1).Info("new new feed request")

	req := recipes.NewFeedRequest{}
	gRes := helper.GenericResponse{}
	if err := helper.BindRequest(ctx, &req); err != nil {
		return err.SendError(ctx, fiber.StatusBadRequest)
	}

	cto, cancel := context.WithTimeout(
		c,
		helper.Timeout)
	defer cancel()

	err := fr.rec.New(cto, req)
	if err == nil {
		gRes.Code = errro.Success
		gRes.Message = "success add new feed"
		res := recipes.NewFeedResponse{
			GenericResponse: gRes,
			Detail:          req,
		}
		return ctx.Status(fiber.StatusCreated).JSON(&res)
	}

	gRes.Code = errro.ErrFeedsNewFail
	gRes.Message = "failed to create new feed"
	res := recipes.NewFeedResponse{
		GenericResponse: gRes,
		Detail:          req,
	}
	e := errro.New(res.Code, res.Message).WithDetail(res.JSON(), errro.TDetailJSON)
	return e.SendError(ctx, fiber.StatusInternalServerError)
}
