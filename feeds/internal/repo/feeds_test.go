package repo_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/booleanism/tetek/db"
	"github.com/booleanism/tetek/feeds/internal/model"
	"github.com/booleanism/tetek/feeds/internal/repo"
	"github.com/booleanism/tetek/feeds/schemas"
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

	conStr, err := pg.ConnectionString(ctx, "sslmode=disable")
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

	pgContainer.ConnectionString = conStr
	pgContainer.PostgresContainer = pg
}

func TestMain(m *testing.M) {
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
	Feed     []model.Feed
	Hf       *model.HiddenFeed
	ff       repo.FeedsFilter
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

func TestGetFeeds(t *testing.T) {
	p := db.Register(pgContainer.ConnectionString)
	defer p.Close()
	data := []repoData{
		{Feed: []model.Feed{}, ff: repo.FeedsFilter{}, expected: dberr{nil, ""}},
		{Feed: []model.Feed{}, ff: repo.FeedsFilter{Type: "M"}, expected: dberr{nil, ""}},
		{Feed: []model.Feed{}, ff: repo.FeedsFilter{HiddenTo: "rootz"}, expected: dberr{nil, ""}},
		{Feed: []model.Feed{}, ff: repo.FeedsFilter{Type: "M", HiddenTo: "rootz", Order: 0}, expected: dberr{nil, ""}},
		{Feed: []model.Feed{}, ff: repo.FeedsFilter{Type: "M", HiddenTo: "rootz", Order: 1}, expected: dberr{nil, ""}},
		{Feed: []model.Feed{}, ff: repo.FeedsFilter{Type: "M", HiddenTo: "rootz", Order: 0, Offset: 0}, expected: dberr{nil, ""}},
		{Feed: []model.Feed{}, ff: repo.FeedsFilter{Type: "M", HiddenTo: "rootz", Order: 0, Offset: 1}, expected: dberr{pgx.ErrNoRows, ""}},
		{Feed: []model.Feed{}, ff: repo.FeedsFilter{ID: uuid.New()}, expected: dberr{pgx.ErrNoRows, ""}}, // no such feed
	}

	sq := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	repo := repo.New(p, sq)
	for _, v := range data {
		err := repo.Feeds(context.Background(), v.ff, &v.Feed)

		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) {
			if err == nil {
				continue
			}

			if err == v.expected.Err() {
				continue
			}
		}

		if err != v.expected.Err() {
			t.Fail()
		}
	}
}

func TestNewFeeds(t *testing.T) {
	p := db.Register(pgContainer.ConnectionString)
	defer p.Close()
	idExist, err := uuid.Parse("cdff23df-b62c-446d-a2c1-b759fa342c5c")
	if err != nil {
		t.FailNow()
	}
	now := time.Now()

	data := []repoData{
		{Feed: []model.Feed{{}}, ff: repo.FeedsFilter{}, expected: dberr{nil, "23502"}},
		{Feed: []model.Feed{{ID: idExist, By: "root", URL: "localhost", Points: 0, NCommnents: 0, CreatedAt: &now}}, ff: repo.FeedsFilter{}, expected: dberr{nil, "23514"}},            // missing type check
		{Feed: []model.Feed{{ID: idExist, By: "root", URL: "localhost", Type: "M", Points: 0, NCommnents: 0, CreatedAt: &now}}, ff: repo.FeedsFilter{}, expected: dberr{nil, "23505"}}, // duplicate key
	}

	sq := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	repo := repo.New(p, sq)
	for _, v := range data {
		f := &v.Feed[0]
		err := repo.NewFeed(context.Background(), &f)

		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) {
			if err == nil {
				continue
			}

			if err == v.expected.Err() {
				continue
			}
		}

		if pgErr.Code != v.expected.Code() {
			t.Fail()
		}
	}
}

func TestDeleteFeeds(t *testing.T) {
	p := db.Register(pgContainer.ConnectionString)
	defer p.Close()
	idExist, err := uuid.Parse("cdff23df-b62c-446d-a2c1-b759fa342c5c")
	if err != nil {
		t.FailNow()
	}

	data := []repoData{
		{Feed: []model.Feed{{}}, ff: repo.FeedsFilter{}, expected: dberr{pgx.ErrNoRows, ""}},  // no such feed
		{Feed: []model.Feed{{}}, ff: repo.FeedsFilter{ID: idExist}, expected: dberr{nil, ""}}, // success
	}

	sq := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	repo := repo.New(p, sq)
	for _, v := range data {
		f := &v.Feed[0]
		err := repo.DeleteFeed(context.Background(), v.ff, &f)

		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) {
			if err == nil {
				continue
			}

			if err == v.expected.Err() {
				continue
			}
		}

		if pgErr.Code != v.expected.Code() {
			t.Fail()
		}
	}
}

func TestHideFeeds(t *testing.T) {
	p := db.Register(pgContainer.ConnectionString)
	defer p.Close()
	idExist, err := uuid.Parse("cdff23df-b62c-446d-a2c1-b759fa342c5c")
	if err != nil {
		t.FailNow()
	}

	data := []repoData{
		{Feed: []model.Feed{{}}, Hf: &model.HiddenFeed{}, ff: repo.FeedsFilter{}, expected: dberr{nil, "23503"}},           // no such feed
		{Feed: []model.Feed{{}}, Hf: &model.HiddenFeed{FeedID: idExist}, ff: repo.FeedsFilter{}, expected: dberr{nil, ""}}, // no such feed
	}

	sq := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	repo := repo.New(p, sq)
	for _, v := range data {
		err := repo.HideFeed(context.Background(), &v.Hf)

		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) {
			if err == nil {
				continue
			}

			if err == v.expected.Err() {
				continue
			}
		}

		if pgErr.Code != v.expected.Code() {
			t.Fail()
		}
	}
}
