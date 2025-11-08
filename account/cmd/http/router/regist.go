package router

import (
	"encoding/json"

	"github.com/booleanism/tetek/account/internal/model"
	"github.com/booleanism/tetek/account/recipes"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/gofiber/fiber/v3"
)

type registRequest struct {
	Uname  string `json:"uname"`
	Email  string `json:"email"`
	Passwd string `json:"passwd"`
}

type registResponse struct {
	helper.GenericResponse
	Detail registRequest `json:"detail"`
}

func (r registResponse) Json() []byte {
	j, _ := json.Marshal(r)
	return j
}

func Regist(rec recipes.RegistRecipes) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		req := registRequest{}
		if err := helper.BindRequest(ctx, &req); err != nil {
			return err.SendError(ctx, fiber.StatusBadRequest)
		}

		err := rec.Regist(ctx, model.User{
			Uname:  req.Uname,
			Email:  req.Email,
			Passwd: req.Passwd,
		})

		if err == nil {
			res := helper.GenericResponse{
				Code:    errro.SUCCESS,
				Message: "register success",
			}
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

			return err.WithDetail(res.Json(), errro.TDETAIL_JSON).SendError(ctx, fiber.StatusConflict)
		}

		if err.Code() == errro.EACCOUNT_INVALID_REGIST_PARAM {
			res := registResponse{
				GenericResponse: helper.GenericResponse{
					Code:    err.Code(),
					Message: err.Error(),
				},
				Detail: req,
			}

			return err.WithDetail(res.Json(), errro.TDETAIL_JSON).SendError(ctx, fiber.StatusBadRequest)
		}

		res := registResponse{
			GenericResponse: helper.GenericResponse{
				Code:    errro.EACCOUNT_REGIST_FAIL,
				Message: "cannot create user",
			},
			Detail: req,
		}
		return err.WithDetail(res.Json(), errro.TDETAIL_JSON).SendError(ctx, fiber.StatusInternalServerError)
	}
}
