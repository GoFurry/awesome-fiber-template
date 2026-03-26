package service

import (
	"strings"

	"github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/modules/user/models"
	"github.com/GoFurry/awesome-go-template/fiber/v3/basic/pkg/common"
)

type ProfileRepository interface {
	BuildTemplateProfile(name, pass string) models.Profile
}

type ProfileService interface {
	GetTemplateProfile(name, pass string) (models.Profile, common.Error)
}

type UserService struct {
	repository ProfileRepository
}

func NewUserService(repository ProfileRepository) *UserService {
	return &UserService{repository: repository}
}

func (s *UserService) GetTemplateProfile(name, pass string) (models.Profile, common.Error) {
	if s.repository == nil {
		return models.Profile{}, common.NewServiceError("user repository is not configured")
	}
	if strings.TrimSpace(name) == "" {
		return models.Profile{}, common.NewValidationError("name is required")
	}
	if strings.TrimSpace(pass) == "" {
		return models.Profile{}, common.NewValidationError("password is required")
	}
	return s.repository.BuildTemplateProfile(name, pass), nil
}
