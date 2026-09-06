package gmailsync

import (
	"context"
	"encoding/base64"
	"log"
	"net/mail"
	"time"

	"golang.org/x/oauth2"
	"gorm.io/gorm"

	gmail "google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	"warrantykeeper/server/internal/config"
	"warrantykeeper/server/internal/crypto"
	"warrantykeeper/server/internal/models"
)

// LookbackDays bounds how far back a scan searches, keeping each run cheap
// and avoiding re-scanning a user's entire mailbox history every time.
const LookbackDays = 14

// candidateMessage is what a scan needs from one email, independent of the
// Gmail API's own types -- the seam that lets RunScan be unit tested
// against a fake messageSource instead of a real Gmail account.
type candidateMessage struct {
	ID      string
	From    string
	Subject string
	Date    time.Time
	Body    string
}

type messageSource interface {
	ListCandidates(ctx context.Context) ([]string, error)
	FetchMessage(ctx context.Context, id string) (candidateMessage, error)
}

// newGmailSource is a seam over constructing a real Gmail-backed
// messageSource, swapped out in tests since a real inbox can't be faked.
var newGmailSource = func(ctx context.Context, ts oauth2.TokenSource) (messageSource, error) {
	svc, err := gmail.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, err
	}
	return &liveGmailSource{svc: svc}, nil
}

type liveGmailSource struct {
	svc *gmail.Service
}

func (s *liveGmailSource) ListCandidates(ctx context.Context) ([]string, error) {
	resp, err := s.svc.Users.Messages.List("me").
		Q(GmailSearchQuery(LookbackDays)).
		MaxResults(25).
		Context(ctx).
		Do()
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(resp.Messages))
	for _, m := range resp.Messages {
		ids = append(ids, m.Id)
	}
	return ids, nil
}

func (s *liveGmailSource) FetchMessage(ctx context.Context, id string) (candidateMessage, error) {
	msg, err := s.svc.Users.Messages.Get("me", id).Format("full").Context(ctx).Do()
	if err != nil {
		return candidateMessage{}, err
	}

	var from, subject, dateHeader string
	if msg.Payload != nil {
		for _, h := range msg.Payload.Headers {
			switch h.Name {
			case "From":
				from = h.Value
			case "Subject":
				subject = h.Value
			case "Date":
				dateHeader = h.Value
			}
		}
	}

	date := time.Now()
	if parsed, err := mail.ParseDate(dateHeader); err == nil {
		date = parsed
	} else if msg.InternalDate > 0 {
		date = time.UnixMilli(msg.InternalDate)
	}

	body := extractPlainTextBody(msg.Payload)
	if body == "" {
		body = msg.Snippet
	}

	return candidateMessage{ID: id, From: from, Subject: subject, Date: date, Body: body}, nil
}

func extractPlainTextBody(part *gmail.MessagePart) string {
	if part == nil {
		return ""
	}
	if part.MimeType == "text/plain" && part.Body != nil && part.Body.Data != "" {
		if decoded, ok := decodeGmailBase64(part.Body.Data); ok {
			return decoded
		}
	}
	for _, p := range part.Parts {
		if body := extractPlainTextBody(p); body != "" {
			return body
		}
	}
	return ""
}

func decodeGmailBase64(data string) (string, bool) {
	if decoded, err := base64.RawURLEncoding.DecodeString(data); err == nil {
		return string(decoded), true
	}
	if decoded, err := base64.URLEncoding.DecodeString(data); err == nil {
		return string(decoded), true
	}
	return "", false
}

// ScanCounts summarizes one RunScan invocation for logging.
type ScanCounts struct {
	ConnectionsScanned int
	MessagesMatched    int
	ReceiptsCreated    int
}

