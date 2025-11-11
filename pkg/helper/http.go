package helper

import (
	"context"
	"encoding/json"

	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/keystore"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type GenericResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (r GenericResponse) Json() []byte {
	j, _ := json.Marshal(r)
	return j
}

func BindRequest(ctx fiber.Ctx, req any) errro.ResError {
	err := ctx.Bind().All(req)
	if err != nil {
		r := &GenericResponse{
			Code:    errro.INVALID_REQ,
			Message: "malformat request, failed to bind the request",
		}
		return errro.New(errro.INVALID_REQ, "malformat request, failed to bind the request").WithDetail(r.Json(), errro.TDETAIL_JSON)
	}
	return nil
}

func GenerateRequestId(ctx fiber.Ctx) error {
	ctx.SetContext(context.WithValue(ctx.Context(), keystore.RequestId{}, uuid.NewString()))
	return ctx.Next()
}
