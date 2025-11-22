package middlewares

import (
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/go-logr/logr"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

func Logger(name string, zl *zerolog.Logger) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		reqID, ok := ctx.Context().Value(keystore.RequestID{}).(string)
		if !ok {
			reqID = "<invalid requestId>"
		}

		log := loggr.NewLogger(name, zl)
		log = log.WithValues("requestID", reqID).V(1)
		c := logr.NewContext(ctx.Context(), log)
		ctx.SetContext(c)

		return ctx.Next()
	}
}
