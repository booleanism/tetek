package main

import (
	"context"
	"os"

	messaging "github.com/booleanism/tetek/account/internal/infra/messaging/rabbitmq"
	handlers "github.com/booleanism/tetek/account/internal/presentation/handlers/fiber"
	"github.com/booleanism/tetek/account/internal/usecases"
	"github.com/booleanism/tetek/account/internal/usecases/repo"
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
	ServiceName = "account"
	LogV        = 2
)

func main() {
	zl := zerolog.New(os.Stderr)
	zerologr.SetMaxV(LogV)

	dbStr := os.Getenv("ACCOUNT_DB_STR")
	if dbStr == "" {
		panic("database connection string empty")
	}

	dbPool := db.Register(dbStr)
	defer dbPool.Close()

	mqStr := os.Getenv("BROKER_STR")
	if mqStr == "" {
		panic("amqp connection string empty")
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

	authContr := contracts.AuthAssent(mqCon)
	authLisCtx := logr.NewContext(baseCtx, loggr.NewLogger(ServiceName, &zl))
	if err := authContr.AuthResListener(authLisCtx, ServiceName); err != nil {
		panic(err)
	}

	rep := repo.NewUserRepo(dbPool)
	uc := usecases.NewAccountUseCases(rep)
	acc := messaging.NewAccount(mqCon, uc)

	workerCtx := logr.NewContext(baseCtx, loggr.NewLogger(ServiceName, &zl))
	ch, err := acc.WorkerAccountListener(workerCtx)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := ch.Close(); err != nil {
			panic(err)
		}
	}()

	h := handlers.NewHandlers(uc)

	app := fiber.New()
	api := app.Group("/api/v0")
	{
		api.Use(middlewares.GenerateRequestID)
		api.Use(middlewares.Logger(ServiceName, &zl))
		api.Post("/", h.RegistUser).Name("registration-handler")
		api.Get("/:uname", middlewares.Auth(authContr), h.Profile).Name("profile-handler")
	}

	if err := app.Listen(":8082"); err != nil {
		panic(err)
	}
}
