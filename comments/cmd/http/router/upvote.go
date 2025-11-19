package router

import (
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/gofiber/fiber/v3"
)

func (cr commRouter) Upvote(ctx fiber.Ctx) error {
	_, log := loggr.GetLogger(ctx.Context(), ctx.Route().Name)
	log.Info("new comment upvote request")
	return fiber.NewError(fiber.StatusNotImplemented)
}
