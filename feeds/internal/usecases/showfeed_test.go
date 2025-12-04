//go:build ignore

// This test are broken because coupled with comments worker.
// For now, it's not possible to runs comments worker outside of service it serlf.
package usecases_test

import (
	"context"
	"testing"

	"github.com/booleanism/tetek/feeds/internal/domain/model"
	"github.com/booleanism/tetek/feeds/internal/domain/repo"
	"github.com/booleanism/tetek/feeds/internal/usecases"
	db "github.com/booleanism/tetek/infra/db/sql"
	"github.com/booleanism/tetek/pkg/contracts"
	"github.com/booleanism/tetek/pkg/contracts/adapter"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

type testShowFeed struct {
	sfr      usecases.ShowFeedRequest
	expected int
}

func TestShowFeed(t *testing.T) {
	d := db.Register(container.dbConStr)
	m, err := amqp091.Dial(container.mqConStr)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()

	ctx := context.Background()
	ctx = context.WithValue(ctx, keystore.RequestID{}, "test")

	c := contracts.CommentsAssent(m)
	if err := c.CommentsResListener(ctx, "test"); err != nil {
		t.Fatal(err)
	}

	r := repo.NewFeedsRepo(d)
	uc := usecases.NewFeedsUsecase(r)

	existFeedID, err := uuid.Parse("14ed0f31-7a38-4c0b-ac48-8714378810f0")
	if err != nil {
		t.Fatal(err)
	}

	data := []testShowFeed{
		{
			sfr:      usecases.ShowFeedRequest{ID: existFeedID, User: usecases.User{Uname: "root"}},
			expected: errro.Success,
		},
		{
			sfr:      usecases.ShowFeedRequest{ID: uuid.New(), User: usecases.User{Uname: "root"}},
			expected: errro.ErrFeedsNoFeeds,
		},
		{
			sfr:      usecases.ShowFeedRequest{ID: uuid.New()},
			expected: errro.ErrFeedsMissingRequiredField,
		},
		{
			sfr:      usecases.ShowFeedRequest{},
			expected: errro.ErrFeedsMissingRequiredField,
		},
	}

	buf := &model.FeedWithComments{}

	for _, v := range data {
		err := uc.ShowFeed(ctx, c, adapter.CommentsAdapter, v.sfr, &buf)
		if err == nil {
			continue
		}

		if err.Code() != v.expected {
			t.Fail()
		}
	}
}
