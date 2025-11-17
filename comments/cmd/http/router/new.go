package router

import (
	"context"
	"time"

	"github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/comments/recipes"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/gofiber/fiber/v3"
)

const Timeout = 5

func (cr commRouter) NewComment(ctx fiber.Ctx) error {
	c, log := loggr.GetLogger(ctx.Context(), ctx.Route().Name)
	log.Info("new new comment request")

	gRes := helper.GenericResponse{}
	req := recipes.NewCommentRequest{}
	if err := helper.BindRequest(ctx, &req); err != nil {
		log.V(1).Info(err.Error())
		return err.SendError(ctx, fiber.StatusBadRequest)
	}

	_, ok := c.Value(keystore.AuthRes{}).(*amqp.AuthResult)
	if !ok {
		gRes.Code = errro.EAUTH_EMPTY_JWT
		gRes.Message = "empty jwt"
		e := errro.New(gRes.Code, gRes.Message)
		log.V(1).Info(e.Error())
		return e.WithJson(gRes).SendError(ctx, fiber.StatusBadRequest)
	}

	cto, cancel := context.WithTimeout(c, Timeout*time.Minute)
	defer cancel()

	com, err := cr.r.NewComment(cto, req)
	if err != nil {
		gRes.Code = err.Code()
		gRes.Message = err.Error()
		return ctx.Status(fiber.StatusInternalServerError).JSON(&gRes)
	}

	gRes.Code = errro.SUCCESS
	gRes.Message = "success adding comment"
	res := recipes.NewCommentResponse{GenericResponse: gRes, Detail: com}
	return ctx.Status(fiber.StatusOK).JSON(&res)
}
