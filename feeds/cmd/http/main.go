package main

import (
	"context"
	"os"

	"github.com/Masterminds/squirrel"
	"github.com/booleanism/tetek/db"
	"github.com/booleanism/tetek/docs"
	"github.com/booleanism/tetek/feeds/cmd/http/router"
	"github.com/booleanism/tetek/feeds/internal/contract"
	"github.com/booleanism/tetek/feeds/internal/repo"
	"github.com/booleanism/tetek/feeds/recipes"
	"github.com/booleanism/tetek/pkg/contracts"
	"github.com/booleanism/tetek/pkg/helper/http/middlewares"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

const (
	ServiceName = "feeds"
	LogV        = 2
)

func main() {
	zl := zerolog.New(os.Stderr)
	zerologr.SetMaxV(LogV)

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

	baseCtx := context.Background()

	acc := contracts.SubscribeAccount(mqCon)
	accLisCtx := logr.NewContext(baseCtx, loggr.NewLogger(ServiceName, &zl))
	if err := acc.AccountResListener(accLisCtx, ServiceName); err != nil {
		panic(err)
	}

	authContr := contracts.SubsribeAuth(mqCon)
	authLisCtx := logr.NewContext(baseCtx, loggr.NewLogger(ServiceName, &zl))
	if err := authContr.AuthResListener(authLisCtx, ServiceName); err != nil {
		panic(err)
	}

	rec := recipes.NewRecipes(repo, acc)

	feedsContr := contract.NewFeeds(mqCon, repo)
	if _, err := feedsContr.WorkerFeedsListener(); err != nil {
		panic(err)
	}

	router := router.NewFeedRouter(rec)

	endp := "/api/v0"
	app := fiber.New()
	apiEp := app.Group(endp)

	d, ui := docs.OapiDocs(apiEp, docs.Feeds, endp)

	apiEp.Use(cors.New())
	apiEp.Use(middlewares.GenerateRequestID)
	apiEp.Use(middlewares.Logger(ServiceName, &zl))

	{
		apiEp.Get("openapi.yaml", func(ctx fiber.Ctx) error {
			return ctx.SendString(d())
		})

		apiEp.Get("docs/", func(ctx fiber.Ctx) error {
			ctx.Set("Content-Type", "text/html")

			return ctx.SendString(ui)
		})

		apiEp.Get("/", middlewares.Auth(authContr), router.GetFeeds).Name("getFeeds-handler")
		apiEp.Post("/", middlewares.Auth(authContr), router.NewFeed).Name("newFeed-handler")
		apiEp.Delete("/:id", middlewares.Auth(authContr), router.DeleteFeed).Name("deleteFeed-handler")
		apiEp.Patch("/hide", middlewares.Auth(authContr), router.HideFeed).Name("hideFeed-handler")
	}

	if err := app.Listen(":8083"); err != nil {
		panic(err)
	}
}
