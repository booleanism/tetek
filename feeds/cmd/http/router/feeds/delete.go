package feeds

import (
	"strings"

	"github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/feeds/internal/repo"
	"github.com/booleanism/tetek/feeds/recipes"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/go-logr/logr"
	"github.com/gofiber/fiber/v3"
)

type deleteRequest struct {
	Id string `json:"id"`
}

type deleteResponse struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Detail  deleteRequest `json:"detail"`
}

func Delete(rec recipes.FeedRecipes) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		loggr.Log.V(4).Info("new incoming delete request")
		req := deleteRequest{}
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

		ff := repo.FeedsFilter{
			Id: req.Id,
		}

		// only moderator freely to delete feed
		if strings.ToLower(jwt.Claims.Role) != "m" {
			loggr.Log.V(4).Info("normal user not moderator")
			ff.By = jwt.Claims.Uname
		}

		err := rec.Delete(ctx, ff, jwt)
		if err == nil {
			res := deleteResponse{
				Code:    errro.SUCCESS,
				Message: "feed deleted",
				Detail:  req,
			}
			return ctx.Status(fiber.StatusOK).JSON(&res)
		}

		if err.Code() == errro.EFEEDS_NO_FEEDS && ff.By != "" {
			res := deleteResponse{
				Code:    errro.EFEEDS_DELETE_FAIL,
				Message: "unauthorized user or there is no such feed",
			}
			return ctx.Status(fiber.StatusUnauthorized).JSON(&res)
		}

		if err.Code() == errro.EFEEDS_NO_FEEDS {
			res := deleteResponse{
				Code:    errro.EFEEDS_DELETE_FAIL,
				Message: err.Error(),
			}
			return ctx.Status(fiber.StatusInternalServerError).JSON(&res)
		}

		res := deleteResponse{
			Code:    err.Code(),
			Message: err.Error(),
		}
		return ctx.Status(fiber.StatusNotModified).JSON(&res)
	}
}
