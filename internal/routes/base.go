package routes

import (
	"finance_tracker/internal/logger"
	"finance_tracker/internal/service/user"
	"finance_tracker/internal/utils"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/google/uuid"
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
			AllowHeaders: []string{fiber.HeaderAuthorization, fiber.HeaderContentType},
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

	auth := s.httpEngine.Group("/api/v1", func(ctx fiber.Ctx) error {
		// extracts and validates the bearer token from the Authorization header,
		// returning the parsed user UUID. Returns a non-nil error with an appropriate HTTP
		// status code embedded if validation fails.
		token, ok := bearerToken(ctx.Get("Authorization"))
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, "missing token")
		}
		usr, err := utils.ParseAuthToken(token)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}
		userID, err := uuid.Parse(usr.ID)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid user id")
		}
		ctx.Locals("userID", userID)
		return ctx.Next()
	})
	auth.Get("/me", withUserID(s.handleCurrentUser))
	auth.Post("/category", withUserID(s.handleCreateCategory))
	auth.Put("/category/:id<int>", withUserID(s.handleUpdateCategory))
	auth.Delete("/category/:id<int>", withUserID(s.handleDeleteCategory))
}

// Run starts the HTTP Server.
func (s *Router) Run() error {
	s.log.Info("Starting HTTP server", logger.WithString("port", s.appAddr))
	return s.httpEngine.Listen(s.appAddr, fiber.ListenConfig{DisableStartupMessage: true})
}

func (s *Router) Stop() error {
	return s.httpEngine.Shutdown()
}

type authHandler func(ctx fiber.Ctx, userID uuid.UUID) error

func withUserID(h authHandler) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		userID, ok := ctx.Locals("userID").(uuid.UUID)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, "missing user id")
		}

		return h(ctx, userID)
	}
}
