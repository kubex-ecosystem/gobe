package gdbasez

import (
	"context"

	svc "github.com/kubex-ecosystem/gdbase/factory"
	user "github.com/kubex-ecosystem/gdbase/factory/models"
)

type UserService = user.UserService
type UserModel = user.UserModel
type UserModelType = user.UserModel
type UserRepo = user.UserRepo

func NewUserService(db user.UserRepo) UserService {
	return user.NewUserService(db)
}

func NewUserRepo(ctx context.Context, dbService *svc.DBServiceImpl) UserRepo {
	return user.NewUserRepo(ctx, dbService)
}

func NewUserModel(username, name, email string) UserModel {
	return user.NewUserModel(username, name, email)
}
