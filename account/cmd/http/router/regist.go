package router

import (
	"github.com/booleanism/tetek/account/recipes"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/gofiber/fiber/v3"
)

func (r router) Regist(ctx fiber.Ctx) error {
	c, log := loggr.GetLogger(ctx.Context(), ctx.Route().Name)
	log.V(1).Info("new registration request")

	gRes := helper.GenericResponse{}
	req := recipes.RegistRequest{}
	if err := helper.BindRequest(ctx, &req); err != nil {
		return err.SendError(ctx, fiber.StatusBadRequest)
	}

	err := r.rec.Regist(c, req)
	if err == nil {
		res := recipes.RegistResponse{
			GenericResponse: helper.GenericResponse{
				Code:    errro.Success,
				Message: "register success",
			},
			Detail: req,
		}
		return ctx.Status(fiber.StatusOK).JSON(&res)
	}

	gRes.Code = err.Code()
	gRes.Message = err.Error()
	if err.Code() == errro.ErrAccountUserAlreadyExist {
		return err.WithDetail(gRes.JSON(), errro.TDetailJSON).SendError(ctx, fiber.StatusConflict)
	}

	if err.Code() == errro.ErrAccountInvalidRegistParam {
		return err.WithDetail(gRes.JSON(), errro.TDetailJSON).SendError(ctx, fiber.StatusBadRequest)
	}

	gRes.Code = errro.ErrAccountRegistFail
	gRes.Message = "cannot create user"
	return err.WithDetail(gRes.JSON(), errro.TDetailJSON).SendError(ctx, fiber.StatusInternalServerError)
}
