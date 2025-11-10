package router

import (
	"context"
	"encoding/json"
	"time"

	"github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/feeds/cmd/http/api"
	"github.com/booleanism/tetek/feeds/internal/model"
	"github.com/booleanism/tetek/feeds/internal/repo"
	"github.com/booleanism/tetek/feeds/recipes"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

const TIMEOUT = 10

type getFeedsRequest struct {
	Id     uuid.UUID `query:"id"`
	Offset int       `query:"offset"`
	Type   string    `query:"type"` // M, J, A, S
}

type getFeedsResponse struct {
	helper.GenericResponse
	Detail []model.Feed `json:"detail"`
}

func (r getFeedsResponse) Json() []byte {
	j, _ := json.Marshal(r)
	return j
}

type FeedsRouter struct {
	rec recipes.FeedRecipes
}

func NewFeedRouter(rec recipes.FeedRecipes) FeedsRouter {
	return FeedsRouter{rec}
}

func (fr FeedsRouter) GetFeeds(ctx fiber.Ctx, param api.GetFeedsParams) error {
	if ctx.IsMiddleware() {
		return ctx.Next()
	}

	req := getFeedsRequest{}
	if err := helper.BindRequest(ctx, &req); err != nil {
		return err.SendError(ctx, fiber.StatusBadRequest)
	}

	jwt, _ := ctx.Locals("jwt").(*amqp.AuthResult)
	filter := repo.FeedsFilter{
		Offset: uint64(req.Offset),
		Type:   req.Type,
		Id:     req.Id,
	}

	cto, cancel := context.WithTimeout(
		context.WithValue(
			context.Background(),
			helper.RequestIdKey{},
			ctx.Locals(helper.RequestIdKey{})),
		TIMEOUT*time.Second)
	defer cancel()

	f, err := fr.rec.Feeds(cto, filter, jwt)
	if err == nil {
		res := getFeedsResponse{
			GenericResponse: helper.GenericResponse{
				Code:    errro.SUCCESS,
				Message: "fetch feeds success",
			},
			Detail: f,
		}
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
