package middleware

import (
	"strings"

	"github.com/booleanism/tetek/account/internal/contract"
	"github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type authResponse struct {
	helper.GenericResponse
}

type authRequest struct {
	Authorization string `header:"Authorization"`
}

func Auth(auth *contract.LocalAuthContr) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		req := authRequest{}
		if err := helper.BindRequest(ctx, &req); err != nil {
			return err.SendError(ctx, fiber.StatusBadRequest)
		}

		if req.Authorization == "" {
			res := authResponse{
				GenericResponse: helper.GenericResponse{
					Code:    errro.EAUTH_MISSING_HEADER,
					Message: "missing authorization header",
				},
			}
			e := errro.New(res.Code, res.Message)
			return e.WithDetail(res.Json(), errro.TDETAIL_JSON).SendError(ctx, fiber.StatusBadRequest)
		}

		jwt, ok := strings.CutPrefix(req.Authorization, "Bearer ")
		if !ok {
			res := authResponse{
				GenericResponse: helper.GenericResponse{
					Code:    errro.EAUTH_MISSMATCH_AUTH_MECHANISM,
					Message: "mismatch authorization mechanism",
				},
			}
			e := errro.New(res.Code, res.Message)
			return e.WithDetail(res.Json(), errro.TDETAIL_JSON).SendError(ctx, fiber.StatusBadRequest)
		}

		id := uuid.NewString()
		task := amqp.AuthTask{Jwt: jwt}
		if err := auth.Publish(id, task); err != nil {
			res := authResponse{
				GenericResponse: helper.GenericResponse{
					Code:    errro.EAUTH_SERVICE_UNAVAILABLE,
					Message: "auth service unavailable",
				},
			}
			e := errro.New(res.Code, res.Message)
			return e.WithDetail(res.Json(), errro.TDETAIL_JSON).SendError(ctx, fiber.StatusServiceUnavailable)
		}

		authRes, err := auth.Consume(id)
		if err != nil {
			res := authResponse{
				GenericResponse: helper.GenericResponse{
					Code:    errro.EAUTH_SERVICE_UNAVAILABLE,
					Message: "auth service unavailable",
				},
			}
			e := errro.New(res.Code, res.Message)
			return e.WithDetail(res.Json(), errro.TDETAIL_JSON).SendError(ctx, fiber.StatusServiceUnavailable)
		}

		if authRes.Code == errro.SUCCESS {
			ctx.Locals("jwt", authRes)
			return ctx.Next()
		}

		res := authResponse{
			GenericResponse: helper.GenericResponse{
				Code:    errro.EAUTH_JWT_VERIFY_FAIL,
				Message: "authorization failed",
			},
		}
		e := errro.New(res.Code, res.Message)
		return e.WithDetail(res.Json(), errro.TDETAIL_JSON).SendError(ctx, fiber.StatusInternalServerError)
	}
}
