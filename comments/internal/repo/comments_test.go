package repo_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/booleanism/tetek/comments/internal/model"
	"github.com/booleanism/tetek/comments/internal/repo"
	"github.com/booleanism/tetek/comments/schemas"
	"github.com/booleanism/tetek/db"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

type testData struct {
	repo.CommentFilter
	model.Comment
	expected error
	c        int
}

func TestGetComments(t *testing.T) {
	p := db.Register(pgContainer.ConnectionString)
	defer p.Close()

	headPass, err := uuid.Parse("14ed0f31-7a38-4c0b-ac48-8714378810f0")
	if err != nil {
		t.Fatal(err)
	}

	data := []testData{
		{
			CommentFilter: repo.CommentFilter{Head: headPass},
		},
		{
			CommentFilter: repo.CommentFilter{},
			expected:      pgx.ErrNoRows,
		},
		{
			CommentFilter: repo.CommentFilter{},
			c:             errro.ErrCommDBError,
		},
	}

	ctx := context.WithValue(context.Background(), keystore.RequestID{}, "test")
	res := []model.Comment{}

	for _, v := range data {
		if v.c == errro.ErrCommDBError {
			p.Close()
		}

		repo := repo.NewCommRepo(p)
		_, err := repo.GetComments(ctx, v.CommentFilter, &res)

		if v.expected != nil {
			if err != v.expected {
				t.Fail()
			}
		}

		var e *errro.Err
		if errors.As(err, &e) {
			if v.c != e.Code() {
				t.Fail()
			}
		}

	}
}

func TestNewComments(t *testing.T) {
	p := db.Register(pgContainer.ConnectionString)
	defer p.Close()

	now := time.Now()

	data := []testData{
		{
			Comment: model.Comment{},
			c:       errro.ErrCommInsertError,
		},
		{
			Comment: model.Comment{
				ID:        uuid.New(),
				Parent:    uuid.New(),
				Text:      "test",
				By:        "tester",
				CreatedAt: &now,
			},
		},
		{
			Comment: model.Comment{},
			c:       errro.ErrCommDBError,
		},
	}

	ctx := context.WithValue(context.Background(), keystore.RequestID{}, "test")

	for _, v := range data {
		if v.c == errro.ErrCommDBError {
			p.Close()
		}

		repo := repo.NewCommRepo(p)
		c := &v.Comment
		err := repo.NewComment(ctx, &c)

		if v.expected != nil {
			if err != v.expected {
				t.Fail()
			}
		}

		var e *errro.Err
		if errors.As(err, &e) {
			if v.c != e.Code() {
				t.Fail()
			}
		}
	}
}
