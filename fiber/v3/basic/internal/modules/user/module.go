package user

import (
	"github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/modules/user/controller"
	"github.com/gofiber/fiber/v3"
)

type Module struct {
	handler *controller.Handler
}

func NewModule(handler *controller.Handler) *Module {
	return &Module{handler: handler}
}

func (module *Module) Name() string {
	return "user"
}

func (module *Module) RegisterRoutes(root fiber.Router) {
	if module == nil || module.handler == nil {
		return
	}

	userAPI := root.Group("/user")
	userAPI.Get("/profile", module.handler.GetUserProfile)
}
