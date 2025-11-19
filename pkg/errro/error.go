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
	ErrAccountCantLogin          = -8
	ErrAccountPasswdHashFail     = -6
	ErrAccountServiceUnavailable = -4
	ErrAccountDBError            = -2
	ErrAccountParseFail          = -1
	ErrAccountNoUser             = 1
	ErrAccountEmptyParam         = 2
	ErrAccountUserAlreadyExist   = 7
	ErrAccountRegistFail         = 8
	ErrAccountInvalidRegistParam = ErrAuthInvalidLoginParam
)

const (
	Success           = 0
	ErrInvalidRequest = ErrAccountParseFail
	ErrLogging        = -9
	ErrTimeout        = -13
)

const (
	ErrAuthJWTGenerationFail = -7
	ErrAuthFailRetrieveUser  = -5
	ErrServiceUnavailable    = ErrAccountServiceUnavailable
	ErrAuthParseFail         = ErrAccountParseFail
	ErrAuthJWTVerifyFail     = 4
	ErrAuthEmptyJWT          = 19
	ErrAuthMissmatchScheme   = 5
	ErrAuthInvalidLoginParam = 6
	ErrAuthInvalidCreds      = 9
	ErrAuthMissingHeader     = 10
	ErrAuthInvalidType       = 12
	ErrAuthJWTMalformat      = 19
)

const (
	ErrFeedsMissingRequiredField = 11
	ErrFeedsDBError              = ErrAccountDBError
	ErrFeedsNewFail              = 13
	ErrFeedsNoFeeds              = ErrAccountNoUser
	ErrFeedsDeleteFail           = 15
	ErrFeedsParseFail            = ErrAccountParseFail
)

const (
	ErrCommDBError       = ErrFeedsDBError
	ErrCommQueryError    = -14
	ErrCommScanError     = -15
	ErrCommBuildTreeFail = -10
	ErrCommPubFail       = -11
	ErrCommConsumeFail   = -12
	ErrCommNoConsume     = ErrAccountNoUser
	ErrCommDownvoteFail  = -16
	ErrCommUpvoteFail    = -17
)

type err struct {
	c      int
	m      string
	e      error
	detail []byte
	t      int
}

const (
	TDetailRaw = iota
	TDetailJSON
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

func (e *err) WithJSON(res Jsonable) *resErr {
	e.detail = res.Json()
	e.t = TDetailJSON
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

	if e.t == TDetailJSON {
		ctx.Set("Content-Type", "application/json")
		return ctx.Status(status).Send(e.detail)
	}

	return ctx.Status(status).Send(e.detail)
}

func (e *resErr) Code() int {
	return e.c
}
