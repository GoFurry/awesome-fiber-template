package middleware

import (
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/infra/logging"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/utils/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/infra/observability/metrics"
)

var (
	MetricsHandler    fiber.Handler
	once, routesOnce  sync.Once
	registeredRoutes  map[string]struct{}
	skipPaths         map[string]bool
	ignoreStatusCodes map[int]bool
)

type FiberPromConf struct {
	Namespace         string
	SkipPaths         []string
	IgnoreStatusCodes []int
}

func InitPrometheus(cfg ...FiberPromConf) {
	log.Debug("[InitPrometheus init] try to init prometheus middleware...")
	once.Do(func() {
		conf := FiberPromConf{}
		if len(cfg) > 0 {
			conf = cfg[0]
		}

		defaultSkipPaths := []string{"/metrics"}
		if conf.SkipPaths != nil {
			conf.SkipPaths = append(defaultSkipPaths, conf.SkipPaths...)
		} else {
			conf.SkipPaths = defaultSkipPaths
		}

		setSkipPaths(conf.SkipPaths)
		setIgnoreStatusCodes(conf.IgnoreStatusCodes)

		metrics.Init(conf.Namespace)

		registry := prometheus.DefaultRegisterer
		registry.MustRegister(metrics.HttpRequestsTotal)
		registry.MustRegister(metrics.HttpRequestDuration)
		registry.MustRegister(metrics.HttpActiveRequests)
		MetricsHandler = adaptor.HTTPHandler(promhttp.Handler())
	})
	log.Debug("[InitPrometheus init] init prometheus middleware ok.")
}

func PrometheusMiddleware(c fiber.Ctx) error {
	method := utils.CopyString(c.Method())

	metrics.HttpActiveRequests.Inc()
	defer metrics.HttpActiveRequests.Dec()

	start := time.Now()
	err := c.Next()

	routesOnce.Do(func() {
		registeredRoutes = make(map[string]struct{})
		for _, route := range c.App().GetRoutes(true) {
			path := route.Path
			if path != "" && path != "/" {
				path = normalizePath(path)
			}
			registeredRoutes[route.Method+" "+path] = struct{}{}
		}
	})

	routePath := utils.CopyString(c.Route().Path)
	if routePath == "/" {
		routePath = utils.CopyString(c.Path())
	}
	if routePath != "" && routePath != "/" {
		routePath = normalizePath(routePath)
	}

	if _, ok := registeredRoutes[method+" "+routePath]; !ok {
		log.Warn("[Try to req unregistered route] 灏濊瘯璇锋眰鏈敞鍐岀殑璺敱: ", method+" "+routePath)
		return err
	}
	if skipPaths[routePath] {
		return err
	}

	status := fiber.StatusInternalServerError
	if err != nil {
		if fiberErr, ok := err.(*fiber.Error); ok {
			status = fiberErr.Code
		}
	} else {
		status = c.Response().StatusCode()
	}
	if ignoreStatusCodes[status] {
		return err
	}

	writeNormalMetrics(method, routePath, strconv.Itoa(status), start)
	return err
}

func writeNormalMetrics(method, path, status string, start time.Time) {
	metrics.HttpRequestsTotal.WithLabelValues(method, path, status).Inc()
	metrics.HttpRequestDuration.WithLabelValues(method, path).Observe(time.Since(start).Seconds())
}

func normalizePath(routePath string) string {
	normalized := strings.TrimRight(routePath, "/")
	if normalized == "" {
		return "/"
	}
	return normalized
}

func setSkipPaths(paths []string) {
	if skipPaths == nil {
		skipPaths = make(map[string]bool)
	}
	for _, path := range paths {
		skipPaths[path] = true
	}
}

func setIgnoreStatusCodes(codes []int) {
	if ignoreStatusCodes == nil {
		ignoreStatusCodes = make(map[int]bool)
	}
	for _, code := range codes {
		ignoreStatusCodes[code] = true
	}
}
