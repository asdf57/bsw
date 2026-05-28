package router

import (
	"github.com/asdf57/bsw/internal/http/handlers"
	"github.com/gin-gonic/gin"
)

// Return the router group for the v1 API routes
func RegisterV1(r *gin.Engine, h *handlers.Handlers) *gin.RouterGroup {
	v1 := r.Group("/api/v1")

	payments := v1.Group("/payment")
	payments.GET("/all", h.GetAllPayments)
	payments.GET("/tag/:tag", h.GetPaymentsByTag)
	payments.GET("/tags", h.GetPaymentsByTags)
	payments.GET("/:id", h.GetPayment)
	payments.POST("", h.PostPayment)
	payments.PUT("/:id", h.PutPayment)
	payments.DELETE("/:id", h.DeletePayment)

	tags := v1.Group("/tags")
	tags.GET("", h.GetTags)
	tags.POST("", h.CreateTag)

	users := v1.Group("/user")
	users.GET("", h.GetUsers)
	users.POST("", h.CreateUser)
	users.DELETE("/:id", h.DeleteUser)

	debts := v1.Group("/debts")
	debts.GET("", h.GetDebts)
	debts.GET("/debts/users", h.GetAllUserDebts)
	debts.GET("/settlements", h.GetSettlements)
	debts.POST("/settle", h.SettleDebts)
	debts.POST("/settlements/:id/reverse", h.ReverseSettlement)

	stats := v1.Group("/stats")
	stats.GET("/user", h.GetUserStats)

	health := v1.Group("/health")
	health.GET("", h.GetDbHealth)

	admin := v1.Group("/admin")
	admin.POST("/backup", h.BackupDB)
	admin.GET("/backup/:filename", h.GetBackup)
	admin.GET("/exchange-rate", h.GetExchangeRate)
	admin.GET("/exchange-rates", h.GetExchangeRatesFromDB)

	return v1
}
