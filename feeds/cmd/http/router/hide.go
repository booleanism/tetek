package router

import (
	"github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/feeds/internal/repo"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/loggr"
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
	loggr.LogInfo(func(z loggr.LogInf) {
		z.V(4).Info("new incoming hide request")
	})
	req := hideRequest{}
	if err := helper.BindRequest(ctx, &req); err != nil {
		return loggr.LogRes(func(z loggr.LogErr) errro.ResError {
			z.V(4).Error(err, "failed to bind request", "body", ctx.Body())
			return err
		}).SendError(ctx, fiber.StatusBadRequest)
	}

	jwt, ok := ctx.Locals("jwt").(*amqp.AuthResult)
	if !ok {
		return loggr.LogRes(func(z loggr.LogErr) errro.ResError {
			res := helper.GenericResponse{
				Code:    errro.EAUTH_INVALID_AUTH_RESULT_TYPE,
				Message: "does not represent jwt type",
			}
			e := errro.New(res.Code, res.Message).WithDetail(res.Json(), errro.TDETAIL_JSON)
			z.V(4).Error(e, res.Message)
			return e
		}).SendError(ctx, fiber.StatusBadRequest)
	}

	ff := repo.FeedsFilter{
		Id:       req.Id,
		HiddenTo: jwt.Claims.Uname,
	}

	err := fr.rec.Hide(ctx, ff, jwt)
	if err != nil {
		return loggr.LogRes(func(z loggr.LogErr) errro.ResError {
			res := helper.GenericResponse{
				Code:    err.Code(),
				Message: err.Error(),
			}
			z.V(4).Error(err, res.Message)
			return errro.New(res.Code, res.Message).WithDetail(res.Json(), errro.TDETAIL_JSON)
		}).SendError(ctx, fiber.StatusInternalServerError)
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
