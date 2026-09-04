package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const expoPushEndpoint = "https://exp.host/--/api/v2/push/send"

// ExpoSender delivers via Expo's push service, which Expo-managed apps use
// regardless of platform — it routes to FCM on Android and APNs on iOS.
// Not wired as the default sender; switch to it in cmd/notify-expiring once
// ready to actually deliver (no credentials needed for Expo's basic tier).
type ExpoSender struct {
	client *http.Client
}

func NewExpoSender() *ExpoSender {
	return &ExpoSender{client: &http.Client{Timeout: 10 * time.Second}}
}

type expoPushMessage struct {
	To    string `json:"to"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

func (s *ExpoSender) Send(ctx context.Context, deviceToken, title, body string) error {
	payload, err := json.Marshal(expoPushMessage{To: deviceToken, Title: title, Body: body})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, expoPushEndpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expo push service returned status %d", resp.StatusCode)
	}
	return nil
}
