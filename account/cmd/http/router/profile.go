package router

import (
	"github.com/booleanism/tetek/account/internal/model"
	"github.com/booleanism/tetek/account/recipes"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/gofiber/fiber/v3"
)

func Profile(rec recipes.ProfileRecipes) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		_, log := loggr.GetLogger(ctx.Context(), "profile-handler")
		gRes := helper.GenericResponse{}
		req := recipes.ProfileRequest{}
		if err := helper.BindRequest(ctx, &req); err != nil {
			log.V(1).Info("failed to bind request", "error", err)
			return err.SendError(ctx, fiber.StatusBadRequest)
		}

		if req.Uname == "" {
			gRes.Code = errro.ErrAccountEmptyParam
			gRes.Message = "uname empty"
			e := errro.New(gRes.Code, gRes.Message)
			return e.WithDetail(gRes.JSON(), errro.TDetailJSON).SendError(ctx, fiber.StatusBadRequest)
		}

		u, err := rec.Profile(ctx.Context(), model.User{Uname: req.Uname})
		if err == nil {
			res := recipes.ProfileResponse{
				GenericResponse: helper.GenericResponse{
					Code:    errro.Success,
					Message: "user profiling success",
				},
				Detail: u,
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
