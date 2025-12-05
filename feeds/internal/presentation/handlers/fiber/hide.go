package handlers

import (
	"context"

	messaging "github.com/booleanism/tetek/auth/infra/messaging/rabbitmq"
	"github.com/booleanism/tetek/feeds/internal/usecases/dto"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/gofiber/fiber/v3"
)

func (fr FeedsRouter) HideFeed(ctx fiber.Ctx) error {
	c, log := loggr.GetLogger(ctx.Context(), ctx.Route().Name)
	log.V(1).Info("new hide feed request")

	req := dto.HideFeedRequest{}
	gRes := helper.GenericResponse{}
	if err := helper.BindRequest(ctx, &req); err != nil {
		return err.SendError(ctx, fiber.StatusBadRequest)
	}

	_, ok := c.Value(keystore.AuthRes{}).(*messaging.AuthResult)
	if !ok {
		gRes.Code = errro.ErrAuthInvalidType
		gRes.Message = "does not represent jwt type"
		e := errro.New(gRes.Code, gRes.Message).WithDetail(gRes.JSON(), errro.TDetailJSON)
		return e.SendError(ctx, fiber.StatusBadRequest)
	}

	cto, cancel := context.WithTimeout(
		c,
		helper.Timeout)
	defer cancel()

	err := fr.rec.HideFeed(cto, req)
	if err != nil {
		gRes.Code = err.Code()
		gRes.Message = err.Error()
		return errro.New(gRes.Code, gRes.Message).WithDetail(gRes.JSON(), errro.TDetailJSON).SendError(ctx, fiber.StatusInternalServerError)
	}

	gRes.Code = errro.Success
	gRes.Message = "feed hidden"
	res := dto.HideResponse{GenericResponse: gRes, Detail: req}
	return ctx.Status(fiber.StatusOK).JSON(&res)
}
