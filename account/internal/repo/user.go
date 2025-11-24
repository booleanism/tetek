package repo

import (
	"context"

	"github.com/booleanism/tetek/account/internal/model"
	"github.com/booleanism/tetek/db"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/jackc/pgx/v5"
)

type UserRepo interface {
	GetUser(context.Context, **model.User) error
	NewUser(context.Context, **model.User) error
}

type userRepo struct {
	pool db.Acquireable
}

func NewUserRepo(pool db.Acquireable) *userRepo {
	return &userRepo{pool}
}

func (r *userRepo) GetUser(ctx context.Context, user **model.User) error {
	ctx, log := loggr.GetLogger(ctx, "getUser-repo")
	corrID := ctx.Value(keystore.RequestID{}).(string)
	c, err := r.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer c.Release()

	err = c.QueryRow(
		ctx,
		"SELECT id, uname, email, passwd, role, created_at FROM users WHERE uname = $1 OR email = $2;",
		(*user).Uname, (*user).Email,
	).Scan(&(*user).ID, &(*user).Uname, &(*user).Email, &(*user).Passwd, &(*user).Role, &(*user).CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.V(1).Info("no user", "requestID", corrID, "user", user)
			return err
		}

		log.V(1).Info("failed to scan user row", "requestID", corrID, "error", err, "user", user)
		return err
	}

	return nil
}

func (r *userRepo) NewUser(ctx context.Context, user **model.User) error {
	ctx, log := loggr.GetLogger(ctx, "newUser-repo")
	corrID := ctx.Value(keystore.RequestID{}).(string)
	c, err := r.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer c.Release()

	err = c.QueryRow(
		ctx,
		"INSERT INTO users (id, uname, email, passwd, role, created_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;",
		(*user).ID, (*user).Uname, (*user).Email, (*user).Passwd, (*user).Role, (*user).CreatedAt,
	).Scan(&(*user).ID, &(*user).Uname, &(*user).Email, &(*user).Passwd, &(*user).Role, &(*user).CreatedAt)
	if err != nil {
		log.V(1).Info("failed to insert user", "requestID", corrID, "error", err, "user", user)
		return err
	}

	return nil
}
