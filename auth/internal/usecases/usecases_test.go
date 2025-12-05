package usecases_test

import (
	"context"
	"errors"
	"testing"

	messaging "github.com/booleanism/tetek/account/infra/messaging/rabbitmq"
	"github.com/booleanism/tetek/auth/internal/usecases"
	"github.com/booleanism/tetek/auth/internal/usecases/dto"
	"github.com/booleanism/tetek/auth/internal/usecases/jwt"
	"github.com/booleanism/tetek/pkg/contracts"
	"github.com/booleanism/tetek/pkg/errro"
	"golang.org/x/crypto/bcrypt"
)

type mockAccSubs struct {
	res  *messaging.AccountResult
	fail error
}

func (m mockAccSubs) Publish(context.Context, any) error {
	return m.fail
}

func (m mockAccSubs) Consume(_ context.Context, res any) error {
	if m.fail == nil {
		*(res.(**messaging.AccountResult)) = m.res
	}

	return m.fail
}

func (m mockAccSubs) Name() string {
	return "test"
}

type mockJwt struct {
	fail error
}

func (m mockJwt) Verify(context.Context, string, **jwt.JwtClaims) error {
	return m.fail
}

func (m mockJwt) Generate(messaging.User) (string, error) {
	if m.fail == nil {
		return "t0k3n", nil
	}

	return "", m.fail
}

type loginRecipes struct {
	jwt.JwtRecipes
	contracts.AccountDealer
}

type test struct {
	loginRecipes
	expected int
}

type testData struct {
	dto.LoginRequest
	test
}

func TestLogin(t *testing.T) {
	pw, err := bcrypt.GenerateFromPassword([]byte("test"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	data := []testData{
		{LoginRequest: dto.LoginRequest{}, test: test{loginRecipes: loginRecipes{JwtRecipes: mockJwt{nil}, AccountDealer: mockAccSubs{nil, nil}}, expected: errro.ErrAuthInvalidLoginParam}},
		{LoginRequest: dto.LoginRequest{Uname: "test"}, test: test{loginRecipes: loginRecipes{JwtRecipes: mockJwt{nil}, AccountDealer: mockAccSubs{nil, nil}}, expected: errro.ErrAuthInvalidLoginParam}},
		{LoginRequest: dto.LoginRequest{Passwd: "test"}, test: test{loginRecipes: loginRecipes{JwtRecipes: mockJwt{nil}, AccountDealer: mockAccSubs{nil, nil}}, expected: errro.ErrAuthInvalidLoginParam}},
		{LoginRequest: dto.LoginRequest{Uname: "test", Passwd: "test"}, test: test{loginRecipes: loginRecipes{JwtRecipes: mockJwt{nil}, AccountDealer: mockAccSubs{&messaging.AccountResult{Detail: messaging.User{Passwd: string(pw)}}, nil}}, expected: errro.Success}},
		{LoginRequest: dto.LoginRequest{Uname: "test", Passwd: "wrong passwd"}, test: test{loginRecipes: loginRecipes{JwtRecipes: mockJwt{nil}, AccountDealer: mockAccSubs{&messaging.AccountResult{Detail: messaging.User{Passwd: string(pw)}}, nil}}, expected: errro.ErrAuthInvalidCreds}},
		{LoginRequest: dto.LoginRequest{Uname: "test", Passwd: "test"}, test: test{loginRecipes: loginRecipes{JwtRecipes: mockJwt{errors.New("fail generate")}, AccountDealer: mockAccSubs{&messaging.AccountResult{Detail: messaging.User{Passwd: string(pw)}}, nil}}, expected: errro.ErrAuthJWTGenerationFail}},
		{LoginRequest: dto.LoginRequest{Uname: "test", Passwd: "test"}, test: test{loginRecipes: loginRecipes{JwtRecipes: mockJwt{nil}, AccountDealer: mockAccSubs{&messaging.AccountResult{Detail: messaging.User{Passwd: string(pw)}, Code: errro.ErrAccountNoUser}, nil}}, expected: errro.ErrAccountNoUser}},
		{LoginRequest: dto.LoginRequest{Uname: "test", Passwd: "test"}, test: test{loginRecipes: loginRecipes{JwtRecipes: mockJwt{nil}, AccountDealer: mockAccSubs{&messaging.AccountResult{Detail: messaging.User{Passwd: string(pw)}, Code: errro.ErrAccountDBError}, nil}}, expected: errro.ErrAuthFailRetrieveUser}},
		{LoginRequest: dto.LoginRequest{Uname: "test", Passwd: "test"}, test: test{loginRecipes: loginRecipes{JwtRecipes: mockJwt{nil}, AccountDealer: mockAccSubs{&messaging.AccountResult{Detail: messaging.User{Passwd: string(pw)}}, errors.New("cannot publish or consume")}}, expected: errro.ErrAccountServiceUnavailable}},
	}

	for _, v := range data {
		rec := usecases.NewAuthUseCases(v.AccountDealer, v.JwtRecipes)
		_, err := rec.Login(context.Background(), v.LoginRequest)

		code := 0
		if err != nil {
			code = err.Code()
		}

		if code != v.expected {
			t.Fail()
		}
	}
}
