package usecases

import "github.com/booleanism/tetek/account/internal/usecases/repo"

type AccountUseCases interface {
	GetUserProfileUseCase
	UserRegistrationUseCase
}

type usecases struct {
	repo repo.UserRepo
}

func NewAccountUseCases(repo repo.UserRepo) AccountUseCases {
	return usecases{repo}
}
