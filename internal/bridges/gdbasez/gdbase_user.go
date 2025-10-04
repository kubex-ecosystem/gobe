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

func NewUserRepo(ctx context.Context, dbService *svc.DBServiceImpl, dbName string) UserRepo {
	if dbService == nil {
		return nil
	}
	db, err := dbService.GetDB(ctx, dbName)
	if err != nil {
		return nil
	}
	if db == nil {
		return nil
	}
	return user.NewUserRepo(db)
}

func NewUserModel(username, name, email string) UserModel {
	return user.NewUserModel(username, name, email)
}
