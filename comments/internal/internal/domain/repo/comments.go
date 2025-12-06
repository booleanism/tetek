package repo

import (
	"context"

	"github.com/booleanism/tetek/comments/internal/internal/domain/entities"
	"github.com/google/uuid"
)

type CommentsEntityRepo interface {
	CommentsAdder
	CommentsGetter
}

type CommentsGetter interface {
	GetComments(context.Context, uuid.UUID, int, int, *[]entities.Comment) (int, error)
}

type CommentsAdder interface {
	PutComment(context.Context, **entities.Comment) error
}
