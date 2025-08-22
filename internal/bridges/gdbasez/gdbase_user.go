package gdbasez

import (
	user "github.com/rafa-mori/gdbase/factory/models"
)

type UserService = user.UserService
type UserModel = user.UserModel
type UserRepo = user.UserRepo

func NewUserService(db user.UserRepo) UserService {
	return user.NewUserService(db)
}
