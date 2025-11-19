package router

import (
	"github.com/booleanism/tetek/account/internal/model"
	"github.com/booleanism/tetek/account/recipes"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/gofiber/fiber/v3"
)

func Regist(rec recipes.RegistRecipes) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		gRes := helper.GenericResponse{}
		req := recipes.RegistRequest{}
		if err := helper.BindRequest(ctx, &req); err != nil {
			return err.SendError(ctx, fiber.StatusBadRequest)
		}

		err := rec.Regist(ctx, model.User{
			Uname:  req.Uname,
			Email:  req.Email,
			Passwd: req.Passwd,
		})

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
}
