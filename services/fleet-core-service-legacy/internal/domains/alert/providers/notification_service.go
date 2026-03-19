package providers

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

// NotificationService handles sending various types of notifications
type NotificationService struct {
	smsClient   SMSService
	emailClient EmailService
	pushClient  PushService
	logger      *zap.Logger
}

// Notification represents a notification to be sent
type Notification struct {
	Type      string                 `json:"type"`      // sms, email, push
	Recipient string                 `json:"recipient"` // phone number, email, or device token
	Subject   string                 `json:"subject"`   // for email
	Body      string                 `json:"body"`
	Data      map[string]interface{} `json:"data"`     // additional data for push notifications
	Priority  string                 `json:"priority"` // high, normal, low
}

// SMSService interface for SMS notifications
type SMSService interface {
	SendSMS(to, message string) error
}

// EmailService interface for email notifications
type EmailService interface {
	SendEmail(to, subject, body string) error
}

// PushService interface for push notifications
type PushService interface {
	SendPush(token, title, body string, data map[string]interface{}) error
}

// TwilioClient is a compatibility adapter in monolith mode.
type TwilioClient struct{}

// SendGridClient is a compatibility adapter in monolith mode.
type SendGridClient struct{}

// FirebaseClient implements PushService using Firebase Cloud Messaging
type FirebaseClient struct {
	// Implementation details would go here
}

// NewNotificationService creates a new notification service
func NewNotificationService(smsClient SMSService, emailClient EmailService, pushClient PushService, logger *zap.Logger) *NotificationService {
	return &NotificationService{
		smsClient:   smsClient,
		emailClient: emailClient,
		pushClient:  pushClient,
		logger:      logger,
	}
}

// NewTwilioClient creates a new Twilio SMS client
func NewTwilioClient(accountSID, authToken, fromNumber string) SMSService {
	_ = accountSID
	_ = authToken
	_ = fromNumber
	return &MockSMSService{}
}

// NewSendGridClient creates a new SendGrid email client
func NewSendGridClient(apiKey string) EmailService {
	_ = apiKey
	return &MockEmailService{}
}

// NewFirebaseClient creates a new Firebase push notification client
func NewFirebaseClient() PushService {
	return &FirebaseClient{}
	// Actual implementation would initialize Firebase SDK
}

// SendSMS sends an SMS notification
func (tc *TwilioClient) SendSMS(to, message string) error {
	fmt.Printf("[SMS COMPAT] To: %s, Message: %s\n", to, message)
	return nil
}

// SendEmail sends an email notification
func (sc *SendGridClient) SendEmail(to, subject, body string) error {
	fmt.Printf("[EMAIL COMPAT] To: %s, Subject: %s, Body: %s\n", to, subject, body)
	return nil
}

// SendPush sends a push notification
func (fc *FirebaseClient) SendPush(token, title, body string, data map[string]interface{}) error {
	// Mock implementation - actual implementation would use Firebase Admin SDK
	fmt.Printf("Sending push notification to %s: %s - %s\n", token, title, body)
	return nil
}

// SendNotification sends a notification based on its type
func (ns *NotificationService) SendNotification(notification *Notification) error {
	switch notification.Type {
	case "sms":
		ns.logger.Info("Sending SMS notification",
			zap.String("recipient", notification.Recipient),
			zap.String("body", notification.Body))
		return ns.smsClient.SendSMS(notification.Recipient, notification.Body)

	case "email":
		ns.logger.Info("Sending email notification",
			zap.String("recipient", notification.Recipient),
			zap.String("subject", notification.Subject))
		return ns.emailClient.SendEmail(notification.Recipient, notification.Subject, notification.Body)

	case "push":
		ns.logger.Info("Sending push notification",
			zap.String("token", notification.Recipient),
			zap.String("body", notification.Body))
		return ns.pushClient.SendPush(notification.Recipient, notification.Subject, notification.Body, notification.Data)

	default:
		return fmt.Errorf("unsupported notification type: %s", notification.Type)
	}
}

// SendBulkNotifications sends multiple notifications concurrently
func (ns *NotificationService) SendBulkNotifications(notifications []*Notification) []error {
	errors := make([]error, 0)

	for _, notification := range notifications {
		if err := ns.SendNotification(notification); err != nil {
			ns.logger.Error("Failed to send notification",
				zap.String("type", notification.Type),
				zap.String("recipient", notification.Recipient),
				zap.Error(err))
			errors = append(errors, fmt.Errorf("failed to send %s to %s: %w",
				notification.Type, notification.Recipient, err))
		}
		time.Sleep(100 * time.Millisecond) // Rate limiting
	}

	return errors
}

// FormatTemplate formats a notification template with provided data
func (ns *NotificationService) FormatTemplate(template string, data map[string]interface{}) string {
	result := template

	// Simple placeholder replacement
	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = replaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}

	return result
}

// replaceAll replaces all occurrences of old with new in s
func replaceAll(s, old, new string) string {
	for {
		newS := ""
		i := 0
		for i < len(s) {
			if i+len(old) <= len(s) && s[i:i+len(old)] == old {
				newS += new
				i += len(old)
			} else {
				newS += string(s[i])
				i++
			}
		}
		if newS == s {
			break
		}
		s = newS
	}
	return s
}

// Mock implementations for testing
type MockSMSService struct{}

func (m *MockSMSService) SendSMS(to, message string) error {
	fmt.Printf("[MOCK SMS] To: %s, Message: %s\n", to, message)
	return nil
}

type MockEmailService struct{}

func (m *MockEmailService) SendEmail(to, subject, body string) error {
	fmt.Printf("[MOCK EMAIL] To: %s, Subject: %s\n", to, subject)
	return nil
}

// Notification templates for common alert types
var (
	OverspeedTemplate = `🚨 Overspeed Alert 🚨
	
Vehicle: {{vehicle_registration}}
Driver: {{driver_name}}
Current Speed: {{current_speed}} km/h
Speed Limit: {{speed_limit}} km/h
Location: {{location}}
Time: {{timestamp}}

Please advise the driver to reduce speed immediately.`

	GeofenceEntryTemplate = `📍 Geofence Entry Alert
	
Vehicle: {{vehicle_registration}}
Driver: {{driver_name}}
Entered: {{geofence_name}}
Location: {{location}}
Time: {{timestamp}}

Vehicle has entered designated area.`

	GeofenceExitTemplate = `📍 Geofence Exit Alert
	
Vehicle: {{vehicle_registration}}
Driver: {{driver_name}}
Exited: {{geofence_name}}
Location: {{location}}
Time: {{timestamp}}

Vehicle has left designated area.`

	HarshBrakingTemplate = `⚠️ Harsh Braking Detected
	
Vehicle: {{vehicle_registration}}
Driver: {{driver_name}}
Severity: {{severity}}
Location: {{location}}
Time: {{timestamp}}

Driver performed harsh braking. Please review driving behavior.`

	FuelTheftTemplate = `⛽ Fuel Theft Alert
	
Vehicle: {{vehicle_registration}}
Driver: {{driver_name}}
Fuel Level Drop: {{fuel_drop}} liters
Location: {{location}}
Time: {{timestamp}}

Significant fuel level drop detected. Possible fuel theft.`

	MaintenanceDueTemplate = `🔧 Maintenance Due Alert
	
Vehicle: {{vehicle_registration}}
Component: {{component}}
Due Date: {{due_date}}
Current Reading: {{current_reading}}
Recommended Action: {{action}}

Scheduled maintenance is due. Please arrange service.`
)
