package router

import (
	"fmt"

	"github.com/booleanism/tetek/auth/recipes"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/gofiber/fiber/v3"
)

func Login(logRec recipes.LoginRecipe) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		c, log := loggr.GetLogger(ctx.Context(), ctx.Route().Name)
		log.V(1).Info("new login request")

		req := recipes.LoginRequest{}
		if err := helper.BindRequest(ctx, &req); err != nil {
			return err.SendError(ctx, fiber.StatusBadRequest)
		}

		jwt, err := logRec.Login(c, req)
		if err == nil {
			res := recipes.LoginResponse{
				Code:    errro.Success,
				Message: "login success",
				Detail: recipes.LoginRequest{
					Uname: req.Uname,
					Email: req.Email,
				},
			}
			ctx.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))
			return ctx.Status(fiber.StatusOK).JSONP(res)
		}

		res := recipes.LoginResponse{
			Code:    err.Code(),
			Message: err.Error(),
		}
		if err.Code() == errro.ErrAuthJWTGenerationFail || err.Code() == errro.ErrAccountServiceUnavailable {
			return err.WithDetail(res.JSON(), errro.TDetailJSON).SendError(ctx, fiber.StatusInternalServerError)
		}

		if err.Code() == errro.ErrAccountNoUser {
			return err.WithDetail(res.JSON(), errro.TDetailJSON).SendError(ctx, fiber.StatusNotFound)
		}

		if err.Code() == errro.ErrAuthInvalidLoginParam {
			return err.WithDetail(res.JSON(), errro.TDetailJSON).SendError(ctx, fiber.StatusExpectationFailed)
		}

		if err.Code() == errro.ErrAuthInvalidCreds {
			return err.WithDetail(res.JSON(), errro.TDetailJSON).SendError(ctx, fiber.StatusUnauthorized)
		}

		res = recipes.LoginResponse{
			Code:    errro.ErrAccountCantLogin,
			Message: "failed to proccess login request",
		}
		return err.WithDetail(res.JSON(), errro.TDetailJSON).SendError(ctx, fiber.StatusInternalServerError)
	}
}
