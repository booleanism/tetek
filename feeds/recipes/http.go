package recipes

import (
	"encoding/json"
	"time"

	"github.com/booleanism/tetek/feeds/internal/model"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/google/uuid"
)

type GetFeedsRequest struct {
	ID     uuid.UUID `query:"id"`
	Offset int       `query:"offset"`
	Type   string    `query:"type"` // M, J, A, S
}

type GetFeedsResponse struct {
	helper.GenericResponse
	Details []model.Feed `json:"details"`
}

func (r GetFeedsResponse) JSON() []byte {
	j, _ := json.Marshal(r)
	return j
}

type ShowFeedRequest struct {
	ID uuid.UUID `uri:"id"`
}

type ShowFeedResponse struct {
	helper.GenericResponse
	Detail model.FeedWithComments `json:"detail"`
}

type HideRequest struct {
	ID uuid.UUID `json:"id"`
}

type HideResponse struct {
	helper.GenericResponse
	Detail HideRequest `json:"detail"`
}

type NewFeedRequest struct {
	Title string `json:"title,omitempty"`
	URL   string `json:"url,omitempty"`
	Text  string `json:"text,omitempty"`
	Type  string `json:"type,omitempty"`
}

type NewFeedResponse struct {
	helper.GenericResponse
	Detail NewFeedRequest `json:"detail"`
}

func (fr NewFeedRequest) ToFeed() model.Feed {
	now := time.Now()
	return model.Feed{
		ID:         uuid.New(),
		Title:      fr.Title,
		URL:        fr.URL,
		Text:       fr.Text,
		Type:       fr.Type,
		Points:     0,
		NCommnents: 0,
		CreatedAt:  &now,
	}
}

func (fr NewFeedResponse) JSON() []byte {
	j, _ := json.Marshal(fr)
	return j
}

type DeleteRequest struct {
	ID uuid.UUID `uri:"id"`
}

type DeleteResponse struct {
	helper.GenericResponse
	Detail DeleteRequest `json:"detail"`
}
