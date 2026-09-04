package models

import "github.com/google/uuid"

// DeviceToken is an Expo push token registered by a user's device, used by
// the expiry-warning notification job. Expo's push service is the delivery
// mechanism for Expo-managed React Native apps; it routes to FCM (Android)
// and APNs (iOS) under the hood, matching the architecture doc's FCM choice
// without requiring native Firebase project wiring in the app itself.
type DeviceToken struct {
	BaseModel
	UserID        uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	ExpoPushToken string    `gorm:"uniqueIndex;not null" json:"expo_push_token"`
}
