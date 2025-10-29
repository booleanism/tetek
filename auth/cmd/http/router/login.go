package router

import (
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
		if res, err := helper.BindRequest(ctx, &req); err != nil {
			return loggr.Log.Error(3, func(z logr.LogSink) errro.Error {
				z.Error(err, res.Message)
				return errro.FromError(res.Code, ctx.Status(fiber.StatusBadRequest).JSON(&res))
			}).ToFiber()
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

		if err.Code() == errro.EAUTH_JWT_GENERATAION_FAIL {
			res := loginResponse{
				Code:    errro.EAUTH_JWT_GENERATAION_FAIL,
				Message: err.Error(),
				Detail:  req,
			}
			return ctx.Status(fiber.StatusInternalServerError).JSON(res)
		}

		if err.Code() == errro.EACCOUNT_SERVICE_UNAVAILABLE {
			res := loginResponse{
				Code:    err.Code(),
				Message: err.Error(),
				Detail:  req,
			}
			return ctx.Status(fiber.StatusServiceUnavailable).JSON(res)
		}

		if err.Code() == errro.EACCOUNT_NO_USER {
			res := loginResponse{
				Code:    err.Code(),
				Message: err.Error(),
				Detail:  req,
			}
			return ctx.Status(fiber.StatusExpectationFailed).JSON(res)
		}

		if err.Code() == errro.EAUTH_INVALID_LOGIN_PARAM {
			res := loginResponse{
				Code:    err.Code(),
				Message: err.Error(),
				Detail:  req,
			}
			return ctx.Status(fiber.StatusExpectationFailed).JSON(res)
		}

		if err.Code() == errro.EAUTH_INVALID_CREDS {
			res := loginResponse{
				Code:    err.Code(),
				Message: err.Error(),
				Detail:  req,
			}
			return ctx.Status(fiber.StatusUnauthorized).JSON(res)
		}

		res := loginResponse{
			Code:    errro.EACCOUNT_CANT_LOGIN,
			Message: "failed to proccess login request",
			Detail:  req,
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(res)
	}
}
