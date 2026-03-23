package service

import (
	"context"
)

type NotificationProvider interface {
	SendSMS(ctx context.Context, phone, message string) error
	SendVoiceCall(ctx context.Context, phone, message string) error
	SendEmail(ctx context.Context, email, subject, message string) error
	GetName() string
	Ping(ctx context.Context) error
}
