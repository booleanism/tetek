package errro

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
)

type ErrMsg interface {
	Msg() string
}

type ErrCode interface {
	Code() int
}

type JSONable interface {
	JSON() []byte
}

type Error interface {
	error
	ErrMsg
	ErrCode
	WithDetail(detail []byte, detailType int) *resErr
	WithJSON(res JSONable) *resErr
}

type ResError interface {
	error
	ErrMsg
	ErrCode
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
	ErrAccountPasswdLength       = ErrAuthInvalidLoginParam
	ErrAccountUnknownCmd         = ErrCommUnknownCmd
)

const (
	Success           = 0
	ErrInvalidRequest = ErrAccountParseFail
	ErrLogging        = -9
	ErrTimeout        = -13
	ErrAcqPool        = -18
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
	ErrAuthJWTMalformat      = 21
)

const (
	ErrFeedsMissingRequiredField = 11
	ErrFeedsDBError              = ErrAccountDBError
	ErrFeedsNewFail              = 13
	ErrFeedsNoFeeds              = ErrAccountNoUser
	ErrFeedsDeleteFail           = 15
	ErrFeedsParseFail            = ErrAccountParseFail
	ErrFeedsUnknownCmd           = ErrCommUnknownCmd
	ErrFeedsUnathorized          = 22
	ErrFeedsUnableToHide         = -20
	ErrFeedsGetHiddenFeeds       = -21
	ErrFeedsHidden               = 23
	ErrFeedsInvalidType          = ErrAuthInvalidType
)

const (
	ErrCommDBError       = ErrFeedsDBError
	ErrCommQueryError    = -14
	ErrCommInsertError   = -19
	ErrCommScanError     = -15
	ErrCommBuildTreeFail = -10
	ErrCommPubFail       = -11
	ErrCommConsumeFail   = -12
	ErrCommNoConsume     = ErrAccountNoUser
	ErrCommAcquirePool   = -16
	ErrCommUnknownCmd    = 20
	ErrCommParseFail     = ErrAccountParseFail
	ErrCommNoComments    = ErrFeedsNoFeeds
)

type Err struct {
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

func New(code int, msg string) *Err {
	return &Err{code, msg, nil, nil, 0}
}

func FromError(code int, msg string, e error) *Err {
	return &Err{code, msg, e, nil, 0}
}

func (e Err) Msg() string {
	return e.m
}

func (e *Err) Error() string {
	if e.e != nil && e.m != e.e.Error() {
		return fmt.Sprintf("%s: %s", e.m, e.e.Error())
	}

	return e.m
}

func (e *Err) Code() int {
	return e.c
}

func (e *Err) WithDetail(detail []byte, detailType int) *resErr {
	e.detail = detail
	e.t = detailType
	return &resErr{e}
}

func (e *Err) WithJSON(res JSONable) *resErr {
	e.detail = res.JSON()
	e.t = TDetailJSON
	return &resErr{e}
}

type resErr struct {
	*Err
}

func (e *resErr) Error() string {
	if e.e != nil && e.m != e.e.Error() {
		return fmt.Sprintf("%s: %s", e.m, e.e.Error())
	}

	return e.m
}

func (e *resErr) Msg() string {
	return e.Err.Msg()
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
