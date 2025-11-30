package contract_test

import (
	"context"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/booleanism/tetek/db"
	"github.com/booleanism/tetek/feeds/amqp"
	"github.com/booleanism/tetek/feeds/internal/contract"
	"github.com/booleanism/tetek/feeds/internal/repo"
	"github.com/booleanism/tetek/feeds/schemas"
	"github.com/booleanism/tetek/pkg/contracts"
	"github.com/booleanism/tetek/pkg/contracts/adapter"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
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

type testData struct {
	amqp.FeedsTask
	expected int
}

func TestWorker(t *testing.T) {
	con, err := amqp091.Dial(container.mqConStr)
	if err != nil {
		t.Fatal(err)
	}

	p := db.Register(container.dbConStr)
	defer p.Close()

	sq := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	repo := repo.New(p, sq)

	ctx := context.WithValue(context.Background(), keystore.RequestID{}, "test")

	acc := contract.NewFeeds(con, repo)
	ch, err := acc.WorkerFeedsListener(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := ch.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	tFeeds := contracts.FeedsAssent(con)
	if err := tFeeds.FeedsResListener(context.Background(), "test"); err != nil {
		t.Fatal(err)
	}

	idPass, err := uuid.Parse("e13cfca7-6bb0-4f67-ab96-e8941c20911a")
	if err != nil {
		t.Fatal(err)
	}

	res := &amqp.FeedsResult{}
	tasks := []testData{
		{
			FeedsTask: amqp.FeedsTask{
				Feeds: amqp.Feeds{
					ID: uuid.New(),
				},
				Cmd: 1,
			},
			expected: errro.ErrFeedsUnknownCmd,
		},
		{
			FeedsTask: amqp.FeedsTask{
				Feeds: amqp.Feeds{
					ID: idPass,
				},
				Cmd: 0,
			},
			expected: errro.Success,
		},
		{
			FeedsTask: amqp.FeedsTask{
				Feeds: amqp.Feeds{
					ID: uuid.New(),
				},
				Cmd: 0,
			},
			expected: errro.ErrFeedsNoFeeds,
		},
		{
			FeedsTask: amqp.FeedsTask{
				Feeds: amqp.Feeds{
					ID: uuid.New(),
				},
				Cmd: 0,
			},
			expected: errro.ErrFeedsDBError,
		},
	}

	for _, task := range tasks {
		if task.expected == errro.ErrFeedsDBError {
			(*p).Close()
		}

		if err := adapter.FeedsAdapter(ctx, tFeeds, task.FeedsTask, &res); err != nil {
			t.Fatal(err)
		}

		if res.Code != task.expected {
			t.Fail()
		}
	}
}
