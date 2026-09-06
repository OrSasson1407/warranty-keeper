package api

import (
	"github.com/gin-gonic/gin"

	"warrantykeeper/server/internal/handlers"
	"warrantykeeper/server/internal/middleware"
)

func NewRouter(h *handlers.Handler) *gin.Engine {
	router := gin.Default()
	router.Use(middleware.CORS())

	router.Static("/uploads", h.Cfg.UploadsDir)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "env": h.Cfg.Env})
	})

	authLimiter := middleware.RateLimit(1, 10) // ~1 req/sec sustained, burst 10, per IP

	auth := router.Group("/auth", authLimiter)
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.RefreshToken)
	}

	api := router.Group("/", middleware.RequireAuth(h.Cfg.JWTSecret))
	{
		api.GET("/households/me", h.GetMyHousehold)
		api.POST("/households/me/upgrade", h.UpgradeHousehold)
		api.POST("/devices", h.RegisterDevice)

		receipts := api.Group("/receipts", authLimiter)
		receipts.POST("", h.UploadReceipt)
		api.GET("/receipts/:id", h.GetReceipt)

		api.GET("/warranty-rules/resolve", h.ResolveWarranty)
		api.GET("/manufacturer-contacts", h.ListManufacturerContacts)

		api.POST("/products", h.CreateProduct)
		api.GET("/products", h.ListProducts)
		api.GET("/products/:id", h.GetProduct)
		api.PUT("/products/:id", h.UpdateProduct)
		api.POST("/products/:id/claims", h.CreateClaim)
		api.GET("/products/:id/claims", h.ListClaims)
		api.POST("/products/:id/costs", h.CreateProductCost)
		api.GET("/products/:id/costs", h.ListProductCosts)
		api.POST("/products/:id/warranty-report", h.ReportWarrantyRule)
	}

	return router
}
