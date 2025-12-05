package middlewares

import (
	"context"
	"regexp"
	"strings"

	messaging "github.com/booleanism/tetek/auth/infra/messaging/rabbitmq"
	"github.com/booleanism/tetek/pkg/contracts"
	"github.com/booleanism/tetek/pkg/contracts/adapter"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/gofiber/fiber/v3"
)

type authRequest struct {
	Authorization string `header:"Authorization"`
}

func OptionalAuth(auth contracts.AuthDealer) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		jwt, er := checkJwt(ctx)
		if er != nil {
			return ctx.Next()
		}

		// The consequencies about this implementation that is
		// system assumes it's valid jwt format if user passes Authorization header.
		// And if auth worker dead, then it will wait for 5 second to make sure.
		// Keep calm, it's affected only and only if auth worker is dead.
		c := context.WithValue(ctx.Context(), keystore.AuthTask{}, &messaging.AuthTask{Jwt: jwt})
		cto, cancel := context.WithTimeout(
			c,
			helper.Timeout,
		)
		defer cancel()

		var authRes *messaging.AuthResult
		if err := actualAuth(cto, auth, &authRes); err != nil {
			return ctx.Next()
		}

		if authRes == nil {
			return ctx.Next()
		}

		ctx.SetContext(context.WithValue(c, keystore.AuthRes{}, authRes))
		return ctx.Next()
	}
}

func checkJwt(ctx fiber.Ctx) (string, errro.Error) {
	req := authRequest{}
	if err := helper.BindRequest(ctx, &req); err != nil {
		return "", errro.New(errro.ErrInvalidRequest, err.Error())
	}

	res := helper.GenericResponse{}
	if req.Authorization == "" {
		res.Code = errro.ErrAuthMissingHeader
		res.Message = "missing authorization header"
		e := errro.New(res.Code, res.Message)
		return "", e.WithDetail(res.JSON(), errro.TDetailJSON)
	}

	jwt, ok := strings.CutPrefix(req.Authorization, "Bearer ")
	if !ok {
		res.Code = errro.ErrAuthMissmatchScheme
		res.Message = "mismatch authorization mechanism"
		e := errro.New(res.Code, res.Message)
		return "", e.WithDetail(res.JSON(), errro.TDetailJSON)
	}

	r := regexp.MustCompile(`^(?:[\w-]*\.){2}[\w-]*$`)
	if !r.Match([]byte(jwt)) {
		res.Code = errro.ErrAuthJWTMalformat
		res.Message = "jwt malformat"
		e := errro.New(res.Code, res.Message)
		return "", e.WithDetail(res.JSON(), errro.TDetailJSON)
	}

	return jwt, nil
}

func Auth(auth contracts.AuthDealer) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		jwt, er := checkJwt(ctx)
		res := helper.GenericResponse{}
		if er != nil {
			res.Code = er.Code()
			res.Message = er.Error()
			return er.WithDetail(res.JSON(), errro.TDetailJSON).SendError(ctx, fiber.StatusBadRequest)
		}

		c := context.WithValue(ctx.Context(), keystore.AuthTask{}, &messaging.AuthTask{Jwt: jwt})
		cto, cancel := context.WithTimeout(
			c,
			helper.Timeout,
		)
		defer cancel()

		var authRes *messaging.AuthResult
		if err := actualAuth(cto, auth, &authRes); err != nil {
			if err.Code() == errro.ErrServiceUnavailable {
				return err.SendError(ctx, fiber.StatusRequestTimeout)
			}
			return err.SendError(ctx, fiber.StatusUnauthorized)
		}

		if authRes == nil {
			res.Code = errro.ErrAuthJWTVerifyFail
			res.Message = "authorization failed"
			e := errro.New(res.Code, res.Message)
			return e.WithDetail(res.JSON(), errro.TDetailJSON).SendError(ctx, fiber.StatusUnauthorized)
		}

		ctx.SetContext(context.WithValue(c, keystore.AuthRes{}, authRes))
		return ctx.Next()
	}
}

func actualAuth(ctx context.Context, auth contracts.AuthDealer, authRes **messaging.AuthResult) errro.ResError {
	authTask, ok := ctx.Value(keystore.AuthTask{}).(*messaging.AuthTask)
	res := helper.GenericResponse{}
	if !ok {
		res.Code = errro.ErrAuthEmptyJWT
		res.Message = "did not receive jwt"
		e := errro.New(res.Code, res.Message)
		return e.WithDetail(res.JSON(), errro.TDetailJSON)
	}

	if err := adapter.AuthAdapter(ctx, auth, *authTask, authRes); err != nil {
		res.Code = errro.ErrServiceUnavailable
		res.Message = "auth service unavailable: publishing auth task"
		e := errro.New(res.Code, res.Message)
		return e.WithDetail(res.JSON(), errro.TDetailJSON)
	}

	if (*authRes).Code == errro.Success {
		return nil
	}

	res.Code = errro.ErrAuthJWTVerifyFail
	res.Message = "authorization failed"
	e := errro.New(res.Code, res.Message)
	return e.WithDetail(res.JSON(), errro.TDetailJSON)
}
