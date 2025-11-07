package router

import (
	"strings"

	"github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/feeds/internal/repo"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/loggr"
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
	loggr.LogInfo(func(z loggr.LogInf) {
		z.V(4).Info("new incoming delete request")
	})
	req := deleteRequest{}
	if err := helper.BindRequest(ctx, &req); err != nil {
		return loggr.LogRes(func(z loggr.LogErr) errro.ResError {
			z.V(4).Error(err, "failed to bind request")
			return err
		}).SendError(ctx, fiber.StatusBadRequest)
	}

	j := ctx.Locals("jwt")
	jwt, ok := j.(*amqp.AuthResult)
	if !ok {
		return loggr.LogRes(func(z loggr.LogErr) errro.ResError {
			res := helper.GenericResponse{
				Code:    errro.EAUTH_INVALID_AUTH_RESULT_TYPE,
				Message: "missing jwt",
			}
			e := errro.New(res.Code, "does not represent jwt type").WithDetail(res.Json(), errro.TDETAIL_JSON)
			z.V(4).Error(e, res.Message, "type", j)
			return e
		}).SendError(ctx, fiber.StatusUnauthorized)
	}

	ff := repo.FeedsFilter{
		Id: req.Id,
	}

	// only moderator freely to delete feed
	if strings.ToLower(jwt.Claims.Role) != "m" {
		loggr.LogInfo(func(z loggr.LogInf) {
			z.V(4).Info("normal user not moderator")
		})
		ff.By = jwt.Claims.Uname
	}

	err := fr.rec.Delete(ctx, ff, jwt)
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
