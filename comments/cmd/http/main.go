package main

import (
	"os"

	"github.com/booleanism/tetek/comments/cmd/http/router"
	"github.com/booleanism/tetek/comments/internal/repo"
	"github.com/booleanism/tetek/comments/recipes"
	"github.com/booleanism/tetek/db"
	"github.com/booleanism/tetek/pkg/contracts"
	"github.com/booleanism/tetek/pkg/helper/http/middlewares"
	"github.com/go-logr/zerologr"
	"github.com/gofiber/fiber/v3"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

const LogV = 3

func main() {
	zl := zerolog.New(os.Stderr)
	zerologr.SetMaxV(LogV)

	dbStr := os.Getenv("COMMS_DB_STR")
	if dbStr == "" {
		panic("dbStr should not empty")
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
	defer func() {
		if err := mqCon.Close(); err != nil {
			panic(err)
		}
	}()

	feedsContr := contracts.SubscribeFeeds(mqCon, "comments")
	authContr := contracts.SubsribeAuth(mqCon, "comments")
	repo := repo.NewCommRepo(dbPool)
	rec := recipes.NewCommentRecipes(repo, feedsContr, authContr)
	router := router.NewCommRouter(rec)

	app := fiber.New()
	apiEp := app.Group("/api/v0")

	apiEp.Use(middlewares.GenerateRequestID)
	apiEp.Use(middlewares.Logger("comments-service", &zl))

	apiEp.Get("/:id", middlewares.OptionalAuth(authContr), router.GetComments).Name("get-comments")
	apiEp.Post("/", middlewares.Auth(authContr), router.NewComment).Name("new-comment")
	apiEp.Patch("/upvote", middlewares.Auth(authContr), router.Upvote).Name("upvote-comments")
	apiEp.Patch("/downvote", middlewares.Auth(authContr), router.Downvote).Name("downvote-comments")

	if err := app.Listen(":8084"); err != nil {
		panic(err)
	}
}
