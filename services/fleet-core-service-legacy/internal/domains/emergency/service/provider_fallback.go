package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core-service/internal/domains/emergency"
)

// ProviderFallbackManager handles the SMS → Voice Call → Email fallback chain.
type ProviderFallbackManager struct {
	SMSProviders  []NotificationProvider
	CallProviders []NotificationProvider
	health        *ProviderHealthService
	logger        *zap.Logger
}

func NewProviderFallbackManager(health *ProviderHealthService, logger *zap.Logger) *ProviderFallbackManager {
	msg91 := NewMSG91Provider(logger)
	twilio := NewTwilioProvider(logger)
	exotel := NewExotelProvider(logger)
	return &ProviderFallbackManager{
		SMSProviders:  []NotificationProvider{msg91, twilio, exotel},
		CallProviders: []NotificationProvider{exotel, twilio},
		health:        health,
		logger:        logger,
	}
}

// SendSMSWithFallback tries each SMS provider in order, then falls back to voice, then email.
func (m *ProviderFallbackManager) SendSMSWithFallback(ctx context.Context, phone, email, message string) error {
	// Try SMS providers
	for _, p := range m.SMSProviders {
		if !m.health.IsHealthy(ctx, p.GetName()) {
			m.logger.Warn("Skipping unhealthy provider", zap.String("provider", p.GetName()))
			continue
		}

		if err := p.SendSMS(ctx, phone, message); err == nil {
			m.logger.Info("SMS sent successfully", zap.String("provider", p.GetName()), zap.String("phone", phone))
			m.health.RecordSuccess(ctx, p.GetName())
			emergency.NotificationRetriesTotal.WithLabelValues(p.GetName(), "sms").Inc()
			return nil
		}
		
		m.health.RecordFailure(ctx, p.GetName())
		m.logger.Warn("SMS provider failed, trying next", zap.String("provider", p.GetName()))
		emergency.NotificationRetriesTotal.WithLabelValues(p.GetName(), "sms_fail").Inc()
	}

	// Fallback: Voice call
	for _, p := range m.CallProviders {
		if !m.health.IsHealthy(ctx, p.GetName()) {
			continue
		}

		if err := p.SendVoiceCall(ctx, phone, message); err == nil {
			m.logger.Info("Voice call fallback succeeded", zap.String("provider", p.GetName()))
			m.health.RecordSuccess(ctx, p.GetName())
			emergency.NotificationRetriesTotal.WithLabelValues(p.GetName(), "voice_fallback").Inc()
			return nil
		}
		m.health.RecordFailure(ctx, p.GetName())
	}

	// Fallback: Email
	if len(m.SMSProviders) > 0 {
		p := m.SMSProviders[0] // Default to first provider for email
		if err := p.SendEmail(ctx, email, "Emergency SOS Alert", message); err == nil {
			m.logger.Info("Email fallback succeeded", zap.String("provider", p.GetName()))
			m.health.RecordSuccess(ctx, p.GetName())
			emergency.NotificationRetriesTotal.WithLabelValues(p.GetName(), "email_fallback").Inc()
			return nil
		}
		m.health.RecordFailure(ctx, p.GetName())
	}

	m.logger.Error("All notification channels exhausted", zap.String("phone", phone))
	return ErrAllProvidersFailed
}
func (m *ProviderFallbackManager) CheckProvidersHealth(ctx context.Context) map[string]string {
	results := make(map[string]string)
	allProviders := append([]NotificationProvider{}, m.SMSProviders...)
	allProviders = append(allProviders, m.CallProviders...)

	for _, p := range allProviders {
		if _, exists := results[p.GetName()]; exists {
			continue
		}
		if err := p.Ping(ctx); err != nil {
			results[p.GetName()] = "unhealthy: " + err.Error()
		} else {
			results[p.GetName()] = "healthy"
		}
	}
	return results
}
