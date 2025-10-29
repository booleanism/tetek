package helper

import (
	"github.com/booleanism/tetek/pkg/errro"
	"github.com/gofiber/fiber/v3"
)

type GenericResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func BindRequest(ctx fiber.Ctx, req any) (*GenericResponse, error) {
	err := ctx.Bind().All(req)
	if err != nil {
		return &GenericResponse{
			Code:    errro.INVALID_REQ,
			Message: "malformat request, failed to bind the request",
		}, nil
	}
	return nil, nil
}
