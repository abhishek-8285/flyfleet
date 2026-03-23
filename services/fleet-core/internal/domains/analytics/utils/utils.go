package utils

// Ptr returns a pointer to the given value
func Ptr[T any](v T) *T {
	return &v
}

// MapStatusToUrgency maps a prediction status to a maintenance urgency level
func MapStatusToUrgency(status string) string {
	switch status {
	case "critical":
		return "critical"
	case "warning":
		return "high"
	case "monitor":
		return "medium"
	default:
		return "low"
	}
}
