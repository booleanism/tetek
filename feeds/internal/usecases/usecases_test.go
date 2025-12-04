package usecases_test

import (
	"context"
	"testing"
	"time"

	"github.com/booleanism/tetek/feeds/schemas"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"github.com/testcontainers/testcontainers-go/wait"
)

type cont struct {
	*rabbitmq.RabbitMQContainer
	*postgres.PostgresContainer
	dbConStr string
	mqConStr string
}

var container = &cont{}

func init() {
	ctx := context.Background()
	mq, err := rabbitmq.Run(ctx,
		"rabbitmq:4.1-management-alpine",
	)
	if err != nil {
		panic(err)
	}

	conStr, err := mq.AmqpURL(ctx)
	if err != nil {
		panic(err)
	}

	container.mqConStr = conStr
	container.RabbitMQContainer = mq

	pg, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase("feedstestdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		panic(err)
	}

	conStr, err = pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(err)
	}

	cmd := []string{
		"sh", "-c",
		"psql -U postgres -d feedstestdb <<'EOF'\n" + schemas.FeedsSQL + "\nEOF",
	}
	_, _, err = pg.Exec(ctx, cmd)
	if err != nil {
		panic(err)
	}

	container.dbConStr = conStr
	container.PostgresContainer = pg
}

func TestMain(m *testing.M) {
	m.Run()
	if err := container.PostgresContainer.Terminate(context.Background()); err != nil {
		panic(err)
	}
	if err := container.RabbitMQContainer.Terminate(context.Background()); err != nil {
		panic(err)
	}
}
