package usecases_test

import (
	"context"
	"testing"

	"github.com/booleanism/tetek/feeds/internal/domain/model/pools"
	"github.com/booleanism/tetek/feeds/internal/usecases"
	"github.com/booleanism/tetek/feeds/internal/usecases/dto"
	"github.com/booleanism/tetek/feeds/internal/usecases/repo"
	db "github.com/booleanism/tetek/infra/db/sql"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/google/uuid"
)

type testGetFeeds struct {
	gfr      dto.GetFeedsRequest
	expected int
	n        int
}

func TestGetFeeds(t *testing.T) {
	d := db.Register(container.dbConStr)
	r := repo.NewFeedsRepo(d)
	uc := usecases.NewFeedsUsecase(r)

	fPool := pools.FeedsPool.Get().(*pools.Feeds)
	defer pools.FeedsPool.Put(fPool)
	defer fPool.Reset()

	fIDPass, err := uuid.Parse("e13cfca7-6bb0-4f67-ab96-e8941c20911a")
	if err != nil {
		t.Fatal(err)
	}

	data := []testGetFeeds{
		{
			gfr:      dto.GetFeedsRequest{User: dto.User{Uname: "root"}, Type: "M"},
			expected: errro.Success,
			n:        1,
		},
		{
			gfr:      dto.GetFeedsRequest{ID: fIDPass},
			expected: errro.Success,
			n:        1,
		},
		{
			gfr:      dto.GetFeedsRequest{ID: uuid.New()},
			expected: errro.ErrFeedsNoFeeds,
		},
		{
			gfr:      dto.GetFeedsRequest{},
			expected: errro.ErrFeedsMissingRequiredField,
		},
	}

	ctx := context.Background()

	for _, v := range data {
		n, err := uc.GetFeeds(ctx, v.gfr, &fPool)
		if err == nil {
			if n != v.n {
				t.Errorf("expected %d, actual %d", v.n, n)
			}
			continue
		}

		if err.Code() != v.expected {
			t.Fail()
		}
	}
}
