package middleware

import (
	"context"
	"regexp"
	"time"

	"github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/comments/cmd/http/api"
	"github.com/booleanism/tetek/feeds/internal/contract"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/gofiber/fiber/v3"
)

func OptionalAuth(auth *contract.LocalAuthContr) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		jwt, er := checkJwt(ctx)
		if er != nil {
			return ctx.Next()
		}

		// The consequencies about this implementation that is
		// system assumes it's valid jwt format if user passes Authorization header.
		// And if auth worker dead, then it will wait for 5 second to make sure.
		// Keep calm, it's affected only and only if auth worker is dead.
		id := ctx.Locals(helper.RequestIdKey{}).(string)
		cto, cancel := context.WithTimeout(
			context.WithValue(
				ctx, helper.RequestIdKey{},
				id,
			),
			5*time.Second,
		)
		defer cancel()

		var authRes *amqp.AuthResult
		if err := actualAuth(cto, auth, jwt, &authRes); err != nil {
			return ctx.Next()
		}

		if authRes == nil {
			ctx.Next()
		}

		ctx.Locals("jwt", authRes)
		return ctx.Next()
	}
}

func checkJwt(ctx fiber.Ctx) (string, errro.Error) {
	jwt, ok := ctx.Locals(api.BearerAuthScopes).(string)
	if !ok {
		res := helper.GenericResponse{
			Code:    errro.EAUTH_MISSING_HEADER,
			Message: "missing authorization header",
		}
		e := errro.New(res.Code, res.Message)
		return "", e.WithDetail(res.Json(), errro.TDETAIL_JSON)
	}

	r := regexp.MustCompile(`^(?:[\w-]*\.){2}[\w-]*$`)
	if !r.Match([]byte(jwt)) {
		res := helper.GenericResponse{
			Code:    errro.EAUTH_JWT_MALFORMAT,
			Message: "malformat jwt",
		}
		e := errro.New(res.Code, res.Message)
		return "", e.WithDetail(res.Json(), errro.TDETAIL_JSON)
	}

	return jwt, nil
}

func Auth(auth *contract.LocalAuthContr) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		jwt, er := checkJwt(ctx)
		if er != nil {
			res := helper.GenericResponse{
				Code:    er.Code(),
				Message: er.Error(),
			}
			return er.WithDetail(res.Json(), errro.TDETAIL_JSON).SendError(ctx, fiber.StatusBadRequest)
		}

		id := ctx.Locals(helper.RequestIdKey{}).(string)
		cto, cancel := context.WithTimeout(
			context.WithValue(
				ctx, helper.RequestIdKey{},
				id,
			),
			10*time.Second,
		)
		defer cancel()

		var authRes *amqp.AuthResult
		if err := actualAuth(cto, auth, jwt, &authRes); err != nil {
			if err.Code() == errro.EAUTH_SERVICE_UNAVAILABLE {
				return err.SendError(ctx, fiber.StatusRequestTimeout)
			}

			return err.SendError(ctx, fiber.StatusUnauthorized)
		}
		if authRes == nil {
			res := helper.GenericResponse{
				Code:    errro.EAUTH_JWT_VERIFY_FAIL,
				Message: "authorization failed",
			}

			e := errro.New(res.Code, res.Message)
			return e.WithDetail(res.Json(), errro.TDETAIL_JSON).SendError(ctx, fiber.StatusUnauthorized)
		}

		ctx.Locals("jwt", authRes)

		return ctx.Next()
	}
}

func actualAuth(ctx context.Context, auth *contract.LocalAuthContr, jwt string, authRes **amqp.AuthResult) errro.ResError {
	task := amqp.AuthTask{Jwt: jwt}
	if err := auth.Publish(ctx, task); err != nil {
		res := helper.GenericResponse{
			Code:    errro.EAUTH_SERVICE_UNAVAILABLE,
			Message: "auth service unavailable: publishing auth task",
		}
		e := errro.New(res.Code, res.Message)
		return e.WithDetail(res.Json(), errro.TDETAIL_JSON)
	}

	authr, err := auth.Consume(ctx)
	if err != nil {
		res := helper.GenericResponse{
			Code:    errro.EAUTH_SERVICE_UNAVAILABLE,
			Message: "auth service unavailable: consuming auth result",
		}
		e := errro.New(res.Code, res.Message)
		return e.WithDetail(res.Json(), errro.TDETAIL_JSON)
	}

	if authr.Code == errro.SUCCESS {
		authRes = &authr
		return nil
	}

	res := helper.GenericResponse{
		Code:    errro.EAUTH_JWT_VERIFY_FAIL,
		Message: "authorization failed",
	}

	e := errro.New(res.Code, res.Message)
	return e.WithDetail(res.Json(), errro.TDETAIL_JSON)
}
