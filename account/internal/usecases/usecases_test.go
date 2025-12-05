package usecases_test

import (
	"context"
	"testing"
	"time"

	"github.com/booleanism/tetek/account/schemas"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type postgresContainer struct {
	*postgres.PostgresContainer
	ConnectionString string
}

var pgContainer = &postgresContainer{}

func init() {
	ctx := context.Background()
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

	conStr, err := pg.ConnectionString(ctx, "sslmode=disable")
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

	pgContainer.ConnectionString = conStr
	pgContainer.PostgresContainer = pg
}

func TestMain(m *testing.M) {
	m.Run()
	if err := postgresContainer.Terminate(*pgContainer, context.Background()); err != nil {
		panic(err)
	}
}
