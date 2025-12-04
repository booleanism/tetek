package usecases_test

import (
	"context"
	"testing"

	"github.com/booleanism/tetek/feeds/internal/usecases"
	"github.com/booleanism/tetek/feeds/internal/usecases/dto"
	"github.com/booleanism/tetek/feeds/internal/usecases/repo"
	db "github.com/booleanism/tetek/infra/db/sql"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/google/uuid"
)

type testNewFeed struct {
	nfr      dto.NewFeedRequest
	expected int
}

func TestNewFeed(t *testing.T) {
	d := db.Register(container.dbConStr)
	r := repo.NewFeedsRepo(d)
	uc := usecases.NewFeedsUsecase(r)

	existFeedID, err := uuid.Parse("b577c5cc-3dcd-42e8-ba21-387588dc6503")
	if err != nil {
		t.Fatal(err)
	}

	data := []testNewFeed{
		{
			nfr: dto.NewFeedRequest{
				ID:   uuid.New(),
				Type: "M",
				User: dto.User{Uname: "root"},
			},
			expected: errro.Success,
		},
		{
			nfr: dto.NewFeedRequest{
				ID:   existFeedID,
				Type: "M",
				User: dto.User{Uname: "root"},
			},
			expected: errro.ErrFeedsNewFail,
		},
		{
			nfr: dto.NewFeedRequest{
				ID:   existFeedID,
				Type: "M",
			},
			expected: errro.ErrFeedsMissingRequiredField,
		},
		{
			nfr:      dto.NewFeedRequest{},
			expected: errro.ErrFeedsMissingRequiredField,
		},
	}

	ctx := context.Background()

	for _, v := range data {
		err := uc.NewFeed(ctx, v.nfr)
		if err == nil {
			continue
		}

		if err.Code() != v.expected {
			t.Fail()
		}
	}
}
