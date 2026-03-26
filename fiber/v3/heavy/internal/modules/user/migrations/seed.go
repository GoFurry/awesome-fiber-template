package migrations

import (
	"github.com/GoFurry/awesome-go-template/fiber/v3/heavy/internal/modules/user/models"
	"gorm.io/gorm"
)

type SeedTemplateUser struct{}

func (SeedTemplateUser) Name() string {
	return "20260327_seed_template_user"
}

func (SeedTemplateUser) Up(tx *gorm.DB) error {
	user := models.User{
		Name:   "Template User",
		Email:  "template@example.com",
		Age:    18,
		Status: "active",
	}

	return tx.Where("email = ?", user.Email).FirstOrCreate(&user).Error
}
