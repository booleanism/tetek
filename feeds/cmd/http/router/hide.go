package router

import (
	"context"
	"time"

	"github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/feeds/cmd/http/middleware"
	"github.com/booleanism/tetek/feeds/internal/repo"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type hideRequest struct {
	Id uuid.UUID `json:"id"`
}

type hideResponse struct {
	helper.GenericResponse
	Detail hideRequest `json:"detail"`
}

func (fr FeedsRouter) HideFeed(ctx fiber.Ctx) error {
	req := hideRequest{}
	if err := helper.BindRequest(ctx, &req); err != nil {
		return err.SendError(ctx, fiber.StatusBadRequest)
	}

	jwt, ok := ctx.Locals(middleware.AuthValueKey{}).(*amqp.AuthResult)
	if !ok {
		res := helper.GenericResponse{
			Code:    errro.EAUTH_INVALID_AUTH_RESULT_TYPE,
			Message: "does not represent jwt type",
		}
		e := errro.New(res.Code, res.Message).WithDetail(res.Json(), errro.TDETAIL_JSON)
		return e.SendError(ctx, fiber.StatusBadRequest)
	}

	ff := repo.FeedsFilter{
		Id:       req.Id,
		HiddenTo: jwt.Claims.Uname,
	}

	cto, cancel := context.WithTimeout(
		context.WithValue(
			context.Background(),
			helper.RequestIdKey{},
			ctx.Locals(helper.RequestIdKey{})),
		TIMEOUT*time.Second)
	defer cancel()

	err := fr.rec.Hide(cto, ff, jwt)
	if err != nil {
		res := helper.GenericResponse{
			Code:    err.Code(),
			Message: err.Error(),
		}
		return errro.New(res.Code, res.Message).WithDetail(res.Json(), errro.TDETAIL_JSON).SendError(ctx, fiber.StatusInternalServerError)
	}

	res := hideResponse{
		GenericResponse: helper.GenericResponse{
			Code:    errro.SUCCESS,
			Message: "feed hidden",
		},
		Detail: req,
	}
	return ctx.Status(fiber.StatusOK).JSON(&res)
}
