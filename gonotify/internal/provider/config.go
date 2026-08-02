package provider

// ProviderConfig represents a tenant-specific provider configuration
// loaded from the provider_configs table. It carries the fields needed
// to construct a provider.Provider instance via the factory.
type ProviderConfig struct {
	ID                   string
	Type                 string
	Channel              string
	Name                 string
	Settings             map[string]any
	CredentialsEncrypted []byte
	Priority             int
	Weight               int
}
