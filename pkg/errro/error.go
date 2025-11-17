package errro

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
)

type Error interface {
	error
	Code() int
	// TODO: use interface returning []byte for the detail
	WithDetail(detail []byte, detailType int) *resErr
}

type ResError interface {
	error
	Code() int
	SendError(fiber.Ctx, int) error
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

const (
	SUCCESS       = 0
	INVALID_REQ   = EACCOUNT_PARSE_FAIL
	YOUR_BAD      = 16
	MY_BAD        = 17
	LOGGING_ERROR = -9
	TIMEOUT       = -13
)

const (
	EAUTH_JWT_GENERATAION_FAIL     = -7
	EAUTH_RETRIEVE_USER_FAIL       = -5
	EAUTH_SERVICE_UNAVAILABLE      = -3
	EAUTH_PARSE_FAIL               = EACCOUNT_PARSE_FAIL
	EAUTH_JWT_VERIFY_FAIL          = 4
	EAUTH_EMPTY_JWT                = 19
	EAUTH_MISSMATCH_AUTH_MECHANISM = 5
	EAUTH_INVALID_LOGIN_PARAM      = 6
	EAUTH_INVALID_CREDS            = 9
	EAUTH_MISSING_HEADER           = 10
	EAUTH_INVALID_AUTH_RESULT_TYPE = 12
	EAUTH_JWT_MALFORMAT            = 19
)

const (
	EFEEDS_MISSING_REQUIRED_FIELD = 11
	EFEEDS_DB_ERR                 = EACCOUNT_DB_ERR
	EFEEDS_NEW_FAIL               = 13
	EFEEDS_NO_FEEDS               = EACCOUNT_NO_USER
	EFEEDS_DELETE_FAIL            = 15
	EFEEDS_PARSE_FAIL             = EACCOUNT_PARSE_FAIL
)

const (
	ECOMM_DB_ERR          = EFEEDS_DB_ERR
	ECOMM_QUERY_ERR       = -14
	ECOMM_SCAN_ERR        = -15
	ECOMM_FAIL_BUILD_TREE = -10
	ECOMM_PUB_FAIL        = -11
	ECOMM_CONSUME_FAIL    = -12
	ECOMM_NO_COMM         = EACCOUNT_NO_USER
)

type err struct {
	c      int
	m      string
	e      error
	detail []byte
	t      int
}

const (
	TDETAIL_RAW = iota
	TDETAIL_JSON
)

func New(code int, msg string) *err {
	return &err{code, msg, nil, nil, 0}
}

func FromError(code int, msg string, e error) *err {
	return &err{code, msg, e, nil, 0}
}

func (e *err) Error() string {
	if e.e != nil && e.m != e.e.Error() {
		return fmt.Sprintf("%s: %s", e.m, e.e.Error())
	}

	return e.m
}

func (e *err) Code() int {
	return e.c
}

func (e *err) WithDetail(detail []byte, detailType int) *resErr {
	e.detail = detail
	e.t = detailType
	return &resErr{e}
}

type Jsonable interface {
	Json() []byte
}

func (e *err) WithJson(res Jsonable) *resErr {
	e.detail = res.Json()
	e.t = TDETAIL_JSON
	return &resErr{e}
}

type resErr struct {
	*err
}

func (e *resErr) Error() string {
	if e.e != nil && e.m != e.e.Error() {
		return fmt.Sprintf("%s: %s", e.m, e.e.Error())
	}

	return e.m
}

func (e *resErr) SendError(ctx fiber.Ctx, status int) error {
	if e.detail == nil {
		return ctx.Status(status).Send([]byte(e.m))
	}

	if e.t == TDETAIL_JSON {
		ctx.Set("Content-Type", "application/json")
		return ctx.Status(status).Send(e.detail)
	}

	return ctx.Status(status).Send(e.detail)
}

func (e *resErr) Code() int {
	return e.c
}
