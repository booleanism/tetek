package handlers

import (
	"github.com/booleanism/tetek/account/internal/usecases"
	"github.com/booleanism/tetek/account/internal/usecases/dto"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/gofiber/fiber/v3"
)

type handlers struct {
	uc usecases.AccountUseCases
}

func NewHandlers(uc usecases.AccountUseCases) handlers {
	return handlers{uc}
}

func (r handlers) Profile(ctx fiber.Ctx) error {
	c, log := loggr.GetLogger(ctx.Context(), ctx.Route().Name)
	log.V(1).Info("new profile request")

	gRes := helper.GenericResponse{}
	req := dto.ProfileRequest{}
	if err := helper.BindRequest(ctx, &req); err != nil {
		return err.SendError(ctx, fiber.StatusBadRequest)
	}

	if req.Uname == "" {
		gRes.Code = errro.ErrAccountEmptyParam
		gRes.Message = "uname empty"
		e := errro.New(gRes.Code, gRes.Message)
		log.V(1).Info("missing required field", "error", e)
		return e.WithDetail(gRes.JSON(), errro.TDetailJSON).SendError(ctx, fiber.StatusBadRequest)
	}

	u := &dto.User{}
	err := r.uc.GetProfile(c, req, &u)
	if err == nil {
		gRes.Code = errro.Success
		gRes.Message = "user profiling success"
		res := dto.ProfileResponse{
			GenericResponse: gRes,
			Detail:          *u,
		}
		return ctx.Status(fiber.StatusOK).JSON(res)
	}

	if err.Code() == errro.ErrAccountNoUser {
		gRes.Code = errro.ErrAccountEmptyParam
		gRes.Message = "user not found"
		return err.WithDetail(gRes.JSON(), errro.TDetailJSON).SendError(ctx, fiber.StatusNotFound)
	}

	gRes.Code = errro.ErrAccountDBError
	gRes.Message = "something happen in our end"
	return err.WithDetail(gRes.JSON(), errro.TDetailJSON).SendError(ctx, fiber.StatusInternalServerError)
}
