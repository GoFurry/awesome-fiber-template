package service

import (
	"github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/modules/user/models"
	"github.com/GoFurry/awesome-go-template/fiber/v3/basic/pkg/common"
)

type userService struct{}

var userSingleton = new(userService)

func GetUserService() *userService { return userSingleton }

func (s userService) GetTemplateProfile(name string, pass string) (models.Profile, common.Error) {
	return models.Profile{
		Module:      "user",
		Description: "example business module for the template",
		Layers:      []string{"controller", "service", "dao", "models"},
	}, nil
}
