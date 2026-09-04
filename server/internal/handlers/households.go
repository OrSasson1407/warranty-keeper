package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"warrantykeeper/server/internal/models"
)

type householdMemberResponse struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

type householdResponse struct {
	ID         string                     `json:"id"`
	Name       string                     `json:"name"`
	InviteCode string                     `json:"invite_code"`
	Members    []householdMemberResponse  `json:"members"`
}

// GetMyHousehold returns the caller's household with its member list, so the
// mobile settings screen can show "household name + members + invite code"
// (UX doc screen 8) in one call.
func (h *Handler) GetMyHousehold(c *gin.Context) {
	var household models.Household
	if err := h.DB.First(&household, "id = ?", householdID(c)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "household not found"})
		return
	}

	var users []models.User
	if err := h.DB.Where("household_id = ?", householdID(c)).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load household members"})
		return
	}

	members := make([]householdMemberResponse, 0, len(users))
	for _, u := range users {
		members = append(members, householdMemberResponse{ID: u.ID.String(), FullName: u.FullName, Email: u.Email})
	}

	c.JSON(http.StatusOK, householdResponse{
		ID:         household.ID.String(),
		Name:       household.Name,
		InviteCode: household.InviteCode,
		Members:    members,
	})
}
