package analytics

import (
	"log"

	"github.com/posthog/posthog-go"
)

var client posthog.Client

// Init initializes the global PostHog client for the backend
func Init(apiKey, endpoint string) {
	if apiKey == "" {
		log.Println("PostHog API key not provided. Analytics disabled.")
		return
	}

	var err error
	client, err = posthog.NewWithConfig(
		apiKey,
		posthog.Config{
			Endpoint: endpoint,
		},
	)
	if err != nil {
		log.Printf("Failed to initialize PostHog client: %v\n", err)
	} else {
		log.Println("PostHog client initialized successfully.")
	}
}

// TrackEvent sends a custom event to PostHog
func TrackEvent(eventName, distinctId string, properties map[string]interface{}) {
	if client == nil {
		return
	}

	err := client.Enqueue(posthog.Capture{
		DistinctId: distinctId,
		Event:      eventName,
		Properties: properties,
	})

	if err != nil {
		log.Printf("PostHog Enqueue Capture failed: %v\n", err)
	}
}

// Identify user explicitly
func Identify(distinctId string, properties map[string]interface{}) {
	if client == nil {
		return
	}

	err := client.Enqueue(posthog.Identify{
		DistinctId: distinctId,
		Properties: properties,
	})

	if err != nil {
		log.Printf("PostHog Enqueue Identify failed: %v\n", err)
	}
}

// Close gracefully flushes and closes the client
func Close() {
	if client != nil {
		client.Close()
	}
}
