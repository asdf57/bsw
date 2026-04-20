package router

import (
	"github.com/asdf57/bsw/internal/http/handlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(h *handlers.Handlers) (*gin.Engine, error) {
	router := gin.Default()
	router.Use(cors.Default())
	RegisterV1(router, h)
	return router, nil
}
