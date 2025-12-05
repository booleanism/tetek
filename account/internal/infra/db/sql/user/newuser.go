package user

import (
	"context"

	"github.com/booleanism/tetek/account/internal/domain/entities"
	"github.com/booleanism/tetek/pkg/loggr"
)

func (ur userRepo) PutUser(ctx context.Context, user **entities.User) error {
	ctx, _ = loggr.GetLogger(ctx, "putUser-repo")

	query := `
	INSERT INTO users
		(id, uname, email, passwd, role, created_at)
	VALUES
		($1, $2, $3, $4, $5, $6)
	RETURNING
		uname, email, passwd, role`

	rws, err := ur.exec(ctx, ur.db, query, &(*user).ID, &(*user).Uname, &(*user).Email, &(*user).Passwd, &(*user).Role, &(*user).CreatedAt)
	if err != nil {
		return err
	}

	buf := &[]entities.User{}
	_, err = ur.scanFn(ctx, rws, buf)
	if err != nil {
		return err
	}
	(*user) = &(*buf)[0]

	return err
}
