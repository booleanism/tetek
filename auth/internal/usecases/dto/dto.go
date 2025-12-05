package dto

import (
	"encoding/json"
)

type LoginRequest struct {
	Uname  string `json:"uname,omitempty"`
	Email  string `json:"email,omitempty"`
	Passwd string `json:"passwd,omitempty"`
}

type LoginResponse struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Detail  LoginRequest `json:"detail"`
}

func (r LoginResponse) JSON() []byte {
	j, _ := json.Marshal(r)
	return j
}
