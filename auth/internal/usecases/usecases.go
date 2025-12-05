package usecases

import (
	"github.com/booleanism/tetek/auth/internal/usecases/jwt"
	"github.com/booleanism/tetek/pkg/contracts"
)

type AuthUseCases interface {
	LoginUseCase
}

type usecases struct {
	l   contracts.AccountDealer
	jwt jwt.JwtRecipes
}

func NewAuthUseCases(accDealer contracts.AccountDealer, jwt jwt.JwtRecipes) AuthUseCases {
	return usecases{accDealer, jwt}
}
