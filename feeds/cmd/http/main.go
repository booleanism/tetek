package main

import (
	"os"

	"github.com/Masterminds/squirrel"
	"github.com/booleanism/tetek/db"
	"github.com/booleanism/tetek/feeds/cmd/http/middleware"
	"github.com/booleanism/tetek/feeds/cmd/http/router"
	"github.com/booleanism/tetek/feeds/cmd/http/router/feeds"
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
	repo := repo.New(dbPool, &sq)
	acc := contract.NewAccount(mqCon)
	auth := contract.NewAuth(mqCon)
	rec := recipes.NewRecipes(repo, acc)

	app := fiber.New()
	api := app.Group("/api/v0")
	{
		api.Get("/", middleware.OptionalAuth(auth), router.Feeds(rec))
		api.Post("/", middleware.Auth(auth), middleware.Feeds(rec), feeds.New(rec))
		api.Delete("/", middleware.Feeds(rec), feeds.Delete(rec))
		api.Put("/hide", middleware.Feeds(rec), feeds.Hide(rec))
	}

	app.Listen(":8083")
}
