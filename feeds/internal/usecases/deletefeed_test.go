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

type testDeleteFeed struct {
	dfr      dto.DeleteFeedRequest
	expected int
}

func TestDeleteFeed(t *testing.T) {
	d := db.Register(container.dbConStr)
	r := repo.NewFeedsRepo(d)
	uc := usecases.NewFeedsUsecase(r)

	ctx := context.Background()

	existFeedID, err := uuid.Parse("cdff23df-b62c-446d-a2c1-b759fa342c5c")
	if err != nil {
		t.Fatal(err)
	}
	deletedFeedID := existFeedID

	data := []testDeleteFeed{
		{
			dfr:      dto.DeleteFeedRequest{ID: existFeedID, User: dto.User{Uname: "root"}},
			expected: errro.ErrFeedsUnathorized,
		},
		{
			dfr:      dto.DeleteFeedRequest{ID: existFeedID},
			expected: errro.ErrFeedsUnathorized,
		},
		{
			dfr:      dto.DeleteFeedRequest{ID: uuid.New()},
			expected: errro.ErrFeedsNoFeeds,
		},
		{
			dfr:      dto.DeleteFeedRequest{},
			expected: errro.ErrFeedsMissingRequiredField,
		},
		{
			dfr:      dto.DeleteFeedRequest{ID: existFeedID, User: dto.User{Uname: "rootz"}},
			expected: errro.Success,
		},
		{
			dfr:      dto.DeleteFeedRequest{ID: deletedFeedID, User: dto.User{Uname: "rootz", Role: "M"}},
			expected: errro.ErrFeedsNoFeeds,
		},
	}

	for _, v := range data {
		err := uc.DeleteFeed(ctx, v.dfr)
		if err == nil {
			continue
		}

		if err.Code() != v.expected {
			t.Fail()
		}
	}
}
