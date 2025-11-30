package main

import (
	"context"
	"os"

	"github.com/Masterminds/squirrel"
	"github.com/booleanism/tetek/db"
	"github.com/booleanism/tetek/feeds/internal/contract"
	"github.com/booleanism/tetek/feeds/internal/repo"
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

	sq := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	rep := repo.New(dbPool, sq)

	baseCtx := context.Background()

	workerCtx := logr.NewContext(baseCtx, loggr.NewLogger(ServiceName, &zl))
	feeds := contract.NewFeeds(mqCon, rep)
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
