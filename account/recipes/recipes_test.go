package recipes_test

import (
	"context"
	"errors"
	"testing"

	"github.com/booleanism/tetek/account/internal/model"
	"github.com/booleanism/tetek/account/recipes"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type userRepoTest struct {
	recipes.ProfileRequest
	recipes.RegistRequest
	expected int
	userRepo
}

type userRepo struct {
	fail error
}

func (r userRepo) GetUser(ctx context.Context, user **model.User) error {
	return r.fail
}

func (r userRepo) NewUser(ctx context.Context, user **model.User) error {
	return r.fail
}

func TestProfile(t *testing.T) {
	data := []userRepoTest{
		{ProfileRequest: recipes.ProfileRequest{Uname: ""}, expected: errro.ErrAccountEmptyParam, userRepo: userRepo{nil}},
		{ProfileRequest: recipes.ProfileRequest{Uname: "test"}, expected: errro.Success},
		{ProfileRequest: recipes.ProfileRequest{Uname: "test"}, expected: errro.ErrAccountDBError, userRepo: userRepo{errors.New("something wrong")}},
		{ProfileRequest: recipes.ProfileRequest{Uname: "test"}, expected: errro.ErrAccountNoUser, userRepo: userRepo{pgx.ErrNoRows}},
	}

	for _, v := range data {
		rec := recipes.New(v.userRepo)
		u := &model.User{}
		err := rec.Profile(context.Background(), v.ProfileRequest, &u)

		code := 0
		if err != nil {
			code = err.Code()
		}

		if code != v.expected {
			t.Fail()
		}
	}
}

func TestRegister(t *testing.T) {
	data := []userRepoTest{
		{RegistRequest: recipes.RegistRequest{}, expected: errro.ErrAccountInvalidRegistParam, userRepo: userRepo{nil}},
		{RegistRequest: recipes.RegistRequest{Uname: "test", Email: "test@test", Passwd: "test"}, expected: errro.ErrAccountUserAlreadyExist, userRepo: userRepo{&pgconn.PgError{Code: "23505"}}},
		{RegistRequest: recipes.RegistRequest{Uname: "test", Email: "test@test", Passwd: "test"}, expected: errro.ErrAccountDBError, userRepo: userRepo{errors.New("something wrong")}},
		{RegistRequest: recipes.RegistRequest{Uname: "test", Email: "test@test", Passwd: "test"}, expected: errro.Success, userRepo: userRepo{nil}},
	}

	for _, v := range data {
		rec := recipes.New(v.userRepo)
		err := rec.Regist(context.Background(), v.RegistRequest)

		code := 0
		if err != nil {
			code = err.Code()
		}

		if code != v.expected {
			t.Fail()
		}
	}
}
