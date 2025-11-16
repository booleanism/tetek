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
		reqId, ok := ctx.Context().Value(keystore.RequestId{}).(string)
		if !ok {
			reqId = "<invalid requestId>"
		}

		log := loggr.NewLogger("comments-service", zl)
		log = log.WithValues("requestId", reqId).V(1)
		c := logr.NewContext(ctx.Context(), log)
		ctx.SetContext(c)

		return ctx.Next()
	}
}
