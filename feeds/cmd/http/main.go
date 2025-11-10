package main

import (
	"os"

	"github.com/Masterminds/squirrel"
	"github.com/booleanism/tetek/db"
	"github.com/booleanism/tetek/docs"
	"github.com/booleanism/tetek/feeds/cmd/http/middleware"
	_ "github.com/booleanism/tetek/feeds/cmd/http/middleware"
	"github.com/booleanism/tetek/feeds/cmd/http/router"
	"github.com/booleanism/tetek/feeds/internal/contract"
	"github.com/booleanism/tetek/feeds/internal/repo"
	"github.com/booleanism/tetek/feeds/recipes"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/rabbitmq/amqp091-go"
)

func main() {
	dbStr := os.Getenv("FEEDS_DB_STR")
	if dbStr == "" {
		panic("feeds database string empty")
	}
	db.Register(dbStr)
	dbPool := db.GetPool()
	defer dbPool.Close()

	mqStr := os.Getenv("BROKER_STR")
	if mqStr == "" {
		panic("broker connection string should not empty")
	}

	mqCon, err := amqp091.Dial(mqStr)
	if err != nil {
		panic(err)
	}

	sq := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	repo := repo.New(dbPool, sq)
	acc := contract.NewAccount(mqCon)
	auth := contract.NewAuth(mqCon)
	rec := recipes.NewRecipes(repo, acc)

	feedsContr := contract.NewFeeds(mqCon, repo)
	feedsContr.WorkerFeedsListener()

	router := router.NewFeedRouter(rec)

	endp := "/api/v0"
	app := fiber.New()
	apiEp := app.Group(endp)

	d, ui := docs.OapiDocs(apiEp, docs.Feeds, endp)

	apiEp.Use(cors.New())
	apiEp.Use(helper.GenerateRequestId)

	apiEp.Get("openapi.yaml", func(ctx fiber.Ctx) error {
		return ctx.SendString(d())
	})

	apiEp.Get("docs/", func(ctx fiber.Ctx) error {
		ctx.Set("Content-Type", "text/html")

		return ctx.SendString(ui)
	})

	apiEp.Get("/", middleware.OptionalAuth(auth), router.GetFeeds)
	apiEp.Post("/", middleware.Auth(auth), router.NewFeed)
	apiEp.Delete("/:id", middleware.Auth(auth), router.DeleteFeed)
	apiEp.Patch("/hide", middleware.Auth(auth), router.HideFeed)

	app.Listen(":8083")
}
