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

func (h *Handler) Init() *gin.Engine {
	router := gin.Default()
	// Init HTML Glob
	h.initHTMLGlob(router)
	// Init general routes.
	h.initGeneralRoutes(router)
	// Init API
	h.initAPI(router)

	return router
}

func (h *Handler) initHTMLGlob(router *gin.Engine) {
	// template: https://gin-gonic.com/docs/examples/html-rendering/
	router.LoadHTMLGlob("./web/template/*")
}

func (h *Handler) initAPI(router *gin.Engine) {
	// Init API v1 h.
	handlersV1 := v1.NewHandlerAPIv1(h.services)
	api := router.Group("/api")
	{
		handlersV1.Init(api)
	}
}
