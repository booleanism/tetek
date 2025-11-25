package repo_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/booleanism/tetek/account/internal/model"
	"github.com/booleanism/tetek/account/internal/repo"
	"github.com/booleanism/tetek/account/schemas"
	"github.com/booleanism/tetek/db"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	db.Register(pgContainer.ConnectionString)
	p := db.GetPool()
	defer p.Close()
	m.Run()
	if err := postgresContainer.Terminate(*pgContainer, context.Background()); err != nil {
		panic(err)
	}
}

type dbError interface {
	error
	Code() string
	Err() error
}

type repoData struct {
	*model.User
	expected dbError
}

type dberr struct {
	e error
	c string
}

func (e dberr) Error() string {
	return e.e.Error()
}

func (e dberr) Code() string {
	return e.c
}

func (e dberr) Err() error {
	return e.e
}

func TestGetUser(t *testing.T) {
	p := db.GetPool()
	repo := repo.NewUserRepo(p)
	data := []repoData{
		{User: &model.User{Uname: "root"}, expected: dberr{}},
		{User: &model.User{Email: "root@localhost"}, expected: dberr{}},
		{User: &model.User{Uname: "nouser"}, expected: dberr{pgx.ErrNoRows, "0"}},
	}

	ctx := context.WithValue(context.Background(), keystore.RequestID{}, "test")

	for _, v := range data {
		og := *v.User
		err := repo.GetUser(ctx, &v.User)
		if err != v.expected.Err() {
			t.Fail()
		}

		if og.Uname != v.Uname && og.Email != v.Email {
			t.Fail()
		}
	}
}

func TestNewUser(t *testing.T) {
	p := db.GetPool()
	repo := repo.NewUserRepo(p)
	now := time.Now()
	data := []repoData{
		{User: &model.User{Uname: "root"}, expected: dberr{nil, "23502"}},                                                                                      // invalid type ID
		{User: &model.User{ID: uuid.New()}, expected: dberr{nil, "23502"}},                                                                                     // missing non null property
		{User: &model.User{ID: uuid.New(), Uname: "root", Email: "root@localhost", Passwd: "test", Role: "N", CreatedAt: &now}, expected: dberr{nil, "23505"}}, // user already exist empty
	}

	ctx := context.WithValue(context.Background(), keystore.RequestID{}, "test")

	for _, v := range data {
		og := *v.User
		err := repo.NewUser(ctx, &v.User)

		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) {
			if err == nil {
				continue
			}
		}

		if pgErr.Code != v.expected.Code() {
			t.Fail()
		}

		if og.Uname != v.Uname && og.Email != v.Email {
			t.Fail()
		}
	}
}
