package repo

import (
	"context"

	"github.com/booleanism/tetek/account/internal/model"
	"github.com/booleanism/tetek/db"
	"github.com/booleanism/tetek/pkg/loggr"
)

type UserRepo interface {
	GetUser(context.Context, model.User) (model.User, error)
	NewUser(context.Context, model.User) (model.User, error)
}

type userRepo struct {
	pool db.Acquireable
}

func NewUserRepo(pool db.Acquireable) *userRepo {
	return &userRepo{pool}
}

func (r *userRepo) GetUser(ctx context.Context, user model.User) (model.User, error) {
	ctx, log := loggr.GetLogger(ctx, "repo/GetUser")
	c, err := r.pool.Acquire(ctx)
	if err != nil {
		log.Error(err, "failed to acquire db pool")
		return model.User{}, err
	}
	defer c.Release()

	u := model.User{}
	err = c.QueryRow(
		ctx,
		"SELECT id, uname, email, passwd, role, created_at FROM users WHERE uname = $1 OR email = $2;",
		user.Uname, user.Email,
	).Scan(&u.ID, &u.Uname, &u.Email, &u.Passwd, &u.Role, &u.CreatedAt)
	if err != nil {
		log.Error(err, "failed to scan user row")
		return model.User{}, err
	}

	return u, nil
}

func (r *userRepo) NewUser(ctx context.Context, user model.User) (model.User, error) {
	ctx, log := loggr.GetLogger(ctx, "repo/NewUser")
	c, err := r.pool.Acquire(ctx)
	if err != nil {
		log.Error(err, "failed to acquire db pool")
		return model.User{}, err
	}
	defer c.Release()

	u := model.User{}
	err = c.QueryRow(
		ctx,
		"INSERT INTO users (id, uname, email, passwd, role, created_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;",
		user.ID, user.Uname, user.Email, user.Passwd, user.Role, user.CreatedAt,
	).Scan(&u.ID, &u.Uname, &u.Email, &u.Passwd, &u.Role, &u.CreatedAt)
	if err != nil {
		log.Error(err, "failed to scan user row")
		return model.User{}, err
	}

	return u, nil
}
