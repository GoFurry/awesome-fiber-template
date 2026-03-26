package bootstrap

import (
	modules "github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/modules"
	usermodule "github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/modules/user"
	usercontroller "github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/modules/user/controller"
	userdao "github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/modules/user/dao"
	userservice "github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/modules/user/service"
)

type Application struct {
	RouteModules []modules.RouteModule
}

func buildApplication() (*Application, error) {
	userRepository := userdao.NewProfileRepository()
	userService := userservice.NewUserService(userRepository)
	userHandler := usercontroller.NewHandler(userService)

	return &Application{
		RouteModules: []modules.RouteModule{
			usermodule.NewModule(userHandler),
		},
	}, nil
}
