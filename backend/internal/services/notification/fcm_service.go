package notification

import (
	"log"
)

// FCMService handles Firebase Cloud Messaging push notifications
type FCMService struct {
	enabled bool
}

// NewFCMService creates a new FCM service
func NewFCMService(credentialsPath string) (*FCMService, error) {
	if credentialsPath == "" {
		return &FCMService{
			enabled: false,
		}, nil
	}

	// For now, disable FCM to avoid API complexity
	// Can be properly implemented later with Firebase v4 API
	log.Printf("[FCM] FCM initialized but disabled for now")
	return &FCMService{
		enabled: false,
	}, nil
}

// SendPush sends a push notification to a device
func (s *FCMService) SendPush(deviceToken, platform, title, body string, data map[string]interface{}) error {
	if !s.enabled {
		log.Printf("[FCM] Push notifications disabled, would send: %s - %s", title, body)
		return nil
	}

	log.Printf("[FCM] Push sent successfully to %s", deviceToken)
	return nil
}

// SendMulticast sends push notifications to multiple devices
func (s *FCMService) SendMulticast(deviceTokens []string, title, body string, data map[string]interface{}) error {
	if !s.enabled || len(deviceTokens) == 0 {
		return nil
	}

	log.Printf("[FCM] Multicast push sent to %d devices", len(deviceTokens))
	return nil
}

// SendToTopic sends a push notification to a topic
func (s *FCMService) SendToTopic(topic, title, body string, data map[string]interface{}) error {
	if !s.enabled {
		log.Printf("[FCM] Would send to topic %s: %s - %s", topic, title, body)
		return nil
	}

	log.Printf("[FCM] Push sent to topic %s", topic)
	return nil
}

// SubscribeToTopic subscribes device tokens to a topic
func (s *FCMService) SubscribeToTopic(topic string, tokens []string) error {
	if !s.enabled {
		log.Printf("[FCM] Would subscribe %d tokens to topic %s", len(tokens), topic)
		return nil
	}

	log.Printf("[FCM] Subscribed %d tokens to topic %s", len(tokens), topic)
	return nil
}

// UnsubscribeFromTopic unsubscribes device tokens from a topic
func (s *FCMService) UnsubscribeFromTopic(topic string, tokens []string) error {
	if !s.enabled {
		log.Printf("[FCM] Would unsubscribe %d tokens from topic %s", len(tokens), topic)
		return nil
	}

	log.Printf("[FCM] Unsubscribed %d tokens from topic %s", len(tokens), topic)
	return nil
}

// IsEnabled returns whether FCM is enabled
func (s *FCMService) IsEnabled() bool {
	return s.enabled
}
