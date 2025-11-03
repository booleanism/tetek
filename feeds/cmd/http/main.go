package main

import (
	"os"

	"github.com/Masterminds/squirrel"
	"github.com/booleanism/tetek/db"
	"github.com/booleanism/tetek/feeds/cmd/http/middleware"
	"github.com/booleanism/tetek/feeds/cmd/http/router"
	"github.com/booleanism/tetek/feeds/internal/contract"
	"github.com/booleanism/tetek/feeds/internal/repo"
	"github.com/booleanism/tetek/feeds/recipes"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/gofiber/fiber/v3"
	"github.com/rabbitmq/amqp091-go"
)

func main() {
	loggr.Register(4)
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

	router := router.NewFeedRouter(rec)

	app := fiber.New()
	api := app.Group("/api/v0")
	{
		api.Get("/", middleware.OptionalAuth(auth), router.GetFeeds)
		// TODO: to use Location header fills with created feeds id and return with 201 code
		api.Post("/", middleware.Auth(auth), router.NewFeed)
		// TODO: Idempotencies, consistent response if the request is identical
		api.Delete("/:id<guid>", middleware.Auth(auth), router.DeleteFeed)
		// TODO: Idempotencies, consistent response if the request is identical
		api.Put("/hide", middleware.Auth(auth), router.HideFeed)
	}

	app.Listen(":8083")
}
