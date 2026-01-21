package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	myMiddleware "github.com/Gustik/gophermart/internal/handler/middleware"
	"github.com/Gustik/gophermart/internal/service"
)

type Router struct {
	jwtSecret    string
	logger       *zap.Logger
	authHandler  *AuthHandler
	orderHandler *OrderHandler
}

func NewRouter(jwtSecret string, logger *zap.Logger, authService service.AuthService, orderService service.OrderService) *Router {
	return &Router{
		jwtSecret:    jwtSecret,
		authHandler:  NewAuthHandler(logger, authService),
		orderHandler: NewOrderHandler(logger, orderService),
	}
}

func (rt *Router) Setup() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Публичные роуты
	r.Group(func(r chi.Router) {
		r.Use(myMiddleware.RequireContentType("application/json"))

		r.Post("/api/user/register", rt.authHandler.Register)
		r.Post("/api/user/login", rt.authHandler.Login)
	})

	// Защищённые роуты
	r.Group(func(r chi.Router) {
		r.Use(myMiddleware.AuthMiddleware(rt.jwtSecret))

		r.Get("/api/user/check", rt.authHandler.Check)
		r.With(myMiddleware.RequireContentType("text/plain")).
			Post("/api/user/orders", rt.orderHandler.UploadOrder)
	})

	return r
}
