package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/booleanism/tetek/comments/internal/usecases"
	"github.com/booleanism/tetek/comments/internal/usecases/dto"
	"github.com/booleanism/tetek/comments/internal/usecases/repo"
	messaging "github.com/booleanism/tetek/feeds/infra/messaging/rabbitmq"
	db "github.com/booleanism/tetek/infra/db/sql"
	"github.com/booleanism/tetek/pkg/contracts/adapter"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/google/uuid"
)

type mockFeedsDealer struct {
	consumeRet error
	res        messaging.FeedsResult
}

func (m mockFeedsDealer) Publish(ctx context.Context, task any) error {
	return nil
}

func (m mockFeedsDealer) Consume(ctx context.Context, result any) error {
	r, ok := result.(**messaging.FeedsResult)
	if !ok {
		return errors.New("oops")
	}
	(*r) = &m.res

	return m.consumeRet
}

func (m mockFeedsDealer) Name() string {
	return "feeds test"
}

type testNewComment struct {
	dealer   mockFeedsDealer
	ncr      dto.NewCommentRequest
	expected int
}

func TestNewComment(t *testing.T) {
	d := db.Register(pgContainer.ConnectionString)
	r := repo.NewCommentsRepo(d)
	uc := usecases.NewCommentsUsecases(r)

	existCommentID, err := uuid.Parse("07099af9-3c49-4f55-85e7-54dfc11ea138")
	if err != nil {
		t.Fatal(err)
	}

	data := []testNewComment{
		{
			dealer: mockFeedsDealer{
				res: messaging.FeedsResult{
					Code:    errro.ErrFeedsNoFeeds,
					Message: "no such feeds",
					Details: []messaging.Feeds{
						{ID: uuid.New()},
					},
				},
			},
			ncr:      dto.NewCommentRequest{ID: uuid.New(), Head: existCommentID, Text: "hello from test", By: "root"},
			expected: errro.Success,
		},
		{
			dealer: mockFeedsDealer{
				res: messaging.FeedsResult{
					Code:    errro.ErrFeedsNoFeeds,
					Message: "no such feeds",
					Details: []messaging.Feeds{
						{ID: uuid.New()},
					},
				},
			},
			ncr:      dto.NewCommentRequest{ID: uuid.New(), Head: uuid.New(), Text: "hello from test", By: "root"},
			expected: errro.ErrCommNoComments,
		},
		{
			dealer: mockFeedsDealer{
				res: messaging.FeedsResult{
					Code:    errro.Success,
					Message: "success",
					Details: []messaging.Feeds{
						{
							ID: uuid.New(),
						},
					},
				},
			},
			ncr:      dto.NewCommentRequest{ID: uuid.New(), Head: uuid.New(), Text: "hello from test", By: "root"},
			expected: errro.Success,
		},
		{
			dealer: mockFeedsDealer{
				res:        messaging.FeedsResult{},
				consumeRet: errors.New("failed consume"),
			},
			ncr:      dto.NewCommentRequest{ID: uuid.New(), Head: uuid.New(), Text: "hello from test", By: "root"},
			expected: errro.ErrServiceUnavailable,
		},
		{
			ncr:      dto.NewCommentRequest{ID: uuid.New(), Head: uuid.New(), By: "root"},
			expected: errro.ErrCommMissingRequiredField,
		},
		{
			ncr:      dto.NewCommentRequest{ID: uuid.New(), Head: uuid.New()},
			expected: errro.ErrCommMissingRequiredField,
		},
		{
			ncr:      dto.NewCommentRequest{ID: uuid.New()},
			expected: errro.ErrCommMissingRequiredField,
		},
		{
			ncr:      dto.NewCommentRequest{},
			expected: errro.ErrCommMissingRequiredField,
		},
	}

	ctx := context.Background()

	for _, v := range data {
		err := uc.NewComment(ctx, v.dealer, adapter.FeedsAdapter, v.ncr)

		if err != nil && v.expected != errro.Success {
			if err.Code() != v.expected {
				t.Fail()
			}
		}

	}
}
