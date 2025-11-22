package main

import (
	"context"
	"os"

	"github.com/booleanism/tetek/account/cmd/http/router"
	"github.com/booleanism/tetek/account/internal/contract"
	"github.com/booleanism/tetek/account/internal/repo"
	"github.com/booleanism/tetek/account/recipes"
	"github.com/booleanism/tetek/db"
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
	db.Register(dbStr)
	dbPool := db.GetPool()
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

	authContr := contracts.SubsribeAuth(mqCon)
	authLisCtx := logr.NewContext(baseCtx, loggr.NewLogger(ServiceName, &zl))
	if err := authContr.AuthResListener(authLisCtx, ServiceName); err != nil {
		panic(err)
	}

	rep := repo.NewUserRepo(dbPool)
	rec := recipes.New(rep)
	acc := contract.NewAccount(mqCon, rep)

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

	app := fiber.New()
	api := app.Group("/api/v0")
	{
		api.Use(middlewares.GenerateRequestID)
		api.Use(middlewares.Logger(ServiceName, &zl))
		api.Post("/", router.Regist(rec)).Name("registration-handler")
		api.Get("/:uname", middlewares.Auth(authContr), router.Profile(rec)).Name("profile-handler")
	}

	if err := app.Listen(":8082"); err != nil {
		panic(err)
	}
}
