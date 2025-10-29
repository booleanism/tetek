package middleware

import (
	"strings"

	"github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/feeds/internal/contract"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/go-logr/logr"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type authResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type authRequest struct {
	Authorization string `header:"Authorization"`
}

func OptionalAuth(auth *contract.LocalAuthContr) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		jwt, er := checkJwt(ctx)
		if er != nil {
			return ctx.Next()
		}

		if err := actualAuth(ctx, auth, jwt); err != nil {
			return ctx.Next()
		}

		return ctx.Next()
	}
}

func checkJwt(ctx fiber.Ctx) (string, errro.Error) {
	req := authRequest{}
	if res, err := helper.BindRequest(ctx, &req); err != nil {
		return "", loggr.Log.Error(3, func(z logr.LogSink) errro.Error {
			z.Error(err, res.Message)
			return errro.New(res.Code, res.Message)
		})
	}

	if req.Authorization == "" {
		res := authResponse{
			Code:    errro.EAUTH_MISSING_HEADER,
			Message: "missing authorization header",
		}
		return "", loggr.Log.Error(4, func(z logr.LogSink) errro.Error {
			e := errro.New(res.Code, res.Message)
			z.Error(e, "missing authorization header")
			return e
		})
	}

	jwt, ok := strings.CutPrefix(req.Authorization, "Bearer ")
	if !ok {
		res := authResponse{
			Code:    errro.EAUTH_MISSMATCH_AUTH_MECHANISM,
			Message: "mismatch authorization mechanism",
		}
		return "", loggr.Log.Error(4, func(z logr.LogSink) errro.Error {
			e := errro.New(res.Code, res.Message)
			z.Error(e, "mismatch authorization mechanism, expected Bearer")
			return e
		})
	}

	return jwt, nil
}

func Auth(auth *contract.LocalAuthContr) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		loggr.Log.V(4).Info("new incoming authorization request")

		jwt, er := checkJwt(ctx)
		if er != nil {
			res := authResponse{
				Code:    er.Code(),
				Message: er.Error(),
			}
			return ctx.Status(fiber.StatusBadRequest).JSON(&res)
		}

		if err := actualAuth(ctx, auth, jwt); err != nil {
			return err
		}

		return ctx.Next()
	}
}

func actualAuth(ctx fiber.Ctx, auth *contract.LocalAuthContr, jwt string) error {
	id := uuid.NewString()
	task := amqp.AuthTask{Jwt: jwt}
	if err := auth.Publish(id, task); err != nil {
		res := authResponse{
			Code:    errro.EAUTH_SERVICE_UNAVAILABLE,
			Message: "auth service unavailable: publishing auth task",
		}
		return loggr.Log.Error(0, func(z logr.LogSink) errro.Error {
			e := errro.FromError(res.Code, ctx.Status(fiber.StatusServiceUnavailable).JSON(&res))
			z.Error(err, res.Message, "id", id, "task", task)
			return e
		}).ToFiber()
	}

	authRes, err := auth.Consume(id)
	if err != nil {
		res := authResponse{
			Code:    errro.EAUTH_SERVICE_UNAVAILABLE,
			Message: "auth service unavailable: consuming auth result",
		}
		return loggr.Log.Error(0, func(z logr.LogSink) errro.Error {
			e := errro.FromError(res.Code, ctx.Status(fiber.StatusServiceUnavailable).JSON(&res))
			z.Error(err, res.Message, "id", id, "task sent", task)
			return e
		}).ToFiber()
	}

	if authRes.Code == errro.SUCCESS {
		ctx.Locals("jwt", authRes)
		loggr.Log.V(4).Info("authorization success. forwarded into next middleware")
		return nil
	}

	res := authResponse{
		Code:    errro.EAUTH_JWT_VERIFY_FAIL,
		Message: "authorization failed",
	}
	return loggr.Log.Error(2, func(z logr.LogSink) errro.Error {
		e := errro.FromError(res.Code, ctx.Status(fiber.StatusUnauthorized).JSON(&res))
		z.Info(2, "authorization failed", "id", id, "task", task, "auth result", authRes)
		return e
	}).ToFiber()
}
