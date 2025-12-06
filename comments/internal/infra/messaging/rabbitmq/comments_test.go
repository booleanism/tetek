package messaging_test

import (
	"context"
	"testing"
	"time"

	messaging "github.com/booleanism/tetek/comments/infra/messaging/rabbitmq"
	imessaging "github.com/booleanism/tetek/comments/internal/infra/messaging/rabbitmq"
	"github.com/booleanism/tetek/comments/internal/usecases"
	"github.com/booleanism/tetek/comments/internal/usecases/repo"
	"github.com/booleanism/tetek/comments/schemas"
	db "github.com/booleanism/tetek/infra/db/sql"
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
	messaging.CommentsTask
	expected int
}

func TestWorker(t *testing.T) {
	p := db.Register(container.dbConStr)
	defer p.Close()

	mqCon, err := amqp091.Dial(container.mqConStr)
	if err != nil {
		t.Fatal(err)
	}

	repo := repo.NewCommentsRepo(p)
	uc := usecases.NewCommentsUsecases(repo)

	ctx := context.WithValue(context.Background(), keystore.RequestID{}, "test")

	contr := imessaging.NewComments(mqCon, uc)
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
			CommentsTask: messaging.CommentsTask{
				Cmd: 1,
				Comment: messaging.Comment{
					Parent: parentPass,
				},
			},
			expected: errro.ErrCommUnknownCmd,
		},
		{
			CommentsTask: messaging.CommentsTask{
				Cmd: 0,
				Comment: messaging.Comment{
					Parent: parentPass,
				},
			},
			expected: errro.Success,
		},
		{
			CommentsTask: messaging.CommentsTask{
				Cmd:     0,
				Comment: messaging.Comment{Parent: uuid.New()},
			},
			expected: errro.ErrCommNoComments,
		},
		{
			CommentsTask: messaging.CommentsTask{
				Cmd:     0,
				Comment: messaging.Comment{},
			},
			expected: errro.ErrCommMissingRequiredField,
		},
		{
			CommentsTask: messaging.CommentsTask{
				Cmd:     0,
				Comment: messaging.Comment{Parent: uuid.New()},
			},
			expected: errro.ErrCommDBError,
		},
	}

	res := &messaging.CommentsResult{}
	assent := contracts.CommentsAssent(mqCon)
	if err := assent.CommentsResListener(ctx, "test"); err != nil {
		t.Fatal(err)
	}

	for _, v := range data {
		if v.expected == errro.ErrCommDBError {
			p.Close()
		}

		if err := adapter.CommentsAdapter(ctx, assent, v.CommentsTask, &res); err != nil {
			t.Fatal(err)
		}

		t.Log(res)

		if v.expected != res.Code {
			t.Fail()
		}
	}
}
