package router

import (
	"context"
	"encoding/json"
	"time"

	"github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/feeds/internal/model"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type newFeedRequest struct {
	Title string `json:"title"`
	Url   string `json:"url"`
	Text  string `json:"text"`
	Type  string `json:"type"`
}

type newFeedResponse struct {
	helper.GenericResponse
	Detail newFeedRequest `json:"detail"`
}

func (fr newFeedRequest) ToFeed() model.Feed {
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

func (fr newFeedResponse) Json() []byte {
	j, _ := json.Marshal(fr)
	return j
}

func (fr FeedsRouter) NewFeed(ctx fiber.Ctx) error {
	req := newFeedRequest{}
	if err := helper.BindRequest(ctx, &req); err != nil {
		return err.SendError(ctx, fiber.StatusBadRequest)
	}

	jwt, ok := ctx.Locals("jwt").(*amqp.AuthResult)
	if !ok {
		res := helper.GenericResponse{
			Code:    errro.EAUTH_INVALID_AUTH_RESULT_TYPE,
			Message: "does not represent jwt type",
		}
		e := errro.New(res.Code, res.Message).WithDetail(res.Json(), errro.TDETAIL_JSON)
		return e.SendError(ctx, fiber.StatusUnauthorized)
	}

	cto, cancel := context.WithTimeout(
		context.WithValue(
			context.Background(),
			helper.RequestIdKey{},
			ctx.Locals(helper.RequestIdKey{})),
		TIMEOUT*time.Second)
	defer cancel()

	f := req.ToFeed()
	err := fr.rec.New(cto, f, jwt)
	if err == nil {
		res := newFeedResponse{
			GenericResponse: helper.GenericResponse{
				Code:    errro.SUCCESS,
				Message: "success add new feed",
			},
			Detail: req,
		}
		return ctx.Status(fiber.StatusCreated).JSON(&res)
	}

	res := newFeedResponse{
		GenericResponse: helper.GenericResponse{
			Code:    errro.EFEEDS_NEW_FAIL,
			Message: "failed to create new feed",
		},
		Detail: req,
	}
	e := errro.New(res.Code, res.Message).WithDetail(res.Json(), errro.TDETAIL_JSON)
	return e.SendError(ctx, fiber.StatusInternalServerError)
}
