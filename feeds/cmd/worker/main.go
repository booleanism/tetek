package main

import (
	"os"

	"github.com/Masterminds/squirrel"
	"github.com/booleanism/tetek/db"
	"github.com/booleanism/tetek/feeds/internal/contract"
	"github.com/booleanism/tetek/feeds/internal/repo"
	"github.com/rabbitmq/amqp091-go"
)

func main() {
	dbStr := os.Getenv("FEEDS_DB_STR")
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

	sq := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	rep := repo.New(dbPool, sq)
	acc := contract.NewFeeds(mqCon, rep)
	ch, err := acc.WorkerFeedsListener()
	if err != nil {
	}
	defer ch.Close()

	select {}

}
