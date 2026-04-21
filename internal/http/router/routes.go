package router

import (
	"github.com/asdf57/bsw/internal/http/handlers"
	"github.com/gin-gonic/gin"
)

// Return the router group for the v1 API routes
func RegisterV1(r *gin.Engine, h *handlers.Handlers) *gin.RouterGroup {
	v1 := r.Group("/api/v1")

	payments := v1.Group("/payment")
	payments.GET("/:id", h.GetPayment)
	payments.GET("/all", h.GetAllPayments)
	payments.POST("", h.PostPayment)
	payments.DELETE("/:id", h.DeletePayment)

	users := v1.Group("/user")
	users.GET("", h.GetUsers)
	users.POST("", h.CreateUser)

	debts := v1.Group("/debts")
	debts.GET("", h.GetDebts)
	debts.GET("/debts/users", h.GetAllUserDebts)

	health := v1.Group("/health")
	health.GET("", h.GetDbHealth)

	admin := v1.Group("/admin")
	admin.POST("/backup", h.BackupDB)
	admin.GET("/backup/:filename", h.GetBackup)
	admin.GET("/exchange-rate", h.GetExchangeRate)

	return v1
}
