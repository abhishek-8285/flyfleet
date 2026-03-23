package service

import (
	"context"

	"go.uber.org/zap"
)

type BaseProvider struct {
	logger *zap.Logger
	name   string
}

func (p *BaseProvider) GetName() string {
	return p.name
}

func (p *BaseProvider) Ping(ctx context.Context) error {
	// In production, this would perform a real connectivity check to the provider's API.
	return nil
}

// MSG91 Provider
type MSG91Provider struct {
	BaseProvider
}

func NewMSG91Provider(logger *zap.Logger) *MSG91Provider {
	return &MSG91Provider{BaseProvider{logger: logger, name: "MSG91"}}
}

func (p *MSG91Provider) SendSMS(ctx context.Context, phone, message string) error {
	p.logger.Info("MSG91: Sending SMS", zap.String("phone", phone), zap.String("message", message))
	return nil
}

func (p *MSG91Provider) SendVoiceCall(ctx context.Context, phone, message string) error {
	p.logger.Info("MSG91: Initiating Voice Call", zap.String("phone", phone), zap.String("message", message))
	return nil
}

func (p *MSG91Provider) SendEmail(ctx context.Context, email, subject, message string) error {
	p.logger.Info("MSG91: Sending Email", zap.String("email", email), zap.String("subject", subject))
	return nil
}

// Twilio Provider
type TwilioProvider struct {
	BaseProvider
}

func NewTwilioProvider(logger *zap.Logger) *TwilioProvider {
	return &TwilioProvider{BaseProvider{logger: logger, name: "Twilio"}}
}

func (p *TwilioProvider) SendSMS(ctx context.Context, phone, message string) error {
	p.logger.Info("Twilio: Sending SMS", zap.String("phone", phone), zap.String("message", message))
	return nil
}

func (p *TwilioProvider) SendVoiceCall(ctx context.Context, phone, message string) error {
	p.logger.Info("Twilio: Initiating Voice Call", zap.String("phone", phone), zap.String("message", message))
	return nil
}

func (p *TwilioProvider) SendEmail(ctx context.Context, email, subject, message string) error {
	p.logger.Info("Twilio: Sending Email", zap.String("email", email), zap.String("subject", subject))
	return nil
}

// Exotel Provider
type ExotelProvider struct {
	BaseProvider
}

func NewExotelProvider(logger *zap.Logger) *ExotelProvider {
	return &ExotelProvider{BaseProvider{logger: logger, name: "Exotel"}}
}

func (p *ExotelProvider) SendSMS(ctx context.Context, phone, message string) error {
	p.logger.Info("Exotel: Sending SMS", zap.String("phone", phone), zap.String("message", message))
	return nil
}

func (p *ExotelProvider) SendVoiceCall(ctx context.Context, phone, message string) error {
	p.logger.Info("Exotel: Initiating Voice Call", zap.String("phone", phone), zap.String("message", message))
	return nil
}

func (p *ExotelProvider) SendEmail(ctx context.Context, email, subject, message string) error {
	p.logger.Info("Exotel: Sending Email", zap.String("email", email), zap.String("subject", subject))
	return nil
}
