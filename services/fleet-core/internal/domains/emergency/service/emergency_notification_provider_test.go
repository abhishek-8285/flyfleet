package service

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

func TestProviderFallbackManager_Initializes(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mgr := NewProviderFallbackManager(nil, logger)
	if mgr == nil {
		t.Fatal("Expected non-nil ProviderFallbackManager")
	}
	if len(mgr.SMSProviders) == 0 {
		t.Error("Expected at least one SMS provider registered")
	}
	if len(mgr.CallProviders) == 0 {
		t.Error("Expected at least one voice call provider registered")
	}
}

func TestMSG91Provider_GetName(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	p := NewMSG91Provider(logger)
	if p.GetName() != "MSG91" {
		t.Errorf("Expected MSG91, got %s", p.GetName())
	}
}

func TestTwilioProvider_GetName(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	p := NewTwilioProvider(logger)
	if p.GetName() != "Twilio" {
		t.Errorf("Expected Twilio, got %s", p.GetName())
	}
}

func TestExotelProvider_GetName(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	p := NewExotelProvider(logger)
	if p.GetName() != "Exotel" {
		t.Errorf("Expected Exotel, got %s", p.GetName())
	}
}

func TestMSG91Provider_SendSMS_OK(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	p := NewMSG91Provider(logger)
	err := p.SendSMS(context.Background(), "+919999999999", "Emergency SOS Alert")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
