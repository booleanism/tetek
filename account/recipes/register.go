package recipes

import (
	"context"
	"errors"
	"time"

	"github.com/booleanism/tetek/account/internal/model"
	"github.com/booleanism/tetek/account/internal/repo"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type RegistRecipes interface {
	Regist(context.Context, RegistRequest) errro.Error
}

type registRecipes struct {
	repo repo.UserRepo
}

func (r *registRecipes) Regist(ctx context.Context, req RegistRequest) errro.Error {
	ctx, log := loggr.GetLogger(ctx, "regist-recipes")

	user := &model.User{
		Uname:  req.Uname,
		Passwd: req.Passwd,
		Email:  req.Email,
	}

	if err := setupUser(log, &user); err != nil {
		return err
	}

	err := r.repo.NewUser(ctx, &user)
	if err != nil {
		pgErr := &pgconn.PgError{}
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				e := errro.FromError(errro.ErrAccountUserAlreadyExist, "user already exist", pgErr)
				return e
			}
		}

		e := errro.FromError(errro.ErrAccountDBError, "error happen with our database", err)
		return e
	}

	return nil
}

func setupUser(log logr.Logger, user **model.User) errro.Error {
	if (*user).Uname == "" || (*user).Email == "" || (*user).Passwd == "" {
		e := errro.New(errro.ErrAccountInvalidRegistParam, "uname, email or passwd should not empty")
		log.V(1).Info("missing user property", "user", user)
		return e
	}

	id := (*user).ID
	if id == "" {
		id = uuid.NewString()
	}

	role := (*user).Role
	if role == "" {
		role = "N"
	}

	at := (*user).CreatedAt
	if (*user).CreatedAt == nil {
		now := time.Now()
		at = &now
	}

	passwd, err := bcrypt.GenerateFromPassword([]byte((*user).Passwd), bcrypt.DefaultCost)
	if err != nil {
		e := errro.FromError(errro.ErrAccountPasswdHashFail, "failed to hash passwd", err)
		log.V(1).Info("failed to hash passwd", "user", user, "error", err)
		return e
	}

	(*user).ID = id
	(*user).Passwd = string(passwd)
	(*user).Role = role
	(*user).CreatedAt = at

	return nil
}
