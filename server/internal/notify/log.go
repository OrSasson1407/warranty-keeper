package notify

import (
	"context"
	"log"
)

// LogSender just logs what it would have sent. Default for local/dev so the
// expiry job is runnable and testable end-to-end without any push
// credentials configured — swap in ExpoSender (or a real FCM client) once
// ready to actually deliver.
type LogSender struct{}

func NewLogSender() *LogSender {
	return &LogSender{}
}

func (s *LogSender) Send(_ context.Context, deviceToken, title, body string) error {
	log.Printf("[notify:stub] would push to %s — %q: %q", deviceToken, title, body)
	return nil
}
