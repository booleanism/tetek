package middleware

import (
	"github.com/booleanism/tetek/feeds/cmd/http/router"
	"github.com/booleanism/tetek/feeds/recipes"
	"github.com/gofiber/fiber/v3"
)

func Feeds(rec recipes.FeedRecipes) fiber.Handler {
	return router.Feeds(rec)
}
