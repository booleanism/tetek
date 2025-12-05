package usecases_test

import (
	"context"
	"testing"

	"github.com/booleanism/tetek/account/internal/usecases"
	"github.com/booleanism/tetek/account/internal/usecases/dto"
	"github.com/booleanism/tetek/account/internal/usecases/repo"
	db "github.com/booleanism/tetek/infra/db/sql"
	"github.com/booleanism/tetek/pkg/errro"
)

type testGetUserProfile struct {
	gur      dto.ProfileRequest
	expected int
}

func TestGetUserProfile(t *testing.T) {
	d := db.Register(pgContainer.ConnectionString)
	r := repo.NewUserRepo(d)
	uc := usecases.NewAccountUseCases(r)

	data := []testGetUserProfile{
		{
			gur:      dto.ProfileRequest{Uname: "root"},
			expected: errro.Success,
		},
		{
			gur:      dto.ProfileRequest{Uname: "test"},
			expected: errro.ErrAccountNoUser,
		},
		{
			gur:      dto.ProfileRequest{},
			expected: errro.ErrAccountEmptyParam,
		},
	}

	ctx := context.Background()

	for _, v := range data {
		u := &dto.User{}
		err := uc.GetProfile(ctx, v.gur, &u)
		if err == nil {
			continue
		}

		if err.Code() != v.expected {
			t.Fail()
		}
	}
}
