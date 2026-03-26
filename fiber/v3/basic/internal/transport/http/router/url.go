package router

import (
	user "github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/modules/user/controller"
	"github.com/gofiber/fiber/v3"
)

func api(g fiber.Router) {
	v1(g.Group("/v1"))
}

func v1(g fiber.Router) {
	userApi := g.Group("/user")
	{
		userApi.Get("/profile", user.UserApi.GetUserProfile)
	}
}
