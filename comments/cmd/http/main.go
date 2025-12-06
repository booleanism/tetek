package main

import (
	"context"
	"os"

	messaging "github.com/booleanism/tetek/comments/internal/infra/messaging/rabbitmq"
	handlers "github.com/booleanism/tetek/comments/internal/presentation/handlers/fiber"
	"github.com/booleanism/tetek/comments/internal/usecases"
	"github.com/booleanism/tetek/comments/internal/usecases/repo"
	db "github.com/booleanism/tetek/infra/db/sql"
	"github.com/booleanism/tetek/pkg/contracts"
	"github.com/booleanism/tetek/pkg/helper/http/middlewares"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/gofiber/fiber/v3"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

const (
	ServiceName = "comments"
	LogV        = 2
)

func main() {
	zl := zerolog.New(os.Stderr)
	zerologr.SetMaxV(LogV)

	dbStr := os.Getenv("COMMS_DB_STR")
	if dbStr == "" {
		panic("dbStr should not empty")
	}
	dbPool := db.Register(dbStr)
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

	baseCtx := context.Background()

	feedsContr := contracts.FeedsAssent(mqCon)
	feedsLisCtx := logr.NewContext(baseCtx, loggr.NewLogger(ServiceName, &zl))
	if err := feedsContr.FeedsResListener(feedsLisCtx, ServiceName); err != nil {
		panic(err)
	}

	authContr := contracts.AuthAssent(mqCon)
	authLisCtx := logr.NewContext(baseCtx, loggr.NewLogger(ServiceName, &zl))
	if err := authContr.AuthResListener(authLisCtx, ServiceName); err != nil {
		panic(err)
	}

	repo := repo.NewCommentsRepo(dbPool)
	rec := usecases.NewCommentsUsecases(repo)
	handlers := handlers.NewHandlers(rec, feedsContr)

	uc := usecases.NewCommentsUsecases(repo)

	workerCtx := logr.NewContext(baseCtx, loggr.NewLogger(ServiceName, &zl))
	ch, err := messaging.NewComments(mqCon, uc).WorkerCommentsListener(workerCtx)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := ch.Close(); err != nil {
			panic(err)
		}
	}()

	app := fiber.New()
	apiEp := app.Group("/api/v0")

	apiEp.Use(middlewares.GenerateRequestID)
	apiEp.Use(middlewares.Logger("comments-service", &zl))

	{
		apiEp.Post("/", middlewares.Auth(authContr), handlers.NewComment).Name("new-comment-handler")
	}

	if err := app.Listen(":8084"); err != nil {
		panic(err)
	}
}
