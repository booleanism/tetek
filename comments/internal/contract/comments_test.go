package contract_test

import (
	"context"
	"testing"
	"time"

	"github.com/booleanism/tetek/comments/amqp"
	"github.com/booleanism/tetek/comments/internal/contract"
	"github.com/booleanism/tetek/comments/internal/repo"
	"github.com/booleanism/tetek/comments/schemas"
	"github.com/booleanism/tetek/db"
	"github.com/booleanism/tetek/pkg/contracts"
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
		postgres.WithDatabase("commentstestdb"),
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
		"psql -U postgres -d commentstestdb <<'EOF'\n" + schemas.CommentsSQL + "\nEOF",
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
	amqp.CommentsTask
	expected int
}

func TestWorker(t *testing.T) {
	db.Register(container.dbConStr)
	p := db.GetPool()
	defer p.Close()

	mqCon, err := amqp091.Dial(container.mqConStr)
	if err != nil {
		t.Fatal(err)
	}

	repo := repo.NewCommRepo(p)

	ctx := context.WithValue(context.Background(), keystore.RequestID{}, "test")

	contr := contract.NewComments(mqCon, repo)
	ch, err := contr.WorkerCommentsListener(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := ch.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	parentPass, err := uuid.Parse("14ed0f31-7a38-4c0b-ac48-8714378810f0")
	if err != nil {
		t.Fatal(err)
	}

	data := []testData{
		{
			CommentsTask: amqp.CommentsTask{
				Cmd: 1,
				Comment: amqp.Comment{
					Parent: parentPass,
				},
			},
			expected: errro.ErrCommUnknownCmd,
		},
		{
			CommentsTask: amqp.CommentsTask{
				Cmd: 0,
				Comment: amqp.Comment{
					Parent: parentPass,
				},
			},
			expected: errro.Success,
		},
		{
			CommentsTask: amqp.CommentsTask{
				Cmd:     0,
				Comment: amqp.Comment{},
			},
			expected: errro.ErrCommNoComments,
		},
		{
			CommentsTask: amqp.CommentsTask{
				Cmd:     0,
				Comment: amqp.Comment{},
			},
			expected: errro.ErrCommDBError,
		},
	}

	res := &amqp.CommentsResult{}
	assent := contracts.CommentsAssent(mqCon)
	if err := assent.CommentsResListener(ctx, "test"); err != nil {
		t.Fatal(err)
	}

	for _, v := range data {
		if v.expected == errro.ErrCommDBError {
			p.Close()
		}

		if err := contracts.CommentsAdapter(ctx, assent, v.CommentsTask, &res); err != nil {
			t.Fatal(err)
		}

		t.Log(res)

		if v.expected != res.Code {
			t.Fail()
		}
	}
}
