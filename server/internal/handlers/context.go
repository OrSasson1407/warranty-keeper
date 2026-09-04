package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"warrantykeeper/server/internal/middleware"
)

func householdID(c *gin.Context) uuid.UUID {
	return c.MustGet(middleware.CtxHouseholdID).(uuid.UUID)
}

func userID(c *gin.Context) uuid.UUID {
	return c.MustGet(middleware.CtxUserID).(uuid.UUID)
}
