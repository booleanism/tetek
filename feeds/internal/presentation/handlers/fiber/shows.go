package handlers

import (
	"context"

	"github.com/booleanism/tetek/feeds/internal/usecases/dto"
	"github.com/booleanism/tetek/pkg/contracts/adapter"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/gofiber/fiber/v3"
)

func (fr FeedsRouter) ShowFeed(ctx fiber.Ctx) error {
	c, log := loggr.GetLogger(ctx.Context(), ctx.Route().Name)
	log.V(1).Info("new show feed request")

	req := dto.ShowFeedRequest{}
	res := dto.ShowFeedResponse{}
	if err := helper.BindRequest(ctx, &req); err != nil {
		return err.SendError(ctx, fiber.StatusBadRequest)
	}

	cto, cancel := context.WithTimeout(
		c,
		helper.Timeout)
	defer cancel()

	r := &res.Detail
	err := fr.rec.ShowFeed(cto, fr.commDealer, adapter.CommentsAdapter, req, &r)
	if err == nil {
		res.GenericResponse = helper.GenericResponse{Code: errro.Success, Message: "success show feed"}
		return ctx.Status(fiber.StatusOK).JSON(&res)
	}

	code := err.Code()
	if code == errro.ErrFeedsNoFeeds {
		res.GenericResponse = helper.GenericResponse{Code: code, Message: "no such feed"}
		return ctx.Status(fiber.StatusNotFound).JSON(&res)
	}

	if code == errro.ErrFeedsHidden {
		res.GenericResponse = helper.GenericResponse{Code: code, Message: "no such feed"}
		return ctx.Status(fiber.StatusBadRequest).JSON(&res)
	}

	res.GenericResponse = helper.GenericResponse{Code: code, Message: err.Msg()}
	return ctx.Status(fiber.StatusInternalServerError).JSON(&res)
}
