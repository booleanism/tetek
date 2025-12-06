package handlers

import (
	"context"

	"github.com/booleanism/tetek/comments/internal/usecases"
	"github.com/booleanism/tetek/comments/internal/usecases/dto"
	"github.com/booleanism/tetek/pkg/contracts"
	"github.com/booleanism/tetek/pkg/contracts/adapter"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/gofiber/fiber/v3"
)

type commHandlers struct {
	r usecases.CommentsUseCases
	f contracts.FeedsDealer
}

func NewHandlers(r usecases.CommentsUseCases, f contracts.FeedsDealer) commHandlers {
	return commHandlers{r, f}
}

func (cr commHandlers) NewComment(ctx fiber.Ctx) error {
	c, log := loggr.GetLogger(ctx.Context(), ctx.Route().Name)
	log.V(1).Info("new new comment request")

	gRes := helper.GenericResponse{}
	req := dto.NewCommentRequest{}
	if err := helper.BindRequest(ctx, &req); err != nil {
		return err.SendError(ctx, fiber.StatusBadRequest)
	}

	cto, cancel := context.WithTimeout(c, helper.Timeout)
	defer cancel()

	err := cr.r.NewComment(cto, cr.f, adapter.FeedsAdapter, req)
	if err != nil {
		gRes.Code = err.Code()
		gRes.Message = err.Error()
		return ctx.Status(fiber.StatusInternalServerError).JSON(&gRes)
	}

	gRes.Code = errro.Success
	gRes.Message = "success adding comment"
	res := dto.NewCommentResponse{GenericResponse: gRes}
	return ctx.Status(fiber.StatusOK).JSON(&res)
}
