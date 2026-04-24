package routes

import (
	"finance_tracker/internal/logger"
	"finance_tracker/internal/service/user"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Router struct {
	appAddr    string
	log        logger.AppLogger
	service    *user.Service
	httpEngine *fiber.App
}

// InitAppRouter initializes the HTTP Server.
func InitAppRouter(log logger.AppLogger, service *user.Service, address string, enableTelemetry bool) *Router {
	app := &Router{
		appAddr:    address,
		httpEngine: fiber.New(fiber.Config{}),
		service:    service,
		log:        log.With(logger.WithService("http")),
	}
	app.httpEngine.Use(recover.New())
	if uiOrigin := service.UICORSOrigin(); uiOrigin != "" {
		app.httpEngine.Use(cors.New(cors.Config{
			AllowOrigins: []string{uiOrigin},
			AllowHeaders: []string{fiber.HeaderAuthorization},
		}))
	}
	if enableTelemetry {
		reg := prometheus.NewRegistry()
		reg.MustRegister(
			collectors.NewGoCollector(),
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
			collectors.NewBuildInfoCollector(),
		)
		app.httpEngine.Get("/metrics", adaptor.HTTPHandler(promhttp.HandlerFor(reg, promhttp.HandlerOpts{})))
	}
	app.initRoutes()
	return app
}

func (s *Router) initRoutes() {
	s.httpEngine.Get("/", func(ctx fiber.Ctx) error {
		return ctx.SendString("pong")
	})

	s.httpEngine.Get("/api/auth/google/login", s.handleGoogleLogin)
	s.httpEngine.Get("/api/auth/google/callback", s.handleGoogleCallback)
	s.httpEngine.Get("/api/auth/me", s.handleCurrentUser)
}

// Run starts the HTTP Server.
func (s *Router) Run() error {
	s.log.Info("Starting HTTP server", logger.WithString("port", s.appAddr))
	return s.httpEngine.Listen(s.appAddr)
}

func (s *Router) Stop() error {
	return s.httpEngine.Shutdown()
}
