package router

import (
	"log/slog"
	"os"
	"time"

	"github.com/GoFurry/awesome-go-template/fiber/v3/common"
	"github.com/GoFurry/awesome-go-template/fiber/v3/middleware"
	"github.com/GoFurry/awesome-go-template/fiber/v3/roof/env"
	corazalite "github.com/GoFurry/coraza-fiber-lite"
	swagger "github.com/gofiber/contrib/v3/swaggerui"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/pprof"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

var Router *router

type router struct{}

func NewRouter() *router {
	return &router{}
}

func init() {
	Router = NewRouter()
}

func (router *router) Init() *fiber.App {
	cfg := env.GetServerConfig()
	appName := cfg.Server.AppName
	if appName == "" {
		appName = common.COMMON_PROJECT_NAME
	}

	app := fiber.New(fiber.Config{
		AppName:      appName,
		ServerHeader: appName,
		ErrorHandler: customErrorHandler,
		TrustProxy:   true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	registerMiddlewares(app)

	app.Get("/healthz", func(c fiber.Ctx) error {
		return common.NewResponse(c).SuccessWithData(fiber.Map{
			"name":    appName,
			"version": cfg.Server.AppVersion,
			"status":  "ok",
		})
	})

	api(app.Group("/api"))

	return app
}

func api(g fiber.Router) {
	v1(g.Group("/v1"))
}

func v1(g fiber.Router) {
	userApi(g.Group("/user"))
}

func registerMiddlewares(app *fiber.App) {
	cfg := env.GetServerConfig()

	app.Use(recover.New(recover.Config{
		EnableStackTrace: cfg.Server.Mode == "debug",
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.Middleware.Cors.AllowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"Content-Length"},
		MaxAge:           86400,
	}))

	if cfg.Middleware.Limiter.IsOn {
		app.Use(limiter.New(limiter.Config{
			Max:        cfg.Middleware.Limiter.MaxRequests,
			Expiration: cfg.Middleware.Limiter.Expiration * time.Second,
			KeyGenerator: func(c fiber.Ctx) string {
				return c.IP()
			},
			LimitReached: func(c fiber.Ctx) error {
				return common.NewResponse(c).ErrorWithCode("too many requests", fiber.StatusTooManyRequests)
			},
		}))
	}

	if cfg.Waf.WafSwitch {
		app.Use(corazalite.CorazaMiddleware())
	}

	if cfg.Server.Mode == "debug" {
		app.Use(pprof.New())

		if cfg.Middleware.Swagger.IsOn {
			if _, err := os.Stat(cfg.Middleware.Swagger.FilePath); os.IsNotExist(err) {
				slog.Warn("swagger file does not exist, skip swagger middleware", "file", cfg.Middleware.Swagger.FilePath)
			} else {
				app.Use(swagger.New(swagger.Config{
					BasePath: cfg.Middleware.Swagger.BasePath,
					FilePath: cfg.Middleware.Swagger.FilePath,
					Path:     cfg.Middleware.Swagger.Path,
					Title:    cfg.Middleware.Swagger.Title,
				}))
			}
		}
	}

	app.Use(middleware.PrometheusMiddleware)
	app.Get("/metrics", middleware.MetricsHandler)
}

func customErrorHandler(c fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	response := common.NewResponse(c)
	switch code {
	case fiber.StatusNotFound:
		return response.ErrorWithCode("resource not found", code)
	case fiber.StatusMethodNotAllowed:
		return response.ErrorWithCode("method not allowed", code)
	case fiber.StatusRequestTimeout:
		return response.ErrorWithCode("request timeout", code)
	default:
		if env.GetServerConfig().Server.Mode != "debug" {
			return response.ErrorWithCode("internal server error", code)
		}
		return response.ErrorWithCode(err.Error(), code)
	}
}
