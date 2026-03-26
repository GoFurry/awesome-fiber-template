package controller

import (
	"github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/modules/user/service"
	"github.com/GoFurry/awesome-go-template/fiber/v3/basic/pkg/common"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	service service.ProfileService
}

func NewHandler(svc service.ProfileService) *Handler {
	return &Handler{service: svc}
}

// @Summary 鑾峰彇鐢ㄦ埛淇℃伅
// @Schemes
// @Description 鑾峰彇鐢ㄦ埛淇℃伅
// @Tags User
// @Accept json
// @Produce json
// @Param name query string true "鐢ㄦ埛鍚?"
// @Param password query string true "瀵嗙爜"
// @Success 200 {object} models.Profile
// @Router /api/v1/user/profile [Get]
func (handler *Handler) GetUserProfile(c fiber.Ctx) error {
	name := c.Query("name", "")
	pass := c.Query("password", "")

	data, err := handler.service.GetTemplateProfile(name, pass)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}

	return common.NewResponse(c).SuccessWithData(data)
}
