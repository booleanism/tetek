package usecases_test

import (
	"context"
	"testing"
	"time"

	"github.com/booleanism/tetek/comments/schemas"
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

	conStr, err := pg.ConnectionString(ctx, "sslmode=disable")
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

	pgContainer.ConnectionString = conStr
	pgContainer.PostgresContainer = pg
}

func TestMain(m *testing.M) {
	m.Run()
	if err := pgContainer.Terminate(context.Background()); err != nil {
		panic(err)
	}
}