// RunScan scans every connected Gmail account for new allowlisted
// order-confirmation emails and creates a pending Receipt (Source: gmail)
// for each one not already seen, exactly mirroring what a photo upload
// creates -- so both flows land in the same confirm-and-edit review queue.
func RunScan(gdb *gorm.DB, cfg config.Config, now time.Time) (ScanCounts, error) {
	var counts ScanCounts

	var connections []models.GmailConnection
	if err := gdb.Find(&connections).Error; err != nil {
		return counts, err
	}

	ctx := context.Background()
	for i := range connections {
		conn := &connections[i]
		counts.ConnectionsScanned++

		matched, created, err := scanOneConnection(ctx, gdb, cfg, conn, now)
		if err != nil {
			log.Printf("gmail scan failed for connection %s: %v", conn.ID, err)
			continue
		}
		counts.MessagesMatched += matched
		counts.ReceiptsCreated += created
	}

	return counts, nil
}

func scanOneConnection(ctx context.Context, gdb *gorm.DB, cfg config.Config, conn *models.GmailConnection, now time.Time) (matched, created int, err error) {
	accessToken, err := crypto.Decrypt(conn.EncryptedAccessToken, cfg.TokenEncryptionKey)
	if err != nil {
		return 0, 0, err
	}
	refreshToken, err := crypto.Decrypt(conn.EncryptedRefreshToken, cfg.TokenEncryptionKey)
	if err != nil {
		return 0, 0, err
	}

	oauthCfg := OAuthConfig(cfg, "")
	token := &oauth2.Token{AccessToken: accessToken, RefreshToken: refreshToken, Expiry: conn.TokenExpiry}
	ts := oauthCfg.TokenSource(ctx, token)

	source, err := newGmailSource(ctx, ts)
	if err != nil {
		return 0, 0, err
	}

	ids, err := source.ListCandidates(ctx)
	if err != nil {
		return 0, 0, err
	}

	var user models.User
	if err := gdb.First(&user, "id = ?", conn.UserID).Error; err != nil {
		return 0, 0, err
	}

	for _, id := range ids {
		var existing models.Receipt
		if err := gdb.Where("gmail_message_id = ?", id).First(&existing).Error; err == nil {
			continue // already processed in an earlier scan
		}

		msg, err := source.FetchMessage(ctx, id)
		if err != nil {
			log.Printf("gmail fetch failed for message %s: %v", id, err)
			continue
		}

		vendor, ok := MatchVendor(msg.From)
		if !ok {
			continue
		}
		matched++

		parsed := ParseOrderEmail(msg.From, msg.Subject, msg.Date, msg.Body)
		receipt := models.Receipt{
			HouseholdID:    user.HouseholdID,
			Status:         models.ReceiptStatusPending,
			Source:         models.ReceiptSourceGmail,
			GmailMessageID: id,
			RawOCRText:     parsed.Snippet,
			ParsedVendor:   vendor,
			ParsedDate:     parsed.Date,
			ParsedAmount:   parsed.Amount,
			Confidence:     parsed.Confidence,
		}
		if err := gdb.Create(&receipt).Error; err != nil {
			log.Printf("failed to save gmail receipt for message %s: %v", id, err)
			continue
		}
		created++
	}

	persistRefreshedToken(gdb, cfg, conn, ts, accessToken)
	gdb.Model(conn).Update("last_scan_at", now)

	return matched, created, nil
}

// persistRefreshedToken saves a new access token (and refresh token, if
// Google rotated it) back to the connection row after the oauth2 library's
// TokenSource silently refreshed it during the scan above.
func persistRefreshedToken(gdb *gorm.DB, cfg config.Config, conn *models.GmailConnection, ts oauth2.TokenSource, previousAccessToken string) {
	newTok, err := ts.Token()
	if err != nil || newTok.AccessToken == previousAccessToken {
		return
	}

	encAccess, err := crypto.Encrypt(newTok.AccessToken, cfg.TokenEncryptionKey)
	if err != nil {
		log.Printf("failed to encrypt refreshed access token for connection %s: %v", conn.ID, err)
		return
	}
	updates := map[string]any{
		"encrypted_access_token": encAccess,
		"token_expiry":           newTok.Expiry,
	}
	if newTok.RefreshToken != "" {
		if encRefresh, err := crypto.Encrypt(newTok.RefreshToken, cfg.TokenEncryptionKey); err == nil {
			updates["encrypted_refresh_token"] = encRefresh
		}
	}
	gdb.Model(conn).Updates(updates)
}
