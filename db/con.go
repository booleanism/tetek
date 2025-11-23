package db

import (
	"context"

	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Acquireable interface {
	Acquire(context.Context) (*pgxpool.Conn, error)
	Close()
}

type pool struct {
	*pgxpool.Pool
}

var pl *pool

func Register(cs string) {
	p, err := pgxpool.New(context.Background(), cs)
	if err != nil {
		panic(err)
	}

	err = p.Ping(context.Background())
	if err != nil {
		panic(err)
	}

	pl = &pool{p}
}

func (p *pool) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	pl, err := p.Pool.Acquire(ctx)
	if err != nil {
		ctx, log := loggr.GetLogger(ctx, "acquire-db")
		corrID := ctx.Value(keystore.RequestID{}).(string)
		e := errro.FromError(errro.ErrCommDBError, "failed to acquire db pool", err)
		log.Error(err, "failed to acquire db pool", "requestID", corrID)
		return nil, e
	}
	return pl, nil
}

func GetPool() Acquireable {
	if pl == nil {
		panic("Poor pool, not initialized. Consider calling `Register` first.")
	}

	return pl
}
