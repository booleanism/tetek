package dto

import (
	"encoding/json"
	"time"

	"github.com/booleanism/tetek/comments/internal/internal/domain/entities"
	"github.com/booleanism/tetek/comments/internal/internal/domain/model"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/google/uuid"
)

type (
	Comment       = entities.Comment
	CommentFilter = model.CommentFilter
)

type GetCommentsRequest struct {
	Head   uuid.UUID `uri:"id"`
	Offset int       `query:"offset"`
	Limit  int       `query:"limit"`
}

type GetCommentsResponse struct {
	helper.GenericResponse
	Details []Comment `json:"details"`
}

type NewCommentRequest struct {
	ID   uuid.UUID `json:"id"`
	Head uuid.UUID `json:"head"`
	Text string    `json:"text"`
	By   string    `json:"by"`
}

type newCommentRequest struct {
	NewCommentRequest
	By string `json:"by"`
}

func (c newCommentRequest) toComment() Comment {
	now := time.Now()
	return Comment{
		ID:        uuid.New(),
		Parent:    c.Head,
		Text:      c.Text,
		By:        c.By,
		Points:    0,
		CreatedAt: &now,
		Child:     nil,
	}
}

type NewCommentResponse struct {
	helper.GenericResponse
	Detail Comment `json:"details"`
}

func (r NewCommentResponse) JSON() []byte {
	j, _ := json.Marshal(&r)
	return j
}
