package middleware

import (
	"strings"

	"github.com/booleanism/tetek/account/internal/contract"
	"github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/go-logr/logr"
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
		if res, err := helper.BindRequest(ctx, &req); err != nil {
			return loggr.Log.Error(3, func(z logr.LogSink) errro.Error {
				z.Error(err, res.Message)
				return errro.FromError(res.Code, ctx.Status(fiber.StatusBadRequest).JSON(&res))
			}).ToFiber()
		}

		if req.Authorization == "" {
			res := authResponse{
				GenericResponse: helper.GenericResponse{
					Code:    errro.EAUTH_MISSING_HEADER,
					Message: "missing authorization header",
				},
			}
			return loggr.Log.Error(4, func(z logr.LogSink) errro.Error {
				e := errro.FromError(res.Code, ctx.Status(fiber.StatusBadRequest).JSON(&res))
				z.Error(e, res.Message, "request", req)
				return e
			}).ToFiber()
		}

		jwt, ok := strings.CutPrefix(req.Authorization, "Bearer ")
		if !ok {
			res := authResponse{
				GenericResponse: helper.GenericResponse{
					Code:    errro.EAUTH_MISSMATCH_AUTH_MECHANISM,
					Message: "mismatch authorization mechanism",
				},
			}
			return loggr.Log.Error(4, func(z logr.LogSink) errro.Error {
				e := errro.FromError(res.Code, ctx.Status(fiber.StatusBadRequest).JSON(&res))
				z.Error(e, res.Message, "request", req)
				return e
			}).ToFiber()
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
			return loggr.Log.Error(0, func(z logr.LogSink) errro.Error {
				e := errro.FromError(res.Code, ctx.Status(fiber.StatusServiceUnavailable).JSON(&res))
				z.Error(err, "cannot publish auth task to auth service", "id", id, "task", task)
				return e
			}).ToFiber()
		}

		authRes, err := auth.Consume(id)
		if err != nil {
			res := authResponse{
				GenericResponse: helper.GenericResponse{
					Code:    errro.EAUTH_SERVICE_UNAVAILABLE,
					Message: "auth service unavailable",
				},
			}
			return loggr.Log.Error(0, func(z logr.LogSink) errro.Error {
				e := errro.FromError(res.Code, ctx.Status(fiber.StatusServiceUnavailable).JSON(&res))
				z.Error(err, "cannot consume auth task to auth service", "id", id, "task", task)
				return e
			}).ToFiber()
		}

		if authRes.Code == errro.SUCCESS {
			ctx.Locals("jwt", authRes)
			loggr.Log.V(4).Info("authorization success. forwarded into next middleware")
			return ctx.Next()
		}

		res := authResponse{
			GenericResponse: helper.GenericResponse{
				Code:    errro.EAUTH_JWT_VERIFY_FAIL,
				Message: "authorization failed",
			},
		}
		return loggr.Log.Error(2, func(z logr.LogSink) errro.Error {
			e := errro.FromError(res.Code, ctx.Status(fiber.StatusUnauthorized).JSON(&res))
			z.Info(2, "authorization failed", "id", id, "auth result", authRes)
			return e
		}).ToFiber()
	}
}
