package user

import (
	"github.com/booleanism/tetek/account/internal/usecases/dto"
	"github.com/booleanism/tetek/account/internal/usecases/repo/scanner"
	db "github.com/booleanism/tetek/infra/db/sql"
)

type userRepo struct {
	db     db.DB
	scanFn scanner.ScanFn[dto.User]
	exec   db.QueryExecutorFn
}

func NewUserRepo(db db.DB) userRepo {
	return userRepo{db, nil, nil}
}

func (ur userRepo) SetScanner(s scanner.ScanFn[dto.User]) userRepo {
	ur.scanFn = s
	return ur
}

func (ur userRepo) SetExecutor(exec db.QueryExecutorFn) userRepo {
	ur.exec = exec
	return ur
}
