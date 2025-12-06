package usecases

import (
	"github.com/booleanism/tetek/comments/internal/usecases/repo"
)

type CommentsUseCases interface {
	GetCommentsUseCase
	NewCommentUseCase
}

type usecases struct {
	repo repo.CommentsRepo
}

func NewCommentsUsecases(repo repo.CommentsRepo) CommentsUseCases {
	return usecases{repo}
}
