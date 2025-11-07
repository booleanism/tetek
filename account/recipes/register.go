package recipes

import (
	"context"
	"errors"
	"time"

	"github.com/booleanism/tetek/account/internal/model"
	"github.com/booleanism/tetek/account/internal/repo"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/loggr"
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
		return loggr.LogError(func(z loggr.LogErr) errro.Error {
			e := errro.New(errro.EACCOUNT_INVALID_REGIST_PARAM, "uname, email or passwd should not empty")
			z.V(3).Error(e, "missing required user field", "user", user)
			return e
		})
	}

	id := user.Id
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
		return loggr.LogError(func(z loggr.LogErr) errro.Error {
			e := errro.New(errro.EACCOUNT_PASSWD_HASH_FAIL, "failed to hash passwd")
			z.V(0).Error(err, e.Error(), "passwd", user.Passwd)
			return e
		})
	}

	user = model.User{
		Id:        id,
		Uname:     user.Uname,
		Email:     user.Email,
		Passwd:    string(passwd),
		Role:      role,
		CreatedAt: at,
	}

	_, err = r.repo.NewUser(ctx, user)
	if err != nil {
		loggr.LogError(func(z loggr.LogErr) errro.Error {
			z.V(0).Error(err, "cannot creating new user", "user", user)
			return nil
		})
		pgErr := &pgconn.PgError{}
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return loggr.LogError(func(z loggr.LogErr) errro.Error {
					e := errro.New(errro.EACCOUNT_USER_ALREADY_EXIST, "user already exist")
					z.V(3).Error(pgErr, e.Error(), "user", user)
					return e
				})
			}
		}

		return loggr.LogError(func(z loggr.LogErr) errro.Error {
			e := errro.New(errro.EACCOUNT_DB_ERR, "error happen with database interaction")
			z.V(0).Error(err, e.Error(), "user", user)
			return e
		})
	}

	return nil
}
