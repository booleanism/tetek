package recipes_test

import (
	"context"
	"errors"
	"testing"

	"github.com/booleanism/tetek/account/amqp"
	"github.com/booleanism/tetek/auth/internal/jwt"
	"github.com/booleanism/tetek/auth/recipes"
	"github.com/booleanism/tetek/pkg/contracts"
	"github.com/booleanism/tetek/pkg/errro"
	"golang.org/x/crypto/bcrypt"
)

type mockAccSubs struct {
	res  *amqp.AccountResult
	fail error
}

func (m mockAccSubs) Publish(context.Context, amqp.AccountTask) error {
	return m.fail
}

func (m mockAccSubs) Consume(_ context.Context, res **amqp.AccountResult) error {
	if m.fail == nil {
		(*res) = m.res
	}

	return m.fail
}

type mockJwt struct {
	fail error
}

func (m mockJwt) Verify(context.Context, string, **jwt.JwtClaims) error {
	return m.fail
}

func (m mockJwt) Generate(amqp.User) (string, error) {
	if m.fail == nil {
		return "t0k3n", nil
	}

	return "", m.fail
}

type loginRecipes struct {
	jwt.JwtRecipes
	contracts.AccountSubscribe
}

type test struct {
	loginRecipes
	expected int
}

type testData struct {
	recipes.LoginRequest
	test
}

func TestLogin(t *testing.T) {
	pw, err := bcrypt.GenerateFromPassword([]byte("test"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	data := []testData{
		{LoginRequest: recipes.LoginRequest{}, test: test{loginRecipes: loginRecipes{JwtRecipes: mockJwt{nil}, AccountSubscribe: mockAccSubs{nil, nil}}, expected: errro.ErrAuthInvalidLoginParam}},
		{LoginRequest: recipes.LoginRequest{Uname: "test"}, test: test{loginRecipes: loginRecipes{JwtRecipes: mockJwt{nil}, AccountSubscribe: mockAccSubs{nil, nil}}, expected: errro.ErrAuthInvalidLoginParam}},
		{LoginRequest: recipes.LoginRequest{Passwd: "test"}, test: test{loginRecipes: loginRecipes{JwtRecipes: mockJwt{nil}, AccountSubscribe: mockAccSubs{nil, nil}}, expected: errro.ErrAuthInvalidLoginParam}},
		{LoginRequest: recipes.LoginRequest{Uname: "test", Passwd: "test"}, test: test{loginRecipes: loginRecipes{JwtRecipes: mockJwt{nil}, AccountSubscribe: mockAccSubs{&amqp.AccountResult{Detail: amqp.User{Passwd: string(pw)}}, nil}}, expected: errro.Success}},
		{LoginRequest: recipes.LoginRequest{Uname: "test", Passwd: "wrong passwd"}, test: test{loginRecipes: loginRecipes{JwtRecipes: mockJwt{nil}, AccountSubscribe: mockAccSubs{&amqp.AccountResult{Detail: amqp.User{Passwd: string(pw)}}, nil}}, expected: errro.ErrAuthInvalidCreds}},
		{LoginRequest: recipes.LoginRequest{Uname: "test", Passwd: "test"}, test: test{loginRecipes: loginRecipes{JwtRecipes: mockJwt{errors.New("fail generate")}, AccountSubscribe: mockAccSubs{&amqp.AccountResult{Detail: amqp.User{Passwd: string(pw)}}, nil}}, expected: errro.ErrAuthJWTGenerationFail}},
		{LoginRequest: recipes.LoginRequest{Uname: "test", Passwd: "test"}, test: test{loginRecipes: loginRecipes{JwtRecipes: mockJwt{nil}, AccountSubscribe: mockAccSubs{&amqp.AccountResult{Detail: amqp.User{Passwd: string(pw)}, Code: errro.ErrAccountNoUser}, nil}}, expected: errro.ErrAccountNoUser}},
		{LoginRequest: recipes.LoginRequest{Uname: "test", Passwd: "test"}, test: test{loginRecipes: loginRecipes{JwtRecipes: mockJwt{nil}, AccountSubscribe: mockAccSubs{&amqp.AccountResult{Detail: amqp.User{Passwd: string(pw)}, Code: errro.ErrAccountDBError}, nil}}, expected: errro.ErrAuthFailRetrieveUser}},
		{LoginRequest: recipes.LoginRequest{Uname: "test", Passwd: "test"}, test: test{loginRecipes: loginRecipes{JwtRecipes: mockJwt{nil}, AccountSubscribe: mockAccSubs{&amqp.AccountResult{Detail: amqp.User{Passwd: string(pw)}}, errors.New("cannot publish or consume")}}, expected: errro.ErrAccountServiceUnavailable}},
	}

	for _, v := range data {
		rec := recipes.NewLogin(v.AccountSubscribe, v.JwtRecipes)
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
