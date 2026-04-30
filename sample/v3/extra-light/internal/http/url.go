package http

import "github.com/gofiber/fiber/v3"

func api(root fiber.Router) {
	v1(root.Group("/v1"))
}

func v1(root fiber.Router) {
}
