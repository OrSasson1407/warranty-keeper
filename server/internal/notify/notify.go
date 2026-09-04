package notify

import "context"

// Sender delivers a single push notification to one device token. Kept
// behind an interface, same pattern as internal/ocr, so a real delivery
// backend can be swapped in without touching the scheduled job's logic.
type Sender interface {
	Send(ctx context.Context, deviceToken, title, body string) error
}
