package main

import (
	"context"
	"os"

	contract "github.com/booleanism/tetek/feeds/internal/infra/messaging/rabbitmq"
	"github.com/booleanism/tetek/feeds/internal/usecases"
	"github.com/booleanism/tetek/feeds/internal/usecases/repo"
	db "github.com/booleanism/tetek/infra/db/sql"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
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

	rep := repo.NewFeedsRepo(dbPool)
	uc := usecases.NewFeedsUsecase(rep)

	baseCtx := context.Background()

	workerCtx := logr.NewContext(baseCtx, loggr.NewLogger(ServiceName, &zl))
	feeds := contract.NewFeeds(mqCon, uc)
	ch, err := feeds.WorkerFeedsListener(workerCtx)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := ch.Close(); err != nil {
			panic(err)
		}
	}()

	select {}
}
