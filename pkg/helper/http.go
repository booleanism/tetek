package helper

import (
	"encoding/json"
	"time"

	"github.com/booleanism/tetek/pkg/errro"
	"github.com/booleanism/tetek/pkg/loggr"
	"github.com/gofiber/fiber/v3"
)

const Timeout = 5 * time.Second

type GenericResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (r GenericResponse) JSON() []byte {
	j, _ := json.Marshal(r)
	return j
}

func BindRequest(ctx fiber.Ctx, req any) errro.ResError {
	_, log := loggr.GetLogger(ctx.Context(), "reques-binder")
	err := ctx.Bind().All(req)
	if err != nil {
		r := &GenericResponse{
			Code:    errro.ErrInvalidRequest,
			Message: "malformat request, failed to bind the request",
		}
		e := errro.FromError(errro.ErrInvalidRequest, r.Message, err)
		log.V(1).Info(e.Msg(), "error", err)
		return e.WithDetail(r.JSON(), errro.TDetailJSON)
	}
	return nil
}
