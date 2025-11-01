package middleware

import (
	"github.com/booleanism/tetek/feeds/cmd/http/router"
	"github.com/gofiber/fiber/v3"
)

func Feeds(router *router.FeedsRouter) fiber.Handler {
	return router.GetFeeds
}
