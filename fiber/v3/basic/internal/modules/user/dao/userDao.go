package dao

import "github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/modules/user/models"

type ProfileRepository struct{}

func NewProfileRepository() *ProfileRepository {
	return &ProfileRepository{}
}

func (repo *ProfileRepository) BuildTemplateProfile(name, pass string) models.Profile {
	return models.Profile{
		Module:      "user",
		Description: "example business module for the template",
		Layers:      []string{"controller", "service", "dao", "models"},
	}
}
