package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/marioidival/job-processing-system/internal/api"
	"github.com/marioidival/job-processing-system/pkg/database"
)

type Server interface {
	GetJobs(ctx echo.Context) error
	SaveJobs(ctx echo.Context) error
}

func SetupHandlers(dbc *database.Client) *echo.Echo {
	e := echo.New()
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(1000)))

	server := api.NewServer(dbc)

	registerHandler(e, server)

	return e
}

func registerHandler(router *echo.Echo, server Server) {
	router.POST("/jobs", server.SaveJobs)
	router.GET("/jobs", server.GetJobs)
}
