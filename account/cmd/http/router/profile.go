package router

import (
	"encoding/json"

	"github.com/booleanism/tetek/account/internal/model"
	"github.com/booleanism/tetek/account/recipes"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/gofiber/fiber/v3"
)

type profileRequest struct {
	Uname string `uri:"uname"`
}

type profileResponse struct {
	helper.GenericResponse
	Detail model.User `json:"detail"`
}

func (r profileResponse) Json() []byte {
	j, _ := json.Marshal(r)
	return j
}

func Profile(rec recipes.ProfileRecipes) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		req := profileRequest{}
		if err := helper.BindRequest(ctx, &req); err != nil {
			return err.SendError(ctx, fiber.StatusBadRequest)
		}

		if req.Uname == "" {
			res := helper.GenericResponse{
				Code:    errro.EACCOUNT_EMPTY_PARAM,
				Message: "uname empty",
			}
			e := errro.New(res.Code, res.Message)
			return e.WithDetail(res.Json(), errro.TDETAIL_JSON).SendError(ctx, fiber.StatusBadRequest)
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
				return err.WithDetail(res.Json(), errro.TDETAIL_JSON).SendError(ctx, fiber.StatusNotFound)
			}

			res := profileResponse{
				GenericResponse: helper.GenericResponse{
					Code:    errro.EACCOUNT_DB_ERR,
					Message: "something happen in our end",
				},
				Detail: model.User{Uname: req.Uname},
			}
			return err.WithDetail(res.Json(), errro.TDETAIL_JSON).SendError(ctx, fiber.StatusInternalServerError)
		}

		res := profileResponse{
			GenericResponse: helper.GenericResponse{
				Code:    errro.SUCCESS,
				Message: "user profiling success",
			},
			Detail: u,
		}
		return ctx.Status(fiber.StatusOK).JSON(res)
	}
}
