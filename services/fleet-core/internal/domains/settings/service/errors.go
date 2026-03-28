package service

import "errors"

// Settings service errors
var (
	// General errors
	ErrSettingsNotFound = errors.New("settings not found")
	ErrInvalidSettings  = errors.New("invalid settings")
	ErrInvalidRequest   = errors.New("invalid request")

	// Platform settings errors
	ErrPlatformSettingsNotFound = errors.New("platform settings not found")
	ErrInvalidPlatformSettings  = errors.New("invalid platform settings")
	ErrInvalidEmailFormat       = errors.New("invalid email format")
	ErrInvalidTimezone          = errors.New("invalid timezone")
	ErrInvalidCurrency          = errors.New("invalid currency code")
	ErrInvalidLanguage          = errors.New("invalid language code")

	// Feature flag errors
	ErrFeatureFlagNotFound   = errors.New("feature flag not found")
	ErrInvalidFeatureFlag    = errors.New("invalid feature flag")
	ErrInvalidFeatureFlagKey = errors.New("invalid feature flag key")
	ErrDuplicateFeatureFlag  = errors.New("duplicate feature flag key")

	// Template errors
	ErrTemplateNotFound   = errors.New("template not found")
	ErrInvalidTemplate    = errors.New("invalid template")
	ErrInvalidTemplateKey = errors.New("invalid template key")
	ErrDuplicateTemplate  = errors.New("duplicate template key")

	// Localization errors
	ErrLocalizationNotFound  = errors.New("localization settings not found")
	ErrInvalidLanguageCode   = errors.New("invalid language code")
	ErrDuplicateLanguageCode = errors.New("duplicate language code")

	// Notification channel errors
	ErrChannelNotFound       = errors.New("notification channel not found")
	ErrInvalidChannel        = errors.New("invalid notification channel")
	ErrInvalidChannelType    = errors.New("invalid channel type")
	ErrDuplicateChannel      = errors.New("duplicate channel type")
	ErrInvalidProviderConfig = errors.New("invalid provider configuration")

	// System limits errors
	ErrSystemLimitsNotFound = errors.New("system limits not found")
	ErrInvalidSystemLimits  = errors.New("invalid system limits")
	ErrInvalidRateLimit     = errors.New("invalid rate limit")
	ErrInvalidRetention     = errors.New("invalid retention period")

	// Audit errors
	ErrAuditLogNotFound  = errors.New("audit log not found")
	ErrInvalidAuditQuery = errors.New("invalid audit query")

	// Import/Export errors
	ErrImportFailed           = errors.New("import failed")
	ErrExportFailed           = errors.New("export failed")
	ErrInvalidImportPayload   = errors.New("invalid import payload")
	ErrMissingRequiredSection = errors.New("at least one settings section is required")

	// Validation errors
	ErrRequiredField = errors.New("required field is missing")
	ErrInvalidFormat = errors.New("invalid format")
	ErrInvalidValue  = errors.New("invalid value")
	ErrNegativeValue = errors.New("value cannot be negative")
)

// SettingsError represents a settings domain error
type SettingsError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func (e *SettingsError) Error() string {
	return e.Message
}

// NewSettingsError creates a new settings error
func NewSettingsError(code, message string) *SettingsError {
	return &SettingsError{
		Code:    code,
		Message: message,
	}
}

// NewValidationError creates a validation error
func NewValidationError(field, message string) *SettingsError {
	return &SettingsError{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Field:   field,
	}
}

// NewPlatformError creates a platform settings error
func NewPlatformError(message string) *SettingsError {
	return &SettingsError{
		Code:    "PLATFORM_ERROR",
		Message: message,
	}
}

// NewFeatureFlagError creates a feature flag error
func NewFeatureFlagError(message string) *SettingsError {
	return &SettingsError{
		Code:    "FEATURE_FLAG_ERROR",
		Message: message,
	}
}

// NewTemplateError creates a template error
func NewTemplateError(message string) *SettingsError {
	return &SettingsError{
		Code:    "TEMPLATE_ERROR",
		Message: message,
	}
}

// NewChannelError creates a notification channel error
func NewChannelError(message string) *SettingsError {
	return &SettingsError{
		Code:    "CHANNEL_ERROR",
		Message: message,
	}
}

// NewImportError creates an import error
func NewImportError(message string) *SettingsError {
	return &SettingsError{
		Code:    "IMPORT_ERROR",
		Message: message,
	}
}
