package recipes

import (
	"encoding/json"
	"time"

	"github.com/booleanism/tetek/feeds/internal/model"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/google/uuid"
)

type GetFeedsRequest struct {
	Id     uuid.UUID `query:"id"`
	Offset int       `query:"offset"`
	Type   string    `query:"type"` // M, J, A, S
}

type GetFeedsResponse struct {
	helper.GenericResponse
	Detail []model.Feed `json:"detail"`
}

func (r GetFeedsResponse) Json() []byte {
	j, _ := json.Marshal(r)
	return j
}

type HideRequest struct {
	Id uuid.UUID `json:"id"`
}

type HideResponse struct {
	helper.GenericResponse
	Detail HideRequest `json:"detail"`
}

type NewFeedRequest struct {
	Title string `json:"title"`
	Url   string `json:"url"`
	Text  string `json:"text"`
	Type  string `json:"type"`
}

type NewFeedResponse struct {
	helper.GenericResponse
	Detail NewFeedRequest `json:"detail"`
}

func (fr NewFeedRequest) ToFeed() model.Feed {
	now := time.Now()
	return model.Feed{
		Id:         uuid.New(),
		Title:      fr.Title,
		Url:        fr.Url,
		Text:       fr.Text,
		Type:       fr.Type,
		Points:     0,
		NCommnents: 0,
		Created_At: &now,
	}
}

func (fr NewFeedResponse) Json() []byte {
	j, _ := json.Marshal(fr)
	return j
}

type DeleteRequest struct {
	Id uuid.UUID `uri:"id"`
}

type DeleteResponse struct {
	helper.GenericResponse
	Detail DeleteRequest `json:"detail"`
}
