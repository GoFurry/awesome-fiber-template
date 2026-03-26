package controller

import (
	"encoding/json"
	"strconv"

	"github.com/GoFurry/awesome-go-template/fiber/v3/heavy/internal/modules/user/dto"
	"github.com/GoFurry/awesome-go-template/fiber/v3/heavy/internal/modules/user/service"
	"github.com/GoFurry/awesome-go-template/fiber/v3/heavy/pkg/common"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	service service.UserServiceAPI
}

func NewHandler(svc service.UserServiceAPI) *Handler {
	return &Handler{service: svc}
}

func (handler *Handler) CreateUser(c fiber.Ctx) error {
	var req dto.CreateUserRequest
	if err := decodeJSONBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}

	data, err := handler.service.Create(req)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}

	return common.NewResponse(c).SuccessWithData(data)
}

func (handler *Handler) GetUser(c fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}

	data, serviceErr := handler.service.GetByID(id)
	if serviceErr != nil {
		return common.NewResponse(c).Error(serviceErr)
	}

	return common.NewResponse(c).SuccessWithData(data)
}

func (handler *Handler) UpdateUser(c fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}

	var req dto.UpdateUserRequest
	if err := decodeJSONBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}

	data, serviceErr := handler.service.Update(id, req)
	if serviceErr != nil {
		return common.NewResponse(c).Error(serviceErr)
	}

	return common.NewResponse(c).SuccessWithData(data)
}

func (handler *Handler) DeleteUser(c fiber.Ctx) error {
	id, err := parseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}

	if serviceErr := handler.service.Delete(id); serviceErr != nil {
		return common.NewResponse(c).Error(serviceErr)
	}

	return common.NewResponse(c).SuccessWithData(fiber.Map{
		"deleted": true,
		"id":      id,
	})
}

func (handler *Handler) ListUsers(c fiber.Ctx) error {
	req := dto.ListUsersRequest{
		PageNum:  parseQueryInt(c, "page_num", 1),
		PageSize: parseQueryInt(c, "page_size", 10),
		Keyword:  c.Query("keyword", ""),
	}

	data, err := handler.service.List(req)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}

	return common.NewResponse(c).SuccessWithData(data)
}

func decodeJSONBody(c fiber.Ctx, target any) common.Error {
	if err := json.Unmarshal(c.Body(), target); err != nil {
		return common.NewValidationError("request body must be valid json")
	}

	return nil
}

func parseIDParam(c fiber.Ctx) (int64, common.Error) {
	id, err := strconv.ParseInt(c.Params("id", "0"), 10, 64)
	if err != nil || id <= 0 {
		return 0, common.NewValidationError("id must be a positive integer")
	}

	return id, nil
}

func parseQueryInt(c fiber.Ctx, key string, fallback int) int {
	raw := c.Query(key, "")
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}

	return value
}
