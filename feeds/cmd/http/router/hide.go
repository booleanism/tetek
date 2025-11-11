package router

import (
	"context"
	"time"

	"github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/feeds/recipes"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/gofiber/fiber/v3"
)

func (fr FeedsRouter) HideFeed(ctx fiber.Ctx) error {
	req := recipes.HideRequest{}
	gRes := helper.GenericResponse{}
	if err := helper.BindRequest(ctx, &req); err != nil {
		return err.SendError(ctx, fiber.StatusBadRequest)
	}

	_, ok := ctx.Context().Value(keystore.AuthRes{}).(*amqp.AuthResult)
	if !ok {
		gRes.Code = errro.EAUTH_INVALID_AUTH_RESULT_TYPE
		gRes.Message = "does not represent jwt type"
		e := errro.New(gRes.Code, gRes.Message).WithDetail(gRes.Json(), errro.TDETAIL_JSON)
		return e.SendError(ctx, fiber.StatusBadRequest)
	}

	cto, cancel := context.WithTimeout(
		ctx.Context(),
		TIMEOUT*time.Second)
	defer cancel()

	err := fr.rec.Hide(cto, req)
	if err != nil {
		gRes.Code = err.Code()
		gRes.Message = err.Error()
		return errro.New(gRes.Code, gRes.Message).WithDetail(gRes.Json(), errro.TDETAIL_JSON).SendError(ctx, fiber.StatusInternalServerError)
	}

	gRes.Code = errro.SUCCESS
	gRes.Message = "feed hidden"
	res := recipes.HideResponse{GenericResponse: gRes, Detail: req}
	return ctx.Status(fiber.StatusOK).JSON(&res)
}
