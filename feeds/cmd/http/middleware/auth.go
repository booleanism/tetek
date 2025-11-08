package middleware

import (
	"strings"

	"github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/feeds/internal/contract"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

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
	if err := helper.BindRequest(ctx, &req); err != nil {
		return "", errro.New(errro.INVALID_REQ, err.Error())
	}

	if req.Authorization == "" {
		res := helper.GenericResponse{
			Code:    errro.EAUTH_MISSING_HEADER,
			Message: "missing authorization header",
		}
		e := errro.New(res.Code, res.Message)
		return "", e.WithDetail(res.Json(), errro.TDETAIL_JSON)
	}

	jwt, ok := strings.CutPrefix(req.Authorization, "Bearer ")
	if !ok {
		res := helper.GenericResponse{
			Code:    errro.EAUTH_MISSMATCH_AUTH_MECHANISM,
			Message: "mismatch authorization mechanism",
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

		if err := actualAuth(ctx, auth, jwt); err != nil {
			return err.SendError(ctx, fiber.StatusUnauthorized)
		}

		return ctx.Next()
	}
}

func actualAuth(ctx fiber.Ctx, auth *contract.LocalAuthContr, jwt string) errro.ResError {
	id := uuid.NewString()
	task := amqp.AuthTask{Jwt: jwt}
	if err := auth.Publish(id, task); err != nil {
		res := helper.GenericResponse{
			Code:    errro.EAUTH_SERVICE_UNAVAILABLE,
			Message: "auth service unavailable: publishing auth task",
		}
		e := errro.New(res.Code, res.Message)
		return e.WithDetail(res.Json(), errro.TDETAIL_JSON)
	}

	authRes, err := auth.Consume(id)
	if err != nil {
		res := helper.GenericResponse{
			Code:    errro.EAUTH_SERVICE_UNAVAILABLE,
			Message: "auth service unavailable: consuming auth result",
		}
		e := errro.New(res.Code, res.Message)
		return e.WithDetail(res.Json(), errro.TDETAIL_JSON)
	}

	if authRes.Code == errro.SUCCESS {
		ctx.Locals("jwt", authRes)
		return nil
	}

	res := helper.GenericResponse{
		Code:    errro.EAUTH_JWT_VERIFY_FAIL,
		Message: "authorization failed",
	}

	e := errro.New(res.Code, res.Message)
	return e.WithDetail(res.Json(), errro.TDETAIL_JSON)
}
