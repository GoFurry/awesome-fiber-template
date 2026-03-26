package modules

import "github.com/gofiber/fiber/v3"

type RouteModule interface {
	Name() string
	RegisterRoutes(root fiber.Router)
}
