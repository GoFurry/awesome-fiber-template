package controller

import (
	"github.com/GoFurry/awesome-go-template/fiber/v3/apps/user/service"
	"github.com/GoFurry/awesome-go-template/fiber/v3/common"
	"github.com/gofiber/fiber/v3"
)

func GetProfile(c fiber.Ctx) error {
	return common.NewResponse(c).SuccessWithData(service.GetTemplateProfile())
}
