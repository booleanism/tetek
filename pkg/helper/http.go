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

func (r GenericResponse) JSON() []byte {
	j, _ := json.Marshal(r)
	return j
}

func BindRequest(ctx fiber.Ctx, req any) errro.ResError {
	err := ctx.Bind().All(req)
	if err != nil {
		r := &GenericResponse{
			Code:    errro.ErrInvalidRequest,
			Message: "malformat request, failed to bind the request",
		}
		return errro.New(errro.ErrInvalidRequest, "malformat request, failed to bind the request").WithDetail(r.JSON(), errro.TDetailJSON)
	}
	return nil
}
