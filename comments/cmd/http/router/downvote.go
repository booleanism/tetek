package router

import (
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/gofiber/fiber/v3"
)

func (r commRouter) Downvote(ctx fiber.Ctx) error {
	_, log := loggr.GetLogger(ctx.Context(), ctx.Route().Name)
	log.Info("new comment downvote request")
	return fiber.NewError(fiber.StatusNotImplemented)
}
