package controller

import (
	"github.com/GoFurry/awesome-go-template/fiber/v3/apps/user/service"
	"github.com/GoFurry/awesome-go-template/fiber/v3/common"
	"github.com/gofiber/fiber/v3"
)

type userApi struct{}

var UserApi *userApi

func init() {
	UserApi = &userApi{}
}

// @Summary 获取用户信息
// @Schemes
// @Description 获取用户信息
// @Tags User
// @Accept json
// @Produce json
// @Param name query string true "用户名"
// @Param password query string true "密码"
// @Success 200 {object} models.Profile
// @Router /api/v1/user/profile [Get]
func (api *userApi) GetUserProfile(c fiber.Ctx) error {
	num := c.Query("name", "")
	pass := c.Query("password", "")
	data, err := service.GetUserService().GetTemplateProfile(num, pass)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}

	return common.NewResponse(c).SuccessWithData(data)
}
