package dto

import (
	"encoding/json"

	"github.com/booleanism/tetek/account/internal/domain/entities"
	"github.com/booleanism/tetek/pkg/helper"
	"github.com/google/uuid"
)

type User = entities.User

type RegistRequest struct {
	ID     uuid.UUID `json:"id"`
	Uname  string    `json:"uname"`
	Email  string    `json:"email"`
	Passwd string    `json:"passwd"`
}

type RegistResponse struct {
	helper.GenericResponse
	Detail RegistRequest `json:"detail"`
}

func (r RegistResponse) JSON() []byte {
	j, _ := json.Marshal(r)
	return j
}

type ProfileRequest struct {
	Uname string `uri:"uname"`
}

type ProfileResponse struct {
	helper.GenericResponse
	Detail User `json:"detail"`
}

func (r ProfileResponse) JSON() []byte {
	j, _ := json.Marshal(r)
	return j
}
