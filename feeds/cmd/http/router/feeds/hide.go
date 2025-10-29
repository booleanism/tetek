package feeds

import (
	"github.com/booleanism/tetek/auth/amqp"
	"github.com/booleanism/tetek/feeds/internal/repo"
	"github.com/booleanism/tetek/feeds/recipes"
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/go-logr/logr"
	"github.com/gofiber/fiber/v3"
)

type hideRequest struct {
	Id string `json:"id"`
}

type hideResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Detail  hideRequest `json:"detail"`
}

func Hide(rec recipes.FeedRecipes) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		loggr.Log.V(4).Info("new incoming hide request")
		req := hideRequest{}
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
			Id:       req.Id,
			HiddenTo: jwt.Claims.Uname,
		}

		err := rec.Hide(ctx, ff, jwt)
		if err != nil {
			res := hideResponse{
				Code:    err.Code(),
				Message: err.Error(),
			}
			return ctx.Status(fiber.StatusInternalServerError).JSON(&res)
		}

		res := hideResponse{
			Code:    errro.SUCCESS,
			Message: "feed hidden",
			Detail:  req,
		}
		return ctx.Status(fiber.StatusOK).JSON(&res)
	}
}
