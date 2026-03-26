package router

import (
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"

	env "github.com/GoFurry/awesome-go-template/fiber/v3/basic/config"
	modules "github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/modules"
	"github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/transport/http/middleware"
	"github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/transport/http/webui"
	"github.com/GoFurry/awesome-go-template/fiber/v3/basic/pkg/common"
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

func (router *router) Init(routeModules ...modules.RouteModule) *fiber.App {
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

	api(app.Group("/api"), routeModules...)

	if cfg.Server.IsFullStack {
		attachEmbeddedUI(app)
	}

	return app
}

func attachEmbeddedUI(app *fiber.App) {
	uiFS, err := fs.Sub(webui.FS, "dist")
	if err != nil {
		return
	}
	index, err := fs.ReadFile(uiFS, "index.html")
	if err != nil {
		return
	}

	sendIndex := func(c fiber.Ctx) error {
		c.Type("html", "utf-8")
		return c.Send(index)
	}

	app.Use(func(c fiber.Ctx) error {
		if c.Method() != fiber.MethodGet && c.Method() != fiber.MethodHead {
			return fiber.ErrNotFound
		}

		reqPath := c.Path()
		if reqPath == "/api" || strings.HasPrefix(reqPath, "/api/") || reqPath == "/v1" || strings.HasPrefix(reqPath, "/v1/") {
			return fiber.ErrNotFound
		}

		if reqPath == "/" || reqPath == "" {
			return sendIndex(c)
		}

		cleaned := path.Clean(reqPath)
		cleaned = strings.TrimPrefix(cleaned, "/")
		if cleaned == "." || cleaned == "" {
			return sendIndex(c)
		}

		if stat, err := fs.Stat(uiFS, cleaned); err == nil && !stat.IsDir() {
			return c.SendFile(cleaned, fiber.SendFile{FS: uiFS})
		}

		return sendIndex(c)
	})
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

	if cfg.Middleware.Limiter.Enabled {
		app.Use(limiter.New(limiter.Config{
			Max:        cfg.Middleware.Limiter.MaxRequests,
			Expiration: cfg.Middleware.Limiter.Expiration * time.Second,
			KeyGenerator: func(c fiber.Ctx) string {
				return c.IP()
			},
			LimitReached: func(c fiber.Ctx) error {
				return common.NewResponse(c).ErrorWithCode(common.NewValidationError("too many requests"), fiber.StatusTooManyRequests)
			},
		}))
	}

	if cfg.Waf.Enabled {
		app.Use(corazalite.CorazaMiddleware())
	}

	if cfg.Server.Mode == "debug" {
		app.Use(pprof.New())

		if cfg.Middleware.Swagger.Enabled {
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

	if cfg.Prometheus.Enabled {
		app.Use(middleware.PrometheusMiddleware)
		app.Get(cfg.Prometheus.Path, middleware.MetricsHandler)
	}
}

func customErrorHandler(c fiber.Ctx, err error) error {
	var appErr common.Error
	if errors.As(err, &appErr) {
		return common.NewResponse(c).ErrorWithCode(appErr, appErr.GetHTTPStatus())
	}

	code := fiber.StatusInternalServerError
	if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
		code = fiberErr.Code
	}

	response := common.NewResponse(c)
	switch code {
	case fiber.StatusNotFound:
		return response.ErrorWithCode(common.NewError(common.RETURN_FAILED, code, "resource not found"), code)
	case fiber.StatusMethodNotAllowed:
		return response.ErrorWithCode(common.NewError(common.RETURN_FAILED, code, "method not allowed"), code)
	case fiber.StatusRequestTimeout:
		return response.ErrorWithCode(common.NewError(common.RETURN_FAILED, code, "request timeout"), code)
	default:
		if env.GetServerConfig().Server.Mode != "debug" {
			return response.ErrorWithCode(common.NewError(common.RETURN_FAILED, code, "internal server error"), code)
		}
		return response.ErrorWithCode(common.NewError(common.RETURN_FAILED, code, err.Error()), code)
	}
}
