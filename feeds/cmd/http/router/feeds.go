package router

import (
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
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Detail  []model.Feed `json:"detail"`
}

func Feeds(rec recipes.FeedRecipes) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		loggr.Log.V(4).Info("new incomming feeds request")
		if ctx.Method() != "GET" {
			return ctx.Next()
		}

		req := getFeedsRequest{}
		if res, err := helper.BindRequest(ctx, &req); err != nil {
			return loggr.Log.Error(3, func(z logr.LogSink) errro.Error {
				z.Error(err, res.Message)
				return errro.New(res.Code, res.Message)
			}).ToFiber()
		}

		jwt, _ := ctx.Locals("jwt").(*amqp.AuthResult)
		filter := repo.FeedsFilter{
			Offset: uint64(req.Offset),
			Type:   req.Type,
		}

		f, err := rec.Feeds(ctx, filter, jwt)
		if err == nil {
			res := getFeedsResponse{
				Code:    errro.SUCCESS,
				Message: "fetch feeds success",
				Detail:  f,
			}
			loggr.Log.V(4).Info(res.Message)
			return ctx.Status(fiber.StatusOK).JSON(&res)
		}

		if err.Code() == errro.EFEEDS_NO_FEEDS {
			res := getFeedsResponse{
				Code:    err.Code(),
				Message: err.Error(),
			}
			return ctx.Status(fiber.StatusNotFound).JSON(res)
		}

		res := getFeedsResponse{
			Code:    errro.EFEEDS_DB_ERR,
			Message: "fail to fetch feeds",
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(&res)
	}
}
