package router

import (
	"context"
	"strings"
	"time"

	"github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/feeds/internal/repo"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
)

type deleteRequest struct {
	Id uuid.UUID `uri:"id"`
}

type deleteResponse struct {
	helper.GenericResponse
	Detail deleteRequest `json:"detail"`
}

func (fr FeedsRouter) DeleteFeed(ctx fiber.Ctx, id types.UUID) error {
	req := deleteRequest{}
	if err := helper.BindRequest(ctx, &req); err != nil {
		return err.SendError(ctx, fiber.StatusBadRequest)
	}

	j := ctx.Locals("jwt")
	jwt, ok := j.(*amqp.AuthResult)
	if !ok {
		res := helper.GenericResponse{
			Code:    errro.EAUTH_INVALID_AUTH_RESULT_TYPE,
			Message: "missing jwt",
		}
		e := errro.New(res.Code, "does not represent jwt type").WithDetail(res.Json(), errro.TDETAIL_JSON)
		return e.SendError(ctx, fiber.StatusUnauthorized)
	}

	ff := repo.FeedsFilter{
		Id: req.Id,
	}

	// only moderator freely to delete feed
	if strings.ToLower(jwt.Claims.Role) != "m" {
		ff.By = jwt.Claims.Uname
	}

	cto, cancel := context.WithTimeout(
		context.WithValue(
			context.Background(),
			helper.RequestIdKey{},
			ctx.Locals(helper.RequestIdKey{})),
		TIMEOUT*time.Second)
	defer cancel()

	err := fr.rec.Delete(cto, ff, jwt)
	if err == nil {
		res := deleteResponse{
			GenericResponse: helper.GenericResponse{
				Code:    errro.SUCCESS,
				Message: "feed deleted",
			},
			Detail: req,
		}
		return ctx.Status(fiber.StatusAccepted).JSON(&res)
	}

	if err.Code() == errro.EFEEDS_NO_FEEDS && ff.By != "" {
		res := helper.GenericResponse{
			Code:    errro.EFEEDS_DELETE_FAIL,
			Message: "unauthorized user to performe this actio",
		}
		return ctx.Status(fiber.StatusUnauthorized).JSON(&res)
	}

	if err.Code() == errro.EFEEDS_NO_FEEDS {
		res := helper.GenericResponse{
			Code:    errro.EFEEDS_DELETE_FAIL,
			Message: err.Error(),
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(&res)
	}

	res := helper.GenericResponse{
		Code:    err.Code(),
		Message: err.Error(),
	}
	return ctx.Status(fiber.StatusNoContent).JSON(&res)
}
