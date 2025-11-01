package router

import (
	"encoding/json"

	"github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/feeds/internal/model"
	"github.com/booleanism/tetek/feeds/internal/repo"
	"github.com/booleanism/tetek/feeds/recipes"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/go-logr/logr"
	"github.com/gofiber/fiber/v3"
)

type getFeedsRequest struct {
	Offset int    `query:"offset"`
	Type   string `query:"type"` // M, J, A, S
}

type getFeedsResponse struct {
	helper.GenericResponse
	Detail []model.Feed `json:"detail"`
}

func (r *getFeedsResponse) Json() []byte {
	j, _ := json.Marshal(r)
	return j
}

type FeedsRouter struct {
	rec recipes.FeedRecipes
}

func NewFeedRouter(rec recipes.FeedRecipes) *FeedsRouter {
	return &FeedsRouter{rec}
}

func (fr *FeedsRouter) GetFeeds(ctx fiber.Ctx) error {
	loggr.Log.V(4).Info("new incomming feeds request")
	if ctx.IsMiddleware() {
		return ctx.Next()
	}

	req := getFeedsRequest{}
	if err := helper.BindRequest(ctx, &req); err != nil {
		return loggr.Log.ErrorRes(3, func(z logr.LogSink) error {
			z.Error(err, "failed to bind")
			return err.SendError(ctx, fiber.StatusBadRequest)
		})
	}

	jwt, _ := ctx.Locals("jwt").(*amqp.AuthResult)
	filter := repo.FeedsFilter{
		Offset: uint64(req.Offset),
		Type:   req.Type,
	}

	f, err := fr.rec.Feeds(ctx, filter, jwt)
	if err == nil {
		res := getFeedsResponse{
			GenericResponse: helper.GenericResponse{
				Code:    errro.SUCCESS,
				Message: "fetch feeds success",
			},
			Detail: f,
		}
		loggr.Log.V(4).Info(res.Message)
		return ctx.Status(fiber.StatusOK).JSON(&res)
	}

	if err.Code() == errro.EFEEDS_NO_FEEDS {
		res := helper.GenericResponse{
			Code:    err.Code(),
			Message: err.Error(),
		}
		return ctx.Status(fiber.StatusNotFound).JSON(res)
	}

	res := helper.GenericResponse{
		Code:    errro.EFEEDS_DB_ERR,
		Message: "fail to fetch feeds",
	}
	return err.WithDetail(res.Json(), errro.TDETAIL_JSON).SendError(ctx, fiber.StatusInternalServerError)
}
