package router

import (
	"github.com/GoFurry/awesome-go-template/fiber/v3/apps/user/controller"
	"github.com/gofiber/fiber/v3"
)

func userApi(g fiber.Router) {
	g.Get("/profile", controller.GetProfile)
}
