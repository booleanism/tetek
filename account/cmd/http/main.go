package main

import (
	"os"

	"github.com/booleanism/tetek/account/cmd/http/middleware"
	"github.com/booleanism/tetek/account/cmd/http/router"
	"github.com/booleanism/tetek/account/internal/contract"
	"github.com/booleanism/tetek/account/internal/repo"
	"github.com/booleanism/tetek/account/recipes"
	"github.com/booleanism/tetek/db"
	"github.com/gofiber/fiber/v3"
	"github.com/rabbitmq/amqp091-go"
)

func main() {
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

	auth := contract.NewAuth(mqCon)

	rep := repo.NewUserRepo(dbPool)
	rec := recipes.New(rep)
	acc := contract.NewAccount(mqCon, rep)
	ch, err := acc.WorkerAccountListener()
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
		api.Post("/", router.Regist(rec))
		api.Get("/:uname", middleware.Auth(auth), router.Profile(rec))
	}

	if err := app.Listen(":8082"); err != nil {
		panic(err)
	}
}
