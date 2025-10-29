package repo

import (
	"context"

	"github.com/booleanism/tetek/account/internal/model"
	"github.com/booleanism/tetek/db"
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

func (r *userRepo) GetUser(ctx context.Context, user model.User) (u model.User, err error) {
	c, err := r.pool.Acquire(ctx)
	if err != nil {
		return
	}
	defer c.Release()

	u = model.User{}
	err = c.QueryRow(
		ctx,
		"SELECT id, uname, email, passwd, role, created_at FROM users WHERE uname = $1 OR email = $2;",
		user.Uname, user.Email,
	).Scan(&u.Id, &u.Uname, &u.Email, &u.Passwd, &u.Role, &u.CreatedAt)

	return
}

func (r *userRepo) NewUser(ctx context.Context, user model.User) (u model.User, err error) {
	c, err := r.pool.Acquire(ctx)
	if err != nil {
		return
	}
	defer c.Release()

	u = model.User{}
	err = c.QueryRow(
		ctx,
		"INSERT INTO users (id, uname, email, passwd, role, created_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;",
		user.Id, user.Uname, user.Email, user.Passwd, user.Role, user.CreatedAt,
	).Scan(&u.Id, &u.Uname, &u.Email, &u.Passwd, &u.Role, &u.CreatedAt)

	return
}
