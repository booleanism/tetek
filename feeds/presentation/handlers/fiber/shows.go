package router

import (
	"github.com/booleanism/tetek/feeds/recipes"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/gofiber/fiber/v3"
)

func (fr FeedsRouter) ShowFeed(ctx fiber.Ctx) error {
	_, log := loggr.GetLogger(ctx.Context(), ctx.Route().Name)
	log.V(1).Info("new show feed request")

	req := recipes.ShowFeedRequest{}
	res := recipes.ShowFeedResponse{}
	if err := helper.BindRequest(ctx, &req); err != nil {
		return err.SendError(ctx, fiber.StatusBadRequest)
	}

	r := &res.Detail
	err := fr.rec.Shows(ctx.Context(), req, &r)
	if err == nil {
		res.GenericResponse = helper.GenericResponse{Code: errro.Success, Message: "success show feed"}
		return ctx.Status(fiber.StatusOK).JSON(&res)
	}

	code := err.Code()
	if code == errro.ErrFeedsNoFeeds {
		res.GenericResponse = helper.GenericResponse{Code: code, Message: "no such feed"}
		return ctx.Status(fiber.StatusNotFound).JSON(&res)
	}

	res.GenericResponse = helper.GenericResponse{Code: code, Message: err.Msg()}
	return ctx.Status(fiber.StatusInternalServerError).JSON(&res)
}
