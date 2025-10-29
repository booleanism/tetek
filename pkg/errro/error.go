package errro

import (
	"errors"

	"github.com/gofiber/fiber/v3"
)

type Error interface {
	error
	Code() int
	ToFiber() *fiber.Error
}

const (
	EACCOUNT_CANT_LOGIN           = -8
	EACCOUNT_PASSWD_HASH_FAIL     = -6
	EACCOUNT_SERVICE_UNAVAILABLE  = -4
	EACCOUNT_DB_ERR               = -2
	EACCOUNT_PARSE_FAIL           = -1
	EACCOUNT_NO_USER              = 1
	EACCOUNT_EMPTY_PARAM          = 2
	EACCOUNT_USER_ALREADY_EXIST   = 7
	EACCOUNT_REGIST_FAIL          = 8
	EACCOUNT_INVALID_REGIST_PARAM = EAUTH_INVALID_LOGIN_PARAM
)

const SUCCESS = 0
const INVALID_REQ = 3

const (
	EAUTH_JWT_GENERATAION_FAIL     = -7
	EAUTH_RETRIEVE_USER_FAIL       = -5
	EAUTH_SERVICE_UNAVAILABLE      = -3
	EAUTH_PARSE_FAIL               = EACCOUNT_PARSE_FAIL
	EAUTH_JWT_VERIFY_FAIL          = 4
	EAUTH_MISSMATCH_AUTH_MECHANISM = 5
	EAUTH_INVALID_LOGIN_PARAM      = 6
	EAUTH_INVALID_CREDS            = 9
	EAUTH_MISSING_HEADER           = 10
	EAUTH_INVALID_AUTH_RESULT_TYPE = 12
)

const (
	EFEEDS_MISSING_REQUIRED_FIELD = 11
	EFEEDS_DB_ERR                 = EACCOUNT_DB_ERR
	EFEEDS_NEW_FAIL               = 13
	EFEEDS_NO_FEEDS               = 14
	EFEEDS_DELETE_FAIL            = 15
)

type err struct {
	c int
	m string
	e error
}

func New(code int, msg string) *err {
	return &err{code, msg, errors.New(msg)}
}

func FromError(code int, e error) *err {
	return &err{code, "", e}
}

func (e *err) ToFiber() *fiber.Error {
	var errFib *fiber.Error
	if ok := errors.As(e.e, &errFib); ok {
		return errFib
	}

	return fiber.NewError(fiber.StatusInternalServerError, "failed to cast error into fiber.Error")
}

func (e *err) Error() string {
	return e.m
}

func (e *err) Code() int {
	return e.c
}
