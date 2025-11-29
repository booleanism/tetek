package main

import (
	"context"
	"os"

	"github.com/booleanism/tetek/account/internal/contract"
	"github.com/booleanism/tetek/account/internal/repo"
	"github.com/booleanism/tetek/db"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

const LogV = 3

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

	rep := repo.NewUserRepo(dbPool)
	acc := contract.NewAccount(mqCon, rep)

	workerCtx := logr.NewContext(context.Background(), loggr.NewLogger("account", &zl))
	ch, err := acc.WorkerAccountListener(workerCtx)
	if err != nil {
		panic(err.Error())
	}
	defer func() {
		if err := ch.Close(); err != nil {
			panic(err)
		}
	}()

	select {}
}
