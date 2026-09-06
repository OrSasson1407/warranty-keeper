// Package gmailsync implements the opt-in Gmail integration: exchanging an
// authorization code for tokens, and periodically scanning a connected
// user's inbox for order-confirmation emails from an allowlisted set of
// retailers, feeding matches into the same receipt-review flow a photo
// upload uses.
package gmailsync

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gmail "google.golang.org/api/gmail/v1"

	"warrantykeeper/server/internal/config"
)

// ReadonlyScope is the single Gmail permission this integration requests.
// It cannot read Sent mail or send/delete anything.
const ReadonlyScope = gmail.GmailReadonlyScope

// OAuthConfig builds the OAuth2 config for the Gmail connect flow. The
// client ID/secret are the same "Web application" OAuth client used for
// Google Sign-In (internal/handlers/google_auth.go) -- reusing it means one
// Google Cloud OAuth client covers both features, since scopes are chosen
// per-authorization-request, not per-client.
func OAuthConfig(cfg config.Config, redirectURI string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.GoogleOAuthClientID,
		ClientSecret: cfg.GoogleOAuthClientSecret,
		RedirectURL:  redirectURI,
		Scopes:       []string{ReadonlyScope},
		Endpoint:     google.Endpoint,
	}
}
