package recipes

import (
	"encoding/json"
	"time"

	"github.com/booleanism/tetek/comments/internal/model"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/google/uuid"
)

type GetCommentsRequest struct {
	Head   uuid.UUID `uri:"id"`
	Offset int       `query:"offset"`
}

type GetCommentsResponse struct {
	helper.GenericResponse
	Details []model.Comment `json:"details"`
}

type NewCommentRequest struct {
	Head uuid.UUID `json:"head"`
	Text string    `json:"text"`
}

type newCommentRequest struct {
	NewCommentRequest
	By string `json:"by"`
}

func (c newCommentRequest) toComment() model.Comment {
	now := time.Now()
	return model.Comment{
		Id:        uuid.New(),
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
	Detail model.Comment `json:"details"`
}

func (r NewCommentResponse) Json() []byte {
	j, _ := json.Marshal(&r)
	return j
}
