package usecases_test

import (
	"context"
	"testing"

	"github.com/booleanism/tetek/comments/internal/usecases"
	"github.com/booleanism/tetek/comments/internal/usecases/dto"
	"github.com/booleanism/tetek/comments/internal/usecases/repo"
	db "github.com/booleanism/tetek/infra/db/sql"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/google/uuid"
)

type testGetComments struct {
	gcr    dto.GetCommentsRequest
	expect int
}

func TestGetComments(t *testing.T) {
	d := db.Register(pgContainer.ConnectionString)
	r := repo.NewCommentsRepo(d)
	uc := usecases.NewCommentsUsecases(r)

	existID, err := uuid.Parse("07099af9-3c49-4f55-85e7-54dfc11ea138")
	if err != nil {
		t.Fatal(err)
	}

	data := []testGetComments{
		{
			gcr:    dto.GetCommentsRequest{Head: existID},
			expect: errro.Success,
		},
		{
			gcr:    dto.GetCommentsRequest{Head: uuid.New()},
			expect: errro.ErrCommNoComments,
		},
		{
			gcr:    dto.GetCommentsRequest{},
			expect: errro.ErrCommMissingRequiredField,
		},
	}

	ctx := context.Background()
	buf := &[]dto.Comment{}

	for _, v := range data {
		_, err := uc.GetComments(ctx, v.gcr, buf)
		if err == nil {
			continue
		}

		if err.Code() != v.expect {
			t.Fail()
		}
	}
}
