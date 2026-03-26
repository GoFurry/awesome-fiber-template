package service

import (
	"net/http"
	"strings"

	"github.com/GoFurry/awesome-go-template/fiber/v3/heavy/internal/modules/user/dao"
	"github.com/GoFurry/awesome-go-template/fiber/v3/heavy/internal/modules/user/dto"
	"github.com/GoFurry/awesome-go-template/fiber/v3/heavy/internal/modules/user/models"
	"github.com/GoFurry/awesome-go-template/fiber/v3/heavy/pkg/common"
	pkgmodels "github.com/GoFurry/awesome-go-template/fiber/v3/heavy/pkg/models"
)

type UserRepository interface {
	Create(user *models.User) common.Error
	GetByID(id int64) (*models.User, common.Error)
	Update(user *models.User) common.Error
	Delete(id int64) common.Error
	List(filter dao.UserListFilter) ([]models.User, int64, common.Error)
}

type UserServiceAPI interface {
	Create(req dto.CreateUserRequest) (*models.User, common.Error)
	GetByID(id int64) (*models.User, common.Error)
	Update(id int64, req dto.UpdateUserRequest) (*models.User, common.Error)
	Delete(id int64) common.Error
	List(req dto.ListUsersRequest) (pkgmodels.PageResponse, common.Error)
}

type UserService struct {
	repository UserRepository
}

func NewUserService(repository UserRepository) *UserService {
	return &UserService{repository: repository}
}

func (s *UserService) Create(req dto.CreateUserRequest) (*models.User, common.Error) {
	if s.repository == nil {
		return nil, common.NewServiceError("user repository is not configured")
	}

	user, err := buildUser(req.Name, req.Email, req.Age, req.Status)
	if err != nil {
		return nil, err
	}

	if err := s.repository.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetByID(id int64) (*models.User, common.Error) {
	if s.repository == nil {
		return nil, common.NewServiceError("user repository is not configured")
	}
	if id <= 0 {
		return nil, common.NewValidationError("id must be greater than 0")
	}

	return s.repository.GetByID(id)
}

func (s *UserService) Update(id int64, req dto.UpdateUserRequest) (*models.User, common.Error) {
	if s.repository == nil {
		return nil, common.NewServiceError("user repository is not configured")
	}
	if id <= 0 {
		return nil, common.NewValidationError("id must be greater than 0")
	}

	user, err := s.repository.GetByID(id)
	if err != nil {
		return nil, err
	}

	updated, buildErr := buildUser(req.Name, req.Email, req.Age, req.Status)
	if buildErr != nil {
		return nil, buildErr
	}

	user.Name = updated.Name
	user.Email = updated.Email
	user.Age = updated.Age
	user.Status = updated.Status

	if err := s.repository.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Delete(id int64) common.Error {
	if s.repository == nil {
		return common.NewServiceError("user repository is not configured")
	}
	if id <= 0 {
		return common.NewValidationError("id must be greater than 0")
	}

	return s.repository.Delete(id)
}

func (s *UserService) List(req dto.ListUsersRequest) (pkgmodels.PageResponse, common.Error) {
	if s.repository == nil {
		return pkgmodels.PageResponse{}, common.NewServiceError("user repository is not configured")
	}

	if req.PageNum <= 0 {
		req.PageNum = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		return pkgmodels.PageResponse{}, common.NewError(common.RETURN_FAILED, http.StatusBadRequest, "page_size must be less than or equal to 100")
	}

	users, total, err := s.repository.List(dao.UserListFilter{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		Keyword:  strings.TrimSpace(req.Keyword),
	})
	if err != nil {
		return pkgmodels.PageResponse{}, err
	}

	return pkgmodels.PageResponse{
		Total: total,
		Data:  users,
	}, nil
}

func buildUser(name, email string, age int, status string) (*models.User, common.Error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))
	status = strings.TrimSpace(status)

	if name == "" {
		return nil, common.NewValidationError("name is required")
	}
	if email == "" {
		return nil, common.NewValidationError("email is required")
	}
	if !strings.Contains(email, "@") {
		return nil, common.NewValidationError("email format is invalid")
	}
	if age < 0 {
		return nil, common.NewValidationError("age must be greater than or equal to 0")
	}
	if status == "" {
		status = "active"
	}

	return &models.User{
		Name:   name,
		Email:  email,
		Age:    age,
		Status: status,
	}, nil
}
