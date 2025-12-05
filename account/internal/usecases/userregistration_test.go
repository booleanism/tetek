package usecases_test

import (
	"context"
	"testing"

	"github.com/booleanism/tetek/account/internal/usecases"
	"github.com/booleanism/tetek/account/internal/usecases/dto"
	"github.com/booleanism/tetek/account/internal/usecases/repo"
	db "github.com/booleanism/tetek/infra/db/sql"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/google/uuid"
)

type testUserRegistration struct {
	rr       dto.RegistRequest
	expected int
}

func TestUserRegistration(t *testing.T) {
	d := db.Register(pgContainer.ConnectionString)
	r := repo.NewUserRepo(d)
	uc := usecases.NewAccountUseCases(r)

	data := []testUserRegistration{
		{
			rr:       dto.RegistRequest{ID: uuid.New(), Uname: "test", Email: "test@email.com", Passwd: "test123"},
			expected: errro.Success,
		},
		{
			rr:       dto.RegistRequest{ID: uuid.New(), Uname: "root", Email: "root@email.com", Passwd: "root123"},
			expected: errro.ErrAccountRegistFail,
		},
		{
			rr:       dto.RegistRequest{ID: uuid.New(), Uname: "root", Email: "root@email.com", Passwd: "root"},
			expected: errro.ErrAccountPasswdLength,
		},
		{
			rr:       dto.RegistRequest{ID: uuid.New(), Uname: "root", Email: "root@email.com"},
			expected: errro.ErrAccountInvalidRegistParam,
		},
		{
			rr:       dto.RegistRequest{Uname: "root", Email: "root@email.com"},
			expected: errro.ErrAccountInvalidRegistParam,
		},
		{
			rr:       dto.RegistRequest{Uname: "root"},
			expected: errro.ErrAccountInvalidRegistParam,
		},
		{
			rr:       dto.RegistRequest{},
			expected: errro.ErrAccountInvalidRegistParam,
		},
	}

	ctx := context.Background()

	for _, v := range data {
		err := uc.RegistUser(ctx, v.rr)
		if err == nil {
			continue
		}

		if err.Code() != v.expected {
			t.Fail()
		}
	}
}
