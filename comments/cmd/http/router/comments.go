package router

import (
	"context"

	"github.com/booleanism/tetek/comments/recipes"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/gofiber/fiber/v3"
)

type commRouter struct {
	r recipes.CommentsRecipes
}

func NewCommRouter(r recipes.CommentsRecipes) commRouter {
	return commRouter{r}
}

func (cr commRouter) NewComment(ctx fiber.Ctx) error {
	c, log := loggr.GetLogger(ctx.Context(), ctx.Route().Name)
	log.V(1).Info("new new comment request")

	gRes := helper.GenericResponse{}
	req := recipes.NewCommentRequest{}
	if err := helper.BindRequest(ctx, &req); err != nil {
		return err.SendError(ctx, fiber.StatusBadRequest)
	}

	cto, cancel := context.WithTimeout(c, helper.Timeout)
	defer cancel()

	com, err := cr.r.NewComment(cto, req)
	if err != nil {
		gRes.Code = err.Code()
		gRes.Message = err.Error()
		return ctx.Status(fiber.StatusInternalServerError).JSON(&gRes)
	}

	gRes.Code = errro.Success
	gRes.Message = "success adding comment"
	res := recipes.NewCommentResponse{GenericResponse: gRes, Detail: com}
	return ctx.Status(fiber.StatusOK).JSON(&res)
}
