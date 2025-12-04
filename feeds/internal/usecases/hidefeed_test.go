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

type testHideFeed struct {
	hfr      dto.HideFeedRequest
	expected int
}

func TestHideFeed(t *testing.T) {
	d := db.Register(container.dbConStr)
	r := repo.NewFeedsRepo(d)
	uc := usecases.NewFeedsUsecase(r)

	existFeedIDByRoot, err := uuid.Parse("b577c5cc-3dcd-42e8-ba21-387588dc6503")
	if err != nil {
		t.Fatal(err)
	}

	data := []testHideFeed{
		{
			hfr:      dto.HideFeedRequest{ID: uuid.New(), FeedID: existFeedIDByRoot, User: dto.User{Uname: "rootz"}},
			expected: errro.Success,
		},
		{
			hfr:      dto.HideFeedRequest{FeedID: existFeedIDByRoot, User: dto.User{Uname: "root"}},
			expected: errro.ErrFeedsMissingRequiredField,
		},
		{
			hfr:      dto.HideFeedRequest{ID: uuid.New(), FeedID: uuid.New(), User: dto.User{Uname: "root"}},
			expected: errro.ErrFeedsNoFeeds,
		},
		{
			hfr:      dto.HideFeedRequest{FeedID: uuid.New()},
			expected: errro.ErrFeedsMissingRequiredField,
		},
		{
			hfr:      dto.HideFeedRequest{User: dto.User{Uname: "root"}},
			expected: errro.ErrFeedsMissingRequiredField,
		},
		{
			hfr:      dto.HideFeedRequest{},
			expected: errro.ErrFeedsMissingRequiredField,
		},
	}

	ctx := context.Background()

	for _, v := range data {
		err := uc.HideFeed(ctx, v.hfr)
		if err == nil {
			continue
		}

		if err.Code() != v.expected {
			t.Fail()
		}
	}
}
