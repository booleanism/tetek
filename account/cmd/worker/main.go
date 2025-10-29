package main

import (
	"os"

	"github.com/booleanism/tetek/account/internal/contract"
	"github.com/booleanism/tetek/account/internal/repo"
	"github.com/booleanism/tetek/db"
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

	rep := repo.NewUserRepo(dbPool)
	acc := contract.NewAccount(mqCon, rep)
	acc.WorkerAccountListener()
}
