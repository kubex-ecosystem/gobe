package gdbasez

import (
	user "github.com/kubex-ecosystem/gdbase/factory/models"
	"gorm.io/gorm"
)

type UserService = user.UserService
type UserModel = user.UserModel
type UserModelType = user.UserModel
type UserRepo = user.UserRepo

func NewUserService(db user.UserRepo) UserService {
	return user.NewUserService(db)
}

func NewUserRepo(db *gorm.DB) UserRepo {
	return user.NewUserRepo(db)
}

func NewUserModel(username, name, email string) UserModel {
	return user.NewUserModel(username, name, email)
}
