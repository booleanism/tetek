package router

import (
	"encoding/json"

	"github.com/booleanism/tetek/account/amqp"
	"github.com/booleanism/tetek/auth/recipes"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/go-logr/logr"
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
	Jwt     string       `json:"jwt"`
	Detail  loginRequest `json:"detail"`
}

func (r loginResponse) Json() []byte {
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
		loggr.Log.V(4).Info("new incoming login request")
		req := loginRequest{}
		if err := helper.BindRequest(ctx, &req); err != nil {
			return loggr.Log.ErrorRes(3, func(z logr.LogSink) error {
				z.Error(err, "failed to bind request", "body", ctx.Body())
				return err.SendError(ctx, fiber.StatusBadRequest)
			})
		}

		jwt, err := logRec.Login(req.toUser())
		if err == nil {
			res := loginResponse{
				Code:    errro.SUCCESS,
				Message: "login success",
				Jwt:     jwt,
				Detail: loginRequest{
					Uname: req.Uname,
				},
			}
			loggr.Log.V(4).Info(res.Message)
			return ctx.Status(fiber.StatusOK).JSONP(res)
		}

		res := loginResponse{
			Code:    err.Code(),
			Message: err.Error(),
			Detail:  req,
		}
		if err.Code() == errro.EAUTH_JWT_GENERATAION_FAIL || err.Code() == errro.EACCOUNT_SERVICE_UNAVAILABLE {
			return err.WithDetail(res.Json(), errro.TDETAIL_JSON).SendError(ctx, fiber.StatusInternalServerError)
		}

		if err.Code() == errro.EACCOUNT_NO_USER {
			return err.WithDetail(res.Json(), errro.TDETAIL_JSON).SendError(ctx, fiber.StatusNotFound)
		}

		if err.Code() == errro.EAUTH_INVALID_LOGIN_PARAM {
			return err.WithDetail(res.Json(), errro.TDETAIL_JSON).SendError(ctx, fiber.StatusExpectationFailed)
		}

		if err.Code() == errro.EAUTH_INVALID_CREDS {
			return err.WithDetail(res.Json(), errro.TDETAIL_JSON).SendError(ctx, fiber.StatusUnauthorized)
		}

		res = loginResponse{
			Code:    errro.EACCOUNT_CANT_LOGIN,
			Message: "failed to proccess login request",
			Detail:  req,
		}
		return err.WithDetail(res.Json(), errro.TDETAIL_JSON).SendError(ctx, fiber.StatusInternalServerError)
	}
}
