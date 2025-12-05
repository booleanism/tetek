package repo

import (
	"context"

	"github.com/booleanism/tetek/account/internal/domain/entities"
)

type UserGetter interface {
	GetUser(ctx context.Context, uname string, buf **entities.User) (n int, err error)
}

type UserAdder interface {
	PutUser(ctx context.Context, user **entities.User) error
}
