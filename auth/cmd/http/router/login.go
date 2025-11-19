package router

import (
	"encoding/json"
	"fmt"

	"github.com/booleanism/tetek/account/amqp"
	"github.com/booleanism/tetek/auth/recipes"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/gofiber/fiber/v3"
)

type loginRequest struct {
	Uname  string `json:"uname"`
	Email  string `json:"email"`
	Passwd string `json:"passwd"`
}

type loginResponse struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Detail  loginRequest `json:"detail"`
}

func (r loginResponse) JSON() []byte {
	j, _ := json.Marshal(r)
	return j
}

func (req loginRequest) toUser() amqp.User {
	return amqp.User{
		Uname:  req.Uname,
		Email:  req.Email,
		Passwd: req.Passwd,
	}
}

func Login(logRec recipes.LoginRecipe) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		req := loginRequest{}
		if err := helper.BindRequest(ctx, &req); err != nil {
			return err.SendError(ctx, fiber.StatusBadRequest)
		}

		jwt, err := logRec.Login(ctx.Context(), req.toUser())
		if err == nil {
			res := loginResponse{
				Code:    errro.Success,
				Message: "login success",
				Detail: loginRequest{
					Uname: req.Uname,
					Email: req.Email,
				},
			}
			ctx.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))
			return ctx.Status(fiber.StatusOK).JSONP(res)
		}

		res := loginResponse{
			Code:    err.Code(),
			Message: err.Error(),
			Detail:  req,
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

		res = loginResponse{
			Code:    errro.ErrAccountCantLogin,
			Message: "failed to proccess login request",
			Detail:  req,
		}
		return err.WithDetail(res.JSON(), errro.TDetailJSON).SendError(ctx, fiber.StatusInternalServerError)
	}
}
