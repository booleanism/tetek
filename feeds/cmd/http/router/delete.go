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

func (fr FeedsRouter) DeleteFeed(ctx fiber.Ctx) error {
	req := recipes.DeleteRequest{}
	gRes := helper.GenericResponse{}
	if err := helper.BindRequest(ctx, &req); err != nil {
		return err.SendError(ctx, fiber.StatusBadRequest)
	}

	jwt, ok := ctx.Context().Value(keystore.AuthRes{}).(*amqp.AuthResult)
	if !ok {
		gRes.Code = errro.EAUTH_INVALID_AUTH_RESULT_TYPE
		gRes.Message = "missing jwt"
		e := errro.New(gRes.Code, "does not represent jwt type").WithDetail(gRes.Json(), errro.TDETAIL_JSON)
		return e.SendError(ctx, fiber.StatusUnauthorized)
	}

	cto, cancel := context.WithTimeout(ctx.Context(), TIMEOUT*time.Second)
	defer cancel()

	err := fr.rec.Delete(cto, req)
	if err == nil {
		gRes.Code = errro.SUCCESS
		gRes.Message = "feed deleted"
		res := recipes.DeleteResponse{GenericResponse: gRes, Detail: req}
		return ctx.Status(fiber.StatusAccepted).JSON(&res)
	}

	if err.Code() == errro.EFEEDS_NO_FEEDS && jwt.Claims.Uname != "" {
		gRes.Code = errro.EFEEDS_DELETE_FAIL
		gRes.Message = "unauthorized user to performe this action"
		return ctx.Status(fiber.StatusUnauthorized).JSON(&gRes)
	}

	if err.Code() == errro.EFEEDS_NO_FEEDS {
		gRes.Code = errro.EFEEDS_DELETE_FAIL
		gRes.Message = err.Error()
		return ctx.Status(fiber.StatusInternalServerError).JSON(&gRes)
	}

	gRes.Code = err.Code()
	gRes.Message = err.Error()
	return ctx.Status(fiber.StatusNoContent).JSON(&gRes)
}
