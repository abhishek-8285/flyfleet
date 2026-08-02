package provider

// SelectedProvider pairs a constructed Provider with the config ID
// it was built from, enabling per-provider delivery tracking in the
// delivery_attempts table.
type SelectedProvider struct {
	Provider  Provider
	ConfigID  string
}
