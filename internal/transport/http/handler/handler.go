package handler

import (
	"github.com/gin-gonic/gin"
	"service-account/internal/service"
	"service-account/internal/transport/http/handler/api/v1"
)

type Handler struct {
	services *service.Services
}

func NewHandler(services *service.Services) *Handler {
	return &Handler{
		services: services,
	}
}

func (handler *Handler) Init() *gin.Engine {
	router := gin.Default()
	// Init HTML Glob
	handler.initHTMLGlob(router)
	// Init general routes.
	handler.initGeneralRoutes(router)
	// Init API
	handler.initAPI(router)

	return router
}

func (handler *Handler) initHTMLGlob(router *gin.Engine) {
	// template: https://gin-gonic.com/docs/examples/html-rendering/
	router.LoadHTMLGlob("./web/template/*")
}

func (handler *Handler) initAPI(router *gin.Engine) {
	// Init API v1 handler.
	handlersV1 := v1.NewHandlerAPIv1(handler.services)
	api := router.Group("/api")
	{
		handlersV1.Init(api)
	}
}
