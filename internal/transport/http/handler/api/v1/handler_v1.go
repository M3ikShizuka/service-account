package v1

import (
	"github.com/gin-gonic/gin"
	"service-account/internal/service"
)

type HandlerAPIv1 struct {
	services *service.Services
}

func NewHandlerAPIv1(services *service.Services) *HandlerAPIv1 {
	return &HandlerAPIv1{
		services: services,
	}
}

func (h *HandlerAPIv1) Init(router *gin.RouterGroup) {
	v1 := router.Group("/v1")
	{
		h.initHandlersAccountManagment(v1)
	}
}

func (h *HandlerAPIv1) initHandlersAccountManagment(router *gin.RouterGroup) {
	router.GET(pathUser, handlerUserGet)
	router.POST(pathUser, handlerUserPost)
}
