package router

import (
	"github.com/booleanism/tetek/account/recipes"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/gofiber/fiber/v3"
)

func Profile(rec recipes.ProfileRecipes) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		c, log := loggr.GetLogger(ctx.Context(), ctx.Route().Name)
		log.V(1).Info("new profile request")

		gRes := helper.GenericResponse{}
		req := recipes.ProfileRequest{}
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

		u, err := rec.Profile(c, req)
		if err == nil {
			gRes.Code = errro.Success
			gRes.Message = "user profiling success"
			res := recipes.ProfileResponse{
				GenericResponse: gRes,
				Detail:          u,
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
}
