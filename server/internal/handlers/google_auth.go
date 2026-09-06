package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"

	"warrantykeeper/server/internal/models"
)

type googleLoginRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

// ValidateGoogleIDToken is a seam over idtoken.Validate: real Google ID
// tokens are signed by Google and can't be fabricated in a test, so tests
// swap this exported var for a fake to exercise GoogleLogin's account
// resolution logic without a live token. Production wiring never touches
// this -- it's already idtoken.Validate by default.
var ValidateGoogleIDToken = idtoken.Validate

func googleClaim(payload *idtoken.Payload, key string) string {
	v, _ := payload.Claims[key].(string)
	return v
}

// GoogleLogin adds Google as a second login method alongside email+password
// (see the "Google OAuth as a second login method" issue) -- it doesn't
// replace the existing register/login flow. The mobile client obtains a
// Google ID token itself (expo-auth-session); this endpoint only verifies
// that token server-side and issues our own access/refresh tokens, the same
// way Register and Login already do via respondWithTokens.
//
// Account resolution order: existing Google-linked user -> existing
// email+password user with a matching email (silently linked) -> brand new
// user + household, mirroring Register's no-invite-code path.
func (h *Handler) GoogleLogin(c *gin.Context) {
	if h.Cfg.GoogleOAuthClientID == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Google sign-in is not configured"})
		return
	}

	var req googleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payload, err := ValidateGoogleIDToken(c.Request.Context(), req.IDToken, h.Cfg.GoogleOAuthClientID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid Google ID token"})
		return
	}

	googleID := googleClaim(payload, "sub")
	email := googleClaim(payload, "email")
	fullName := googleClaim(payload, "name")
	if googleID == "" || email == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Google token missing required claims"})
		return
	}

	var user models.User
	if err := h.DB.Where("google_id = ?", googleID).First(&user).Error; err == nil {
		h.respondWithTokens(c, http.StatusOK, user)
		return
	}

	if err := h.DB.Where("email = ?", email).First(&user).Error; err == nil {
		user.GoogleID = googleID
		if err := h.DB.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to link Google account"})
			return
		}
		h.respondWithTokens(c, http.StatusOK, user)
		return
	}

	userID := uuid.New()
	user = models.User{Email: email, GoogleID: googleID, FullName: fullName}
	user.ID = userID
	err = h.DB.Transaction(func(tx *gorm.DB) error {
		household := models.Household{
			Name:       fmt.Sprintf("הבית של %s", fullName),
			CreatedBy:  userID,
			InviteCode: generateInviteCode(),
		}
		if err := tx.Create(&household).Error; err != nil {
			return err
		}
		user.HouseholdID = household.ID
		return tx.Create(&user).Error
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create account"})
		return
	}

	h.respondWithTokens(c, http.StatusCreated, user)
}
