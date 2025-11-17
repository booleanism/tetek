package router

import (
	"context"
	"time"

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

func (cr commRouter) GetComments(ctx fiber.Ctx) error {
	c, log := loggr.GetLogger(ctx.Context(), ctx.Route().Name)
	log.Info("new get comments request")

	gRes := helper.GenericResponse{}
	req := recipes.GetCommentsRequest{}
	if err := helper.BindRequest(ctx, &req); err != nil {
		log.V(1).Info(err.Error())
		return err.SendError(ctx, fiber.StatusBadRequest)
	}

	cto, cancel := context.WithTimeout(
		c,
		Timeout*time.Second)
	defer cancel()

	comms, err := cr.r.GetComments(cto, req)
	if err == nil {
		gRes.Code = errro.SUCCESS
		gRes.Message = "success fetching comments"
		res := recipes.GetCommentsResponse{GenericResponse: gRes, Details: comms}
		return ctx.Status(fiber.StatusOK).JSON(&res)
	}

	gRes.Code = err.Code()
	gRes.Message = err.Error()
	return ctx.Status(fiber.StatusInternalServerError).JSON(&gRes)
}
