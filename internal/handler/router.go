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
	jwtSecret   string
	logger      *zap.Logger
	authHandler *AuthHandler
}

func NewRouter(jwtSecret string, logger *zap.Logger, authService *service.AuthService) *Router {
	return &Router{
		jwtSecret:   jwtSecret,
		authHandler: NewAuthHandler(logger, authService),
	}
}

func (rt *Router) Setup() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	r.Post("/api/user/register", rt.authHandler.Register)
	r.Post("/api/user/login", rt.authHandler.Login)

	// Защищённые
	r.Group(func(r chi.Router) {
		r.Use(myMiddleware.AuthMiddleware(rt.jwtSecret))
		r.Get("/api/user/check", rt.authHandler.Check)
	})

	return r
}
