package user

import (
	"context"

	"github.com/booleanism/tetek/account/internal/domain/entities"
	"github.com/booleanism/tetek/pkg/loggr"
)

func (ur userRepo) GetUser(ctx context.Context, uname string, buf **entities.User) (int, error) {
	ctx, log := loggr.GetLogger(ctx, "getUser-repo")

	query := `
	SELECT
		uname, email, passwd, role
	FROM users
	WHERE 
		uname = $1;`

	rws, err := ur.exec(ctx, ur.db, query, uname)
	if err != nil {
		log.Error(err, "failed to execute query", "query", query)
		return 0, err
	}

	u := &[]entities.User{}
	n, err := ur.scanFn(ctx, rws, u, "uname", uname)
	if err != nil {
		return 0, err
	}
	(*buf) = &(*u)[0]

	return n, nil
}
