package recipes

import (
	"context"
	"errors"
	"time"

	"github.com/booleanism/tetek/account/internal/model"
	"github.com/booleanism/tetek/account/internal/repo"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type RegistRecipes interface {
	Regist(context.Context, model.User) errro.Error
}

type registRecipes struct {
	repo repo.UserRepo
}

func (r *registRecipes) Regist(ctx context.Context, user model.User) errro.Error {
	if user.Uname == "" || user.Email == "" || user.Passwd == "" {
		e := errro.New(errro.ErrAccountInvalidRegistParam, "uname, email or passwd should not empty")
		return e
	}

	id := user.ID
	if id == "" {
		id = uuid.NewString()
	}

	role := user.Role
	if role == "" {
		role = "N"
	}

	at := user.CreatedAt
	if user.CreatedAt == nil {
		now := time.Now()
		at = &now
	}

	passwd, err := bcrypt.GenerateFromPassword([]byte(user.Passwd), bcrypt.DefaultCost)
	if err != nil {
		e := errro.New(errro.ErrAccountPasswdHashFail, "failed to hash passwd")
		return e
	}

	user = model.User{
		ID:        id,
		Uname:     user.Uname,
		Email:     user.Email,
		Passwd:    string(passwd),
		Role:      role,
		CreatedAt: at,
	}

	_, err = r.repo.NewUser(ctx, user)
	if err != nil {
		pgErr := &pgconn.PgError{}
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				e := errro.New(errro.ErrAccountUserAlreadyExist, "user already exist")
				return e
			}
		}

		e := errro.New(errro.ErrAccountDBError, "error happen with database interaction")
		return e
	}

	return nil
}
