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

func (r profileResponse) JSON() []byte {
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
				Code:    errro.ErrAccountEmptyParam,
				Message: "uname empty",
			}
			e := errro.New(res.Code, res.Message)
			return e.WithDetail(res.Json(), errro.TDetailJSON).SendError(ctx, fiber.StatusBadRequest)
		}

		u, err := rec.Profile(ctx.Context(), model.User{Uname: req.Uname})
		if err != nil {
			if err.Code() == errro.ErrAccountNoUser {
				res := profileResponse{
					GenericResponse: helper.GenericResponse{
						Code:    err.Code(),
						Message: "user not found",
					},
					Detail: model.User{Uname: req.Uname},
				}
				return err.WithDetail(res.JSON(), errro.TDetailJSON).SendError(ctx, fiber.StatusNotFound)
			}

			res := profileResponse{
				GenericResponse: helper.GenericResponse{
					Code:    errro.ErrAccountDBError,
					Message: "something happen in our end",
				},
				Detail: model.User{Uname: req.Uname},
			}
			return err.WithDetail(res.JSON(), errro.TDetailJSON).SendError(ctx, fiber.StatusInternalServerError)
		}

		res := profileResponse{
			GenericResponse: helper.GenericResponse{
				Code:    errro.Success,
				Message: "user profiling success",
			},
			Detail: u,
		}
		return ctx.Status(fiber.StatusOK).JSON(res)
	}
}
