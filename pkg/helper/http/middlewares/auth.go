package middlewares

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/pkg/contracts"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/gofiber/fiber/v3"
)

type authRequest struct {
	Authorization string `header:"Authorization"`
}

func OptionalAuth(auth contracts.AuthSubscribe) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		jwt, er := checkJwt(ctx)
		if er != nil {
			return ctx.Next()
		}

		// The consequencies about this implementation that is
		// system assumes it's valid jwt format if user passes Authorization header.
		// And if auth worker dead, then it will wait for 5 second to make sure.
		// Keep calm, it's affected only and only if auth worker is dead.
		c := context.WithValue(ctx.Context(), keystore.AuthTask{}, &amqp.AuthTask{Jwt: jwt})
		cto, cancel := context.WithTimeout(
			c,
			1*time.Second,
		)
		defer cancel()

		var authRes *amqp.AuthResult
		if err := actualAuth(cto, auth, &authRes); err != nil {
			return ctx.Next()
		}

		if authRes == nil {
			ctx.Next()
		}

		ctx.SetContext(context.WithValue(c, keystore.AuthRes{}, authRes))
		return ctx.Next()
	}
}

func checkJwt(ctx fiber.Ctx) (string, errro.Error) {
	req := authRequest{}
	if err := helper.BindRequest(ctx, &req); err != nil {
		return "", errro.New(errro.INVALID_REQ, err.Error())
	}

	res := helper.GenericResponse{}
	if req.Authorization == "" {
		res.Code = errro.EAUTH_MISSING_HEADER
		res.Message = "missing authorization header"
		e := errro.New(res.Code, res.Message)
		return "", e.WithDetail(res.Json(), errro.TDETAIL_JSON)
	}

	jwt, ok := strings.CutPrefix(req.Authorization, "Bearer ")
	if !ok {
		res.Code = errro.EAUTH_MISSMATCH_AUTH_MECHANISM
		res.Message = "mismatch authorization mechanism"
		e := errro.New(res.Code, res.Message)
		return "", e.WithDetail(res.Json(), errro.TDETAIL_JSON)
	}

	r := regexp.MustCompile(`^(?:[\w-]*\.){2}[\w-]*$`)
	if !r.Match([]byte(jwt)) {
		res.Code = errro.EAUTH_JWT_MALFORMAT
		res.Message = "jwt malformat"
		e := errro.New(res.Code, res.Message)
		return "", e.WithDetail(res.Json(), errro.TDETAIL_JSON)
	}

	return jwt, nil
}

func Auth(auth contracts.AuthSubscribe) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		jwt, er := checkJwt(ctx)
		res := helper.GenericResponse{}
		if er != nil {
			res.Code = er.Code()
			res.Message = er.Error()
			return er.WithDetail(res.Json(), errro.TDETAIL_JSON).SendError(ctx, fiber.StatusBadRequest)
		}

		c := context.WithValue(ctx.Context(), keystore.AuthTask{}, &amqp.AuthTask{Jwt: jwt})
		cto, cancel := context.WithTimeout(
			c,
			5*time.Second,
		)
		defer cancel()

		var authRes *amqp.AuthResult
		if err := actualAuth(cto, auth, &authRes); err != nil {
			if err.Code() == errro.EAUTH_SERVICE_UNAVAILABLE {
				return err.SendError(ctx, fiber.StatusRequestTimeout)
			}
			return err.SendError(ctx, fiber.StatusUnauthorized)
		}

		if authRes == nil {
			res.Code = errro.EAUTH_JWT_VERIFY_FAIL
			res.Message = "authorization failed"
			e := errro.New(res.Code, res.Message)
			return e.WithDetail(res.Json(), errro.TDETAIL_JSON).SendError(ctx, fiber.StatusUnauthorized)
		}

		ctx.SetContext(context.WithValue(c, keystore.AuthRes{}, authRes))
		return ctx.Next()
	}
}

func actualAuth(ctx context.Context, auth contracts.AuthSubscribe, authRes **amqp.AuthResult) errro.ResError {
	authTask, ok := ctx.Value(keystore.AuthTask{}).(*amqp.AuthTask)
	res := helper.GenericResponse{}
	if !ok {
		res.Code = errro.EAUTH_EMPTY_JWT
		res.Message = "did not receive jwt"
		e := errro.New(res.Code, res.Message)
		return e.WithDetail(res.Json(), errro.TDETAIL_JSON)
	}

	if err := authAdapter(ctx, auth, *authTask, authRes); err != nil {
		res.Code = errro.EAUTH_SERVICE_UNAVAILABLE
		res.Message = "auth service unavailable: publishing auth task"
		e := errro.New(res.Code, res.Message)
		return e.WithDetail(res.Json(), errro.TDETAIL_JSON)
	}

	if (*authRes).Code == errro.SUCCESS {
		return nil
	}

	res.Code = errro.EAUTH_JWT_VERIFY_FAIL
	res.Message = "authorization failed"
	e := errro.New(res.Code, res.Message)
	return e.WithDetail(res.Json(), errro.TDETAIL_JSON)
}

func authAdapter(ctx context.Context, auth contracts.AuthSubscribe, t amqp.AuthTask, res **amqp.AuthResult) errro.Error {
	if err := auth.Publish(ctx, t); err != nil {
		e := errro.FromError(errro.ECOMM_PUB_FAIL, "failed to publish auth task", err)
		return e
	}

	err := auth.Consume(ctx, res)
	if err != nil {
		e := errro.FromError(errro.ECOMM_CONSUME_FAIL, "failed to consume auth result", err)
		return e
	}
	return nil
}
