package helper

import (
	"encoding/json"

	"github.com/booleanism/tetek/pkg/errro"
	"github.com/gofiber/fiber/v3"
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
