package metrics

import "time"

// RecordTripCompleted increments the trips completed counter
func RecordTripCompleted() {
	TripsCompletedTotal.Inc()
}

// RecordInvoiceGenerated increments the invoices generated counter
func RecordInvoiceGenerated() {
	InvoicesGeneratedTotal.Inc()
}

// SetVehiclesOnline sets the current number of vehicles online
func SetVehiclesOnline(count float64) {
	VehiclesOnline.Set(count)
}

// SetActiveTrips sets the current number of active trips
func SetActiveTrips(count float64) {
	ActiveTrips.Set(count)
}

// RecordCommandSent increments the command sent counter with result
func RecordCommandSent(commandType, result string) {
	DeviceCommandEventsTotal.WithLabelValues(commandType, result).Inc()
}

// RecordCommandLatency observes command acknowledgement latency
func RecordCommandLatency(commandType string, duration time.Duration) {
	DeviceCommandAckLatency.WithLabelValues(commandType).Observe(duration.Seconds())
}

// SetCommandQueueDepth sets the queue depth for a command type
func SetCommandQueueDepth(commandType string, depth float64) {
	CommandQueueDepth.WithLabelValues(commandType).Set(depth)
}

// RecordAlertTriggered increments alert counter
func RecordAlertTriggered(severity, alertType string) {
	AlertsTriggeredTotal.WithLabelValues(severity, alertType).Inc()
}

// SetActiveUserSessions sets the active user session count
func SetActiveUserSessions(count float64) {
	ActiveUserSessions.Set(count)
}
