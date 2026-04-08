package preference_engine

import (
	"github.com/flyfleet/fleet-core/internal/domains/alerts/models"
	"github.com/google/uuid"
)

type PreferenceEngine interface {
	FilterChannels(userID *uuid.UUID, requestedChannels []models.Channel, userPrefs []models.UserPreference) []models.Channel
}

type preferenceEngine struct{}

func NewPreferenceEngine() PreferenceEngine {
	return &preferenceEngine{}
}

func (e *preferenceEngine) FilterChannels(userID *uuid.UUID, requestedChannels []models.Channel, userPrefs []models.UserPreference) []models.Channel {
	if userID == nil {
		return requestedChannels // No user, always send (e.g. system alert to external recipient)
	}

	// Create a map for quick lookup
	prefMap := make(map[models.Channel]bool)
	for _, p := range userPrefs {
		prefMap[p.Channel] = p.Enabled
	}

	filtered := make([]models.Channel, 0, len(requestedChannels))
	for _, ch := range requestedChannels {
		enabled, exists := prefMap[ch]
		if !exists || enabled {
			// If no preference set, default to enabled. If set, follow preference.
			filtered = append(filtered, ch)
		}
	}

	return filtered
}
