package messaging_test

import (
	"context"
	"testing"

	mqAcc "github.com/booleanism/tetek/account/infra/messaging/rabbitmq"
	messaging "github.com/booleanism/tetek/auth/infra/messaging/rabbitmq"
	contract "github.com/booleanism/tetek/auth/internal/infra/messaging/rabbitmq"
	"github.com/booleanism/tetek/auth/internal/usecases/jwt"
	"github.com/booleanism/tetek/pkg/contracts"
	"github.com/booleanism/tetek/pkg/contracts/adapter"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/rabbitmq/amqp091-go"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

type mqCon struct {
	*rabbitmq.RabbitMQContainer
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
}

func TestMain(m *testing.M) {
	m.Run()
	if err := container.RabbitMQContainer.Terminate(context.Background()); err != nil {
		panic(err)
	}
}

type task struct {
	messaging.AuthTask
	expected int
}

func TestWorker(t *testing.T) {
	con, err := amqp091.Dial(container.mqConStr)
	if err != nil {
		t.Fatal(err)
	}

	j := jwt.NewJwt([]byte("test"))

	acc := contract.NewAuth(con, j)
	ch, err := acc.WorkerAuthListener(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := ch.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	tAuth := contracts.AuthAssent(con)
	if err := tAuth.AuthResListener(context.Background(), "test"); err != nil {
		t.Fatal(err)
	}

	jPass, err := j.Generate(mqAcc.User{Uname: "root", Role: "N"})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.WithValue(context.Background(), keystore.RequestID{}, "test")

	res := &messaging.AuthResult{}
	tasks := []task{
		{
			AuthTask: messaging.AuthTask{},
			expected: errro.ErrAuthJWTVerifyFail,
		},
		{
			AuthTask: messaging.AuthTask{Jwt: jPass},
			expected: errro.Success,
		},
	}

	for _, task := range tasks {
		if err := adapter.AuthAdapter(ctx, tAuth, task.AuthTask, &res); err != nil {
			t.Fatal(err)
		}

		if res.Code != task.expected {
			t.Fail()
		}
	}
}
