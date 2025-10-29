package feeds

import (
	"time"

	"github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/feeds/internal/model"
	"github.com/booleanism/tetek/feeds/recipes"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/go-logr/logr"
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
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Detail  newFeedRequest `json:"detail"`
}

func (fr newFeedRequest) ToFeed() model.Feed {
	now := time.Now()
	return model.Feed{
		Id:         uuid.NewString(),
		Title:      fr.Title,
		Url:        fr.Url,
		Text:       fr.Text,
		Type:       fr.Type,
		Points:     0,
		NCommnents: 0,
		Created_At: &now,
	}
}

func New(rec recipes.FeedRecipes) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		loggr.Log.V(4).Info("new incoming feed request")
		req := newFeedRequest{}
		if res, err := helper.BindRequest(ctx, &req); err != nil {
			return loggr.Log.Error(3, func(z logr.LogSink) errro.Error {
				z.Error(err, res.Message)
				return errro.New(res.Code, res.Message)
			}).ToFiber()
		}

		jwt, ok := ctx.Locals("jwt").(*amqp.AuthResult)
		if !ok {
			res := newFeedResponse{
				Code:    errro.EAUTH_INVALID_AUTH_RESULT_TYPE,
				Message: "does not represent jwt type",
			}
			return loggr.Log.Error(2, func(z logr.LogSink) errro.Error {
				e := errro.FromError(res.Code, ctx.Status(fiber.StatusInternalServerError).JSON(&res))
				z.Error(e, res.Message)
				return e
			}).ToFiber()
		}

		f := req.ToFeed()
		err := rec.New(ctx, f, jwt)
		if err == nil {
			res := newFeedResponse{
				Code:    errro.SUCCESS,
				Message: "success add new feed",
				Detail:  req,
			}
			loggr.Log.V(4).Info("new feed added")
			return ctx.Status(fiber.StatusOK).JSON(&res)
		}

		res := newFeedResponse{
			Code:    errro.EFEEDS_NEW_FAIL,
			Message: "failed to create new feed",
			Detail:  req,
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(&res)
	}
}
