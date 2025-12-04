package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

func Register(cs string) DB {
	p, err := pgxpool.New(context.Background(), cs)
	if err != nil {
		panic(err)
	}

	err = p.Ping(context.Background())
	if err != nil {
		panic(err)
	}

	return stdlib.OpenDBFromPool(p)
}
