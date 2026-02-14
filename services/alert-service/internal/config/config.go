package config

import (
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	Port              string
	DatabaseURL       string
	NatsURL           string
	TwilioAccountSID  string
	TwilioAuthToken   string
	TwilioPhoneNumber string
	SendGridAPIKey    string
	JWTSecret         string
	AWSRegion         string
}

func Load() *Config {
	v := viper.New()

	v.SetDefault("server.port", "8085")
	v.SetDefault("database.url", "")
	v.SetDefault("nats.url", nats.DefaultURL)
	v.SetDefault("twilio.account_sid", "")
	v.SetDefault("twilio.auth_token", "")
	v.SetDefault("twilio.phone_number", "")
	v.SetDefault("sendgrid.api_key", "")
	v.SetDefault("security.jwt_secret", "your-jwt-secret-key-here")
	v.SetDefault("aws.region", "ap-south-1")

	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath("../../") 
	v.AddConfigPath(".")      

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			zap.S().Warnf("Error reading config file: %v", err)
		}
	}

	v.BindEnv("server.port", "PORT")
	v.BindEnv("database.url", "DATABASE_URL")
	v.BindEnv("nats.url", "NATS_URL")
	v.BindEnv("twilio.account_sid", "TWILIO_ACCOUNT_SID")
	v.BindEnv("twilio.auth_token", "TWILIO_AUTH_TOKEN")
	v.BindEnv("twilio.phone_number", "TWILIO_PHONE_NUMBER")
	v.BindEnv("sendgrid.api_key", "SENDGRID_API_KEY")
	v.BindEnv("security.jwt_secret", "JWT_SECRET")
	v.BindEnv("aws.region", "AWS_REGION")

	return &Config{
		Port:              v.GetString("server.port"),
		DatabaseURL:       v.GetString("database.url"),
		NatsURL:           v.GetString("nats.url"),
		TwilioAccountSID:  v.GetString("twilio.account_sid"),
		TwilioAuthToken:   v.GetString("twilio.auth_token"),
		TwilioPhoneNumber: v.GetString("twilio.phone_number"),
		SendGridAPIKey:    v.GetString("sendgrid.api_key"),
		JWTSecret:         v.GetString("security.jwt_secret"),
		AWSRegion:         v.GetString("aws.region"),
	}
}
