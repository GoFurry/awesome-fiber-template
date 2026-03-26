package user

import (
	modules "github.com/GoFurry/awesome-go-template/fiber/v3/heavy/internal/modules"
	"github.com/GoFurry/awesome-go-template/fiber/v3/heavy/internal/modules/user/controller"
	"github.com/GoFurry/awesome-go-template/fiber/v3/heavy/internal/modules/user/dao"
	usermigrations "github.com/GoFurry/awesome-go-template/fiber/v3/heavy/internal/modules/user/migrations"
	"github.com/GoFurry/awesome-go-template/fiber/v3/heavy/internal/modules/user/models"
	"github.com/GoFurry/awesome-go-template/fiber/v3/heavy/internal/modules/user/service"
	"github.com/gofiber/fiber/v3"
)

type Module struct {
	handler *controller.Handler
}

func NewModule(handler *controller.Handler) *Module {
	return &Module{handler: handler}
}

func NewBundle() (modules.Bundle, error) {
	repository := dao.NewUserRepository()
	userService := service.NewUserService(repository)
	handler := controller.NewHandler(userService)

	return modules.Bundle{
		RouteModules: []modules.RouteModule{
			NewModule(handler),
		},
		DatabaseModels: []any{
			&models.User{},
		},
		Migrations: []modules.Migration{
			usermigrations.SeedTemplateUser{},
		},
	}, nil
}

func (module *Module) Name() string {
	return "user"
}

func (module *Module) RegisterRoutes(root fiber.Router) {
	if module == nil || module.handler == nil {
		return
	}

	userAPI := root.Group("/users")
	userAPI.Post("/", module.handler.CreateUser)
	userAPI.Get("/", module.handler.ListUsers)
	userAPI.Get("/:id", module.handler.GetUser)
	userAPI.Put("/:id", module.handler.UpdateUser)
	userAPI.Delete("/:id", module.handler.DeleteUser)
}
