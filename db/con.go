package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Acquireable interface {
	Acquire(context.Context) (*pgxpool.Conn, error)
}

var pool *pgxpool.Pool

func Register(cs string) {
	p, err := pgxpool.New(context.Background(), cs)
	if err != nil {
		panic(err)
	}

	err = p.Ping(context.Background())
	if err != nil {
		panic(err)
	}

	pool = p
}

func GetPool() *pgxpool.Pool {
	if pool == nil {
		panic("Poor pool, not initialized. Consider calling `Register` first.")
	}

	return pool
}
