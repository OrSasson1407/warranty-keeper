package handlers

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"warrantykeeper/server/internal/auth"
	"warrantykeeper/server/internal/models"
)

type registerRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=8"`
	FullName   string `json:"full_name" binding:"required"`
	InviteCode string `json:"invite_code"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type authResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         models.User `json:"user"`
}

func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process password"})
		return
	}

	user := models.User{
		Email:        req.Email,
		PasswordHash: passwordHash,
		FullName:     req.FullName,
	}

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		if req.InviteCode != "" {
			var household models.Household
			if err := tx.Where("invite_code = ?", req.InviteCode).First(&household).Error; err != nil {
				return fmt.Errorf("invalid invite code")
			}

			var memberCount int64
			if err := tx.Model(&models.User{}).Where("household_id = ?", household.ID).Count(&memberCount).Error; err != nil {
				return err
			}
			if memberCount >= 2 {
				return fmt.Errorf("household already has the maximum of 2 members")
			}

			user.HouseholdID = household.ID
			return tx.Create(&user).Error
		}

		userID := uuid.New()
		user.ID = userID

		household := models.Household{
			Name:       fmt.Sprintf("הבית של %s", req.FullName),
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
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.respondWithTokens(c, http.StatusCreated, user)
}

func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	if !auth.CheckPassword(user.PasswordHash, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	h.respondWithTokens(c, http.StatusOK, user)
}

func (h *Handler) RefreshToken(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := auth.ParseRefreshToken(h.Cfg.JWTSecret, req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
		return
	}

	accessToken, err := auth.GenerateAccessToken(h.Cfg.JWTSecret, claims.UserID, claims.HouseholdID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue access token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}

func (h *Handler) respondWithTokens(c *gin.Context, status int, user models.User) {
	accessToken, err := auth.GenerateAccessToken(h.Cfg.JWTSecret, user.ID, user.HouseholdID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue access token"})
		return
	}
	refreshToken, err := auth.GenerateRefreshToken(h.Cfg.JWTSecret, user.ID, user.HouseholdID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue refresh token"})
		return
	}

	c.JSON(status, authResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	})
}

func generateInviteCode() string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // no ambiguous chars (0/O, 1/I)
	b := make([]byte, 8)
	buf := make([]byte, 8)
	_, _ = rand.Read(buf)
	for i, v := range buf {
		b[i] = alphabet[int(v)%len(alphabet)]
	}
	return string(b)
}
