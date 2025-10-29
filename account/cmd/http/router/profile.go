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

type profileRequest struct {
	Uname string `uri:"uname"`
}

type profileResponse struct {
	Detail model.User `json:"detail"`
	helper.GenericResponse
}

func Profile(rec recipes.ProfileRecipes) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		loggr.Log.V(4).Info("new incoming profile request")
		req := profileRequest{}
		if res, err := helper.BindRequest(ctx, &req); err != nil {
			return loggr.Log.Error(3, func(z logr.LogSink) errro.Error {
				z.Error(err, res.Message)
				return errro.FromError(res.Code, ctx.Status(fiber.StatusBadRequest).JSON(&res))
			}).ToFiber()
		}

		if req.Uname == "" {
			res := profileResponse{
				GenericResponse: helper.GenericResponse{
					Code:    errro.EACCOUNT_EMPTY_PARAM,
					Message: "empty param",
				},
			}
			loggr.Log.V(2).Info("invalid param", "param", req.Uname, "response", res)
			return ctx.Status(fiber.StatusBadRequest).JSON(&res)
		}

		u, err := rec.Profile(ctx.Context(), model.User{Uname: req.Uname})
		if err != nil {
			if err.Code() == errro.EACCOUNT_NO_USER {
				res := profileResponse{
					GenericResponse: helper.GenericResponse{
						Code:    err.Code(),
						Message: "user not found",
					},
					Detail: model.User{Uname: req.Uname},
				}
				return ctx.Status(fiber.StatusNotFound).JSON(&res)
			}

			res := profileResponse{
				GenericResponse: helper.GenericResponse{
					Code:    errro.EACCOUNT_DB_ERR,
					Message: "something happen in our end",
				},
				Detail: model.User{Uname: req.Uname},
			}
			return ctx.Status(fiber.StatusInternalServerError).JSON(&res)
		}

		res := profileResponse{
			GenericResponse: helper.GenericResponse{
				Code:    errro.SUCCESS,
				Message: "user profiling success",
			},
			Detail: u,
		}
		loggr.Log.V(4).Info("request success", "response", u)
		return ctx.Status(fiber.StatusOK).JSON(res)
	}
}
