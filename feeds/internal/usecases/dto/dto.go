package dto

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/booleanism/tetek/feeds/internal/domain/entities"
	"github.com/booleanism/tetek/feeds/internal/domain/model"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/google/uuid"
)

type GetFeedsRequest struct {
	ID     uuid.UUID `query:"id"`
	Offset int       `query:"offset"`
	Type   string    `query:"type"` // M, J, A, S
	Limit  int
	User
}

type GetFeedsResponse struct {
	helper.GenericResponse
	Details []entities.Feed `json:"details"`
}

func (r GetFeedsResponse) JSON() []byte {
	j, _ := json.Marshal(r)
	return j
}

type ShowFeedRequest struct {
	ID uuid.UUID `uri:"id"`
	User
}

type ShowFeedResponse struct {
	helper.GenericResponse
	Detail model.FeedWithComments `json:"detail"`
}

type HideFeedRequest struct {
	ID     uuid.UUID `json:"id"`
	FeedID uuid.UUID `json:"feed_id"`
	User
}

type HideResponse struct {
	helper.GenericResponse
	Detail HideFeedRequest `json:"detail"`
}

type NewFeedRequest struct {
	ID    uuid.UUID `json:"id,omitempty"`
	Title string    `json:"title,omitempty"`
	URL   string    `json:"url,omitempty"`
	Text  string    `json:"text,omitempty"`
	Type  string    `json:"type,omitempty"`
	User
}

type NewFeedResponse struct {
	helper.GenericResponse
	Detail NewFeedRequest `json:"detail"`
}

func (fr NewFeedRequest) ToFeed() entities.Feed {
	now := time.Now()
	return entities.Feed{
		ID:         fr.ID,
		Title:      fr.Title,
		URL:        fr.URL,
		Text:       fr.Text,
		By:         fr.Uname,
		Type:       strings.ToUpper(fr.Type),
		Points:     0,
		NCommnents: 0,
		CreatedAt:  &now,
	}
}

func (fr NewFeedResponse) JSON() []byte {
	j, _ := json.Marshal(fr)
	return j
}

type User struct {
	Uname string
	Role  string
}

type DeleteFeedRequest struct {
	ID uuid.UUID `uri:"id"`
	User
}

type DeleteResponse struct {
	helper.GenericResponse
	Detail DeleteFeedRequest `json:"detail"`
}
