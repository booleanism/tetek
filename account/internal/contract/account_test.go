package contract_test

import (
	"context"
	"testing"
	"time"

	"github.com/booleanism/tetek/account/amqp"
	"github.com/booleanism/tetek/account/internal/contract"
	"github.com/booleanism/tetek/account/internal/repo"
	"github.com/booleanism/tetek/account/schemas"
	"github.com/booleanism/tetek/db"
	"github.com/booleanism/tetek/pkg/contracts"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/rabbitmq/amqp091-go"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"github.com/testcontainers/testcontainers-go/wait"
)

type mqCon struct {
	*rabbitmq.RabbitMQContainer
	*postgres.PostgresContainer
	dbConStr string
	mqConStr string
}

var container = &mqCon{}

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
		postgres.WithDatabase("accounttestdb"),
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
		"psql -U postgres -d accounttestdb <<'EOF'\n" + schemas.AccountSQL + "\nEOF",
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

type task struct {
	amqp.AccountTask
	expected int
}

func TestWorker(t *testing.T) {
	con, err := amqp091.Dial(container.mqConStr)
	if err != nil {
		t.Fatal(err)
	}

	p := db.Register(container.dbConStr)
	defer p.Close()

	repo := repo.NewUserRepo(p)

	acc := contract.NewAccount(con, repo)
	ch, err := acc.WorkerAccountListener(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := ch.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	tAccount := contracts.AccountAssent(con)
	if err := tAccount.AccountResListener(context.Background(), "test"); err != nil {
		t.Fatal(err)
	}

	ctx := context.WithValue(context.Background(), keystore.RequestID{}, "test")

	res := &amqp.AccountResult{}
	tasks := []task{
		{
			AccountTask: amqp.AccountTask{
				Cmd: 0, User: amqp.User{Uname: "root"},
			},
			expected: errro.Success,
		},
		{
			AccountTask: amqp.AccountTask{
				Cmd: 1, User: amqp.User{Uname: "root"},
			},
			expected: errro.ErrAccountUnknownCmd,
		},
		{
			AccountTask: amqp.AccountTask{
				Cmd: 0, User: amqp.User{Uname: "nouser"},
			},
			expected: errro.ErrAccountNoUser,
		},
		{
			AccountTask: amqp.AccountTask{
				Cmd: 0, User: amqp.User{Uname: "root"},
			},
			expected: errro.ErrAccountDBError,
		},
	}

	for _, task := range tasks {
		if task.expected == errro.ErrAccountDBError {
			(*p).Close()
		}

		if err := contracts.AccAdapter(ctx, tAccount, task.AccountTask, &res); err != nil {
			t.Fatal(err)
		}

		if res.Code != task.expected {
			t.Fail()
		}
	}
}
