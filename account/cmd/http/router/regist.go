package router

import (
	"github.com/booleanism/tetek/account/internal/model"
	"github.com/booleanism/tetek/account/recipes"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/go-logr/logr"
	"github.com/gofiber/fiber/v3"
)

type registRequest struct {
	Uname  string `json:"uname"`
	Email  string `json:"email"`
	Passwd string `json:"passwd"`
}

type registResponse struct {
	Detail registRequest `json:"detail"`
	helper.GenericResponse
}

func Regist(rec recipes.RegistRecipes) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		loggr.Log.V(4).Info("new incoming regist request")
		req := registRequest{}
		if res, err := helper.BindRequest(ctx, &req); err != nil {
			return loggr.Log.Error(3, func(z logr.LogSink) errro.Error {
				z.Error(err, res.Message)
				return errro.FromError(res.Code, ctx.Status(fiber.StatusBadRequest).JSON(&res))
			}).ToFiber()
		}

		err := rec.Regist(ctx, model.User{
			Uname:  req.Uname,
			Email:  req.Email,
			Passwd: req.Passwd,
		})

		if err == nil {
			res := registResponse{
				GenericResponse: helper.GenericResponse{
					Code:    errro.SUCCESS,
					Message: "register success",
				},
			}
			loggr.Log.V(4).Info("success registering user", "response", res)
			return ctx.Status(fiber.StatusOK).JSON(&res)
		}

		if err.Code() == errro.EACCOUNT_USER_ALREADY_EXIST {
			res := registResponse{
				GenericResponse: helper.GenericResponse{
					Code:    err.Code(),
					Message: err.Error(),
				},
				Detail: req,
			}
			return ctx.Status(fiber.StatusConflict).JSON(&res)
		}

		if err.Code() == errro.EACCOUNT_INVALID_REGIST_PARAM {
			res := registResponse{
				GenericResponse: helper.GenericResponse{
					Code:    err.Code(),
					Message: err.Error(),
				},
				Detail: req,
			}
			return ctx.Status(fiber.StatusBadRequest).JSON(&res)
		}

		res := registResponse{
			GenericResponse: helper.GenericResponse{
				Code:    errro.EACCOUNT_REGIST_FAIL,
				Message: "cannot create user",
			},
			Detail: req,
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(&res)
	}
}
