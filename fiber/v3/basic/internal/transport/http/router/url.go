package router

import (
	modules "github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/modules"
	"github.com/gofiber/fiber/v3"
)

func api(root fiber.Router, routeModules ...modules.RouteModule) {
	v1(root.Group("/v1"), routeModules...)
}

func v1(root fiber.Router, routeModules ...modules.RouteModule) {
	for _, module := range routeModules {
		if module == nil {
			continue
		}
		module.RegisterRoutes(root)
	}
}
