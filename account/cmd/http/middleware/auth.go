package middleware

import (
	"strings"

	"github.com/booleanism/tetek/account/internal/contract"
	"github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/loggr"
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
			return loggr.LogRes(func(z loggr.LogErr) errro.ResError {
				z.V(3).Error(err, "failed to bind request", "header", ctx.GetHeaders())
				return err
			}).SendError(ctx, fiber.StatusBadRequest)
		}

		if req.Authorization == "" {
			res := authResponse{
				GenericResponse: helper.GenericResponse{
					Code:    errro.EAUTH_MISSING_HEADER,
					Message: "missing authorization header",
				},
			}
			return loggr.LogRes(func(z loggr.LogErr) errro.ResError {
				e := errro.New(res.Code, res.Message)
				z.V(4).Error(e, res.Message, "request", req)
				return e.WithDetail(res.Json(), errro.TDETAIL_JSON)
			}).SendError(ctx, fiber.StatusBadRequest)
		}

		jwt, ok := strings.CutPrefix(req.Authorization, "Bearer ")
		if !ok {
			res := authResponse{
				GenericResponse: helper.GenericResponse{
					Code:    errro.EAUTH_MISSMATCH_AUTH_MECHANISM,
					Message: "mismatch authorization mechanism",
				},
			}
			return loggr.LogRes(func(z loggr.LogErr) errro.ResError {
				e := errro.New(res.Code, res.Message)
				z.V(4).Error(e, res.Message, "request", req)
				return e.WithDetail(res.Json(), errro.TDETAIL_JSON)
			}).SendError(ctx, fiber.StatusBadRequest)
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
			return loggr.LogRes(func(z loggr.LogErr) errro.ResError {
				e := errro.New(res.Code, res.Message)
				z.V(0).Error(err, "cannot publish auth task to auth service", "id", id, "task", task)
				return e.WithDetail(res.Json(), errro.TDETAIL_JSON)
			}).SendError(ctx, fiber.StatusServiceUnavailable)
		}

		authRes, err := auth.Consume(id)
		if err != nil {
			res := authResponse{
				GenericResponse: helper.GenericResponse{
					Code:    errro.EAUTH_SERVICE_UNAVAILABLE,
					Message: "auth service unavailable",
				},
			}
			return loggr.LogRes(func(z loggr.LogErr) errro.ResError {
				e := errro.New(res.Code, res.Message)
				z.V(0).Error(err, "cannot consume auth task to auth service", "id", id, "task", task)
				return e.WithDetail(res.Json(), errro.TDETAIL_JSON)
			})
		}

		if authRes.Code == errro.SUCCESS {
			ctx.Locals("jwt", authRes)
			loggr.LogInfo(func(z loggr.LogInf) {
				z.V(4).Info("authorization success. forwarded into next middleware")
			})
			return ctx.Next()
		}

		res := authResponse{
			GenericResponse: helper.GenericResponse{
				Code:    errro.EAUTH_JWT_VERIFY_FAIL,
				Message: "authorization failed",
			},
		}
		return loggr.LogRes(func(z loggr.LogErr) errro.ResError {
			e := errro.New(res.Code, res.Message)
			z.V(2).Error(e, "authorization failed", "id", id, "auth result", authRes)
			return e.WithDetail(res.Json(), errro.TDETAIL_JSON)
		}).SendError(ctx, fiber.StatusInternalServerError)
	}
}
