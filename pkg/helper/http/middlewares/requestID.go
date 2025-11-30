package middlewares

import (
	"context"

	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func GenerateRequestID(ctx fiber.Ctx) error {
	ctx.SetContext(context.WithValue(ctx.Context(), keystore.RequestID{}, uuid.NewString()))
	return ctx.Next()
}
