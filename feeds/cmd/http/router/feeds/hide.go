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
	helper.GenericResponse
	Detail hideRequest `json:"detail"`
}

func Hide(rec recipes.FeedRecipes) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		loggr.Log.V(4).Info("new incoming hide request")
		req := hideRequest{}
		if err := helper.BindRequest(ctx, &req); err != nil {
			return loggr.Log.ErrorRes(3, func(z logr.LogSink) error {
				z.Error(err, "failed to bind request", "body", ctx.Body())
				return err.SendError(ctx, fiber.StatusBadRequest)
			})
		}

		jwt, ok := ctx.Locals("jwt").(*amqp.AuthResult)
		if !ok {
			return loggr.Log.ErrorRes(2, func(z logr.LogSink) error {
				res := helper.GenericResponse{
					Code:    errro.EAUTH_INVALID_AUTH_RESULT_TYPE,
					Message: "does not represent jwt type",
				}
				e := errro.New(res.Code, res.Message).WithDetail(res.Json(), errro.TDETAIL_JSON)
				z.Error(e, res.Message)
				return e
			})
		}

		ff := repo.FeedsFilter{
			Id:       req.Id,
			HiddenTo: jwt.Claims.Uname,
		}

		err := rec.Hide(ctx, ff, jwt)
		if err != nil {
			res := helper.GenericResponse{
				Code:    err.Code(),
				Message: err.Error(),
			}
			return ctx.Status(fiber.StatusInternalServerError).JSON(&res)
		}

		res := hideResponse{
			GenericResponse: helper.GenericResponse{
				Code:    errro.SUCCESS,
				Message: "feed hidden",
			},
			Detail: req,
		}
		return ctx.Status(fiber.StatusOK).JSON(&res)
	}
}
