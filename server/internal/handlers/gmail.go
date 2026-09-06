package handlers

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	gmail "google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	"warrantykeeper/server/internal/crypto"
	"warrantykeeper/server/internal/gmailsync"
	"warrantykeeper/server/internal/models"
)

type connectGmailRequest struct {
	Code         string `json:"code" binding:"required"`
	RedirectURI  string `json:"redirect_uri" binding:"required"`
	CodeVerifier string `json:"code_verifier"`
}

type gmailStatusResponse struct {
	Connected    bool       `json:"connected"`
	GmailAddress string     `json:"gmail_address,omitempty"`
	LastScanAt   *time.Time `json:"last_scan_at"`
}

// ExchangeGmailCode is a seam over the OAuth authorization-code exchange
// plus a Gmail profile lookup: a real authorization code can only come from
// a real Google consent screen, so tests swap this for a fake to exercise
// ConnectGmail's account-linking logic without one.
var ExchangeGmailCode = func(ctx context.Context, h *Handler, redirectURI, code, codeVerifier string) (*oauth2.Token, string, error) {
	oauthCfg := gmailsync.OAuthConfig(h.Cfg, redirectURI)
	var opts []oauth2.AuthCodeOption
	if codeVerifier != "" {
		opts = append(opts, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	}

	token, err := oauthCfg.Exchange(ctx, code, opts...)
	if err != nil {
		return nil, "", err
	}

	svc, err := gmail.NewService(ctx, option.WithTokenSource(oauthCfg.TokenSource(ctx, token)))
	if err != nil {
		return nil, "", err
	}
	profile, err := svc.Users.GetProfile("me").Context(ctx).Do()
	if err != nil {
		return nil, "", err
	}
	return token, profile.EmailAddress, nil
}

// ConnectGmail exchanges an authorization code (obtained on-device via the
// mobile app's own Google consent flow) for tokens, then stores them
// encrypted against the current user. Opt-in only, per household member --
// this never touches anyone else's mailbox, and connecting again just
// re-links the same user's account with a fresh token.
func (h *Handler) ConnectGmail(c *gin.Context) {
	if h.Cfg.GoogleOAuthClientID == "" || h.Cfg.GoogleOAuthClientSecret == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Gmail integration is not configured"})
		return
	}

	var req connectGmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, gmailAddress, err := ExchangeGmailCode(c.Request.Context(), h, req.RedirectURI, req.Code, req.CodeVerifier)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to connect Gmail account"})
		return
	}
	if token.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Google did not grant offline access; remove WarrantyKeeper's access at myaccount.google.com/permissions and try connecting again"})
		return
	}

	encAccess, err := crypto.Encrypt(token.AccessToken, h.Cfg.TokenEncryptionKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store Gmail connection"})
		return
	}
	encRefresh, err := crypto.Encrypt(token.RefreshToken, h.Cfg.TokenEncryptionKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store Gmail connection"})
		return
	}

	uid := userID(c)
	var conn models.GmailConnection
	err = h.DB.Where("user_id = ?", uid).First(&conn).Error
	conn.UserID = uid
	conn.GmailAddress = gmailAddress
	conn.EncryptedAccessToken = encAccess
	conn.EncryptedRefreshToken = encRefresh
	conn.TokenExpiry = token.Expiry

	if err == nil {
		err = h.DB.Save(&conn).Error
	} else {
		err = h.DB.Create(&conn).Error
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store Gmail connection"})
		return
	}

	c.JSON(http.StatusOK, gmailStatusResponse{Connected: true, GmailAddress: conn.GmailAddress, LastScanAt: conn.LastScanAt})
}

func (h *Handler) GmailStatus(c *gin.Context) {
	var conn models.GmailConnection
	if err := h.DB.Where("user_id = ?", userID(c)).First(&conn).Error; err != nil {
		c.JSON(http.StatusOK, gmailStatusResponse{Connected: false})
		return
	}
	c.JSON(http.StatusOK, gmailStatusResponse{Connected: true, GmailAddress: conn.GmailAddress, LastScanAt: conn.LastScanAt})
}

// DisconnectGmail deletes the stored connection and makes a best-effort
// attempt to revoke the token at Google so WarrantyKeeper no longer shows
// up under the user's "Third-party access" list even before the token
// would otherwise expire. A revoke failure doesn't block disconnecting --
// deleting our own copy of the token is what actually matters locally.
func (h *Handler) DisconnectGmail(c *gin.Context) {
	var conn models.GmailConnection
	if err := h.DB.Where("user_id = ?", userID(c)).First(&conn).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"disconnected": true})
		return
	}

	if accessToken, err := crypto.Decrypt(conn.EncryptedAccessToken, h.Cfg.TokenEncryptionKey); err == nil {
		RevokeGoogleOAuthToken(c.Request.Context(), accessToken)
	}

	if err := h.DB.Delete(&conn).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to disconnect Gmail account"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"disconnected": true})
}

// RevokeGoogleOAuthToken is a seam over the best-effort revoke call to
// Google, swapped out in tests to avoid a real network call for a token
// that was never issued by Google in the first place.
var RevokeGoogleOAuthToken = func(ctx context.Context, token string) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://oauth2.googleapis.com/revoke?token="+url.QueryEscape(token), nil)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}
