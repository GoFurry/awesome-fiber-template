package service

import "github.com/GoFurry/awesome-go-template/fiber/v3/apps/user/models"

func GetTemplateProfile() models.Profile {
	return models.Profile{
		Module:      "user",
		Description: "example business module for the template",
		Layers:      []string{"controller", "service", "dao", "models"},
	}
}
