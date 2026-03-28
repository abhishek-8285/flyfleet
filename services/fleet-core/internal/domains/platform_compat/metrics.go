package platform_compat

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Device compatibility metrics
	DeviceCompatibilityChecksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "platform_device_compatibility_checks_total",
		Help: "Total device compatibility checks performed",
	}, []string{"org_id", "device_type", "os_version"})

	DeviceCompatibilityPassRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "platform_device_compatibility_pass_rate_percentage",
		Help: "Device compatibility pass rate",
	}, []string{"org_id", "device_type"})

	IncompatibleDevicesTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "platform_incompatible_devices_total",
		Help: "Count of incompatible devices",
	}, []string{"org_id", "incompatibility_reason"})

	// OS version distribution
	OSVersionCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "platform_os_version_count",
		Help: "Count of devices by OS version",
	}, []string{"org_id", "os_type", "os_version"})

	// Browser compatibility
	BrowserCompatibilityChecksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "platform_browser_compatibility_checks_total",
		Help: "Browser compatibility check count",
	}, []string{"org_id", "browser_type", "browser_version"})

	BrowserSupportMatrix = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "platform_browser_support_status",
		Help: "Browser support status (1=support, 0=unsupported)",
	}, []string{"browser_type", "browser_version"})

	// Capability detection
	DeviceCapabilitiesDetected = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "platform_device_capabilities_detected_total",
		Help: "Device capabilities detected",
	}, []string{"org_id", "capability"})

	MissingCapabilitiesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "platform_missing_capabilities_total",
		Help: "Required capabilities missing on devices",
	}, []string{"org_id", "capability", "device_type"})

	// App version metrics
	AppVersionDistribution = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "platform_app_version_count",
		Help: "Count of devices by app version",
	}, []string{"org_id", "platform", "app_version"})

	OutdatedAppDevices = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "platform_outdated_app_devices_total",
		Help: "Devices running outdated app versions",
	}, []string{"org_id", "current_version", "latest_version"})

	// Screen resolution metrics
	ScreenResolutionCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "platform_screen_resolution_count",
		Help: "Devices by screen resolution",
	}, []string{"org_id", "resolution", "density"})

	UnsupportedScreenResolutions = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "platform_unsupported_screen_resolutions_total",
		Help: "Devices with unsupported screen resolution",
	}, []string{"org_id", "resolution"})

	// Network capability metrics
	NetworkTypeDistribution = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "platform_network_type_count",
		Help: "Devices by network connectivity type",
	}, []string{"org_id", "network_type", "generation"})

	UnsupportedNetworkTypes = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "platform_unsupported_network_types_total",
		Help: "Devices with unsupported network type",
	}, []string{"org_id", "network_type"})

	// Location service metrics
	LocationServiceEnabled = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "platform_location_service_enabled_percentage",
		Help: "Percentage of devices with location service enabled",
	}, []string{"org_id", "platform"})

	LocationAccuracy = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "platform_location_accuracy_meters",
		Help:    "GPS location accuracy in meters",
		Buckets: prometheus.ExponentialBuckets(5, 2, 10),
	}, []string{"org_id", "device_type"})

	// Push notification metrics
	PushNotificationSupport = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "platform_push_notification_support_percentage",
		Help: "Devices with push notification support",
	}, []string{"org_id", "platform"})

	PushTokenRegistrationSuccessRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "platform_push_token_registration_success_rate_percentage",
		Help: "Push token registration success rate",
	}, []string{"org_id", "platform"})

	// Storage metrics
	StorageSpaceAvailable = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "platform_storage_space_available_bytes",
		Help: "Available storage space on devices",
	}, []string{"org_id", "device_type"})

	InsufficientStorageErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "platform_insufficient_storage_errors_total",
		Help: "Errors due to insufficient storage",
	}, []string{"org_id"})

	// Battery metrics
	BatteryLevel = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "platform_battery_level_percentage",
		Help: "Device battery level",
	}, []string{"org_id", "device_type"})

	BatterySaverModeEnabled = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "platform_battery_saver_mode_enabled_percentage",
		Help: "Devices with battery saver mode enabled",
	}, []string{"org_id"})

	// Memory metrics
	MemoryUsage = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "platform_memory_usage_bytes",
		Help: "App memory usage",
	}, []string{"org_id", "device_type"})

	MemoryPressureEvents = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "platform_memory_pressure_events_total",
		Help: "Memory pressure events",
	}, []string{"org_id", "severity"})

	// API level compatibility
	APIVersionSupport = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "platform_api_version_support",
		Help: "API version support matrix (1=support)",
	}, []string{"api_version", "platform", "min_app_version"})

	DeprecatedAPICalls = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "platform_deprecated_api_calls_total",
		Help: "Calls to deprecated API versions",
	}, []string{"org_id", "api_version", "deprecated_since"})

	// Feature flags compatibility
	FeatureFlagSupport = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "platform_feature_flag_support_percentage",
		Help: "Feature availability per platform",
	}, []string{"feature", "platform", "min_version"})

	// SSL/TLS compatibility
	TLSVersionSupported = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "platform_tls_version_supported",
		Help: "TLS version support (1=support)",
	}, []string{"platform", "tls_version"})

	ClientCertificateSupport = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "platform_client_cert_support_percentage",
		Help: "Devices with client certificate support",
	})

	// Accessibility metrics
	AccessibilityFeaturesEnabled = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "platform_accessibility_features_enabled_percentage",
		Help: "Accessibility features enabled by users",
	}, []string{"org_id", "feature"})

	// Localization metrics
	LocalizationSupported = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "platform_localization_supported",
		Help: "Localization support by language/region",
	}, []string{"language", "region"})

	UnlocalizedContent = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "platform_unlocalized_content_total",
		Help: "UI elements without translations",
	}, []string{"language", "component"})
)

// Global metrics instance
var GlobalMetrics = struct {
	DeviceCompatibilityChecksTotal   *prometheus.CounterVec
	DeviceCompatibilityPassRate      *prometheus.GaugeVec
	IncompatibleDevicesTotal         *prometheus.GaugeVec
	OSVersionCount                   *prometheus.GaugeVec
	BrowserCompatibilityChecksTotal  *prometheus.CounterVec
	BrowserSupportMatrix             *prometheus.GaugeVec
	DeviceCapabilitiesDetected       *prometheus.CounterVec
	MissingCapabilitiesTotal         *prometheus.CounterVec
	AppVersionDistribution           *prometheus.GaugeVec
	OutdatedAppDevices               *prometheus.GaugeVec
	ScreenResolutionCount            *prometheus.GaugeVec
	UnsupportedScreenResolutions     *prometheus.GaugeVec
	NetworkTypeDistribution          *prometheus.GaugeVec
	UnsupportedNetworkTypes          *prometheus.GaugeVec
	LocationServiceEnabled           *prometheus.GaugeVec
	LocationAccuracy                 *prometheus.HistogramVec
	PushNotificationSupport          *prometheus.GaugeVec
	PushTokenRegistrationSuccessRate *prometheus.GaugeVec
	StorageSpaceAvailable            *prometheus.GaugeVec
	InsufficientStorageErrors        *prometheus.CounterVec
	BatteryLevel                     *prometheus.GaugeVec
	BatterySaverModeEnabled          *prometheus.GaugeVec
	MemoryUsage                      *prometheus.GaugeVec
	MemoryPressureEvents             *prometheus.CounterVec
	APIVersionSupport                *prometheus.GaugeVec
	DeprecatedAPICalls               *prometheus.CounterVec
	FeatureFlagSupport               *prometheus.GaugeVec
	TLSVersionSupported              *prometheus.GaugeVec
	ClientCertificateSupport         prometheus.Gauge
	AccessibilityFeaturesEnabled     *prometheus.GaugeVec
	LocalizationSupported            *prometheus.GaugeVec
	UnlocalizedContent               *prometheus.CounterVec
}{
	DeviceCompatibilityChecksTotal:   DeviceCompatibilityChecksTotal,
	DeviceCompatibilityPassRate:      DeviceCompatibilityPassRate,
	IncompatibleDevicesTotal:         IncompatibleDevicesTotal,
	OSVersionCount:                   OSVersionCount,
	BrowserCompatibilityChecksTotal:  BrowserCompatibilityChecksTotal,
	BrowserSupportMatrix:             BrowserSupportMatrix,
	DeviceCapabilitiesDetected:       DeviceCapabilitiesDetected,
	MissingCapabilitiesTotal:         MissingCapabilitiesTotal,
	AppVersionDistribution:           AppVersionDistribution,
	OutdatedAppDevices:               OutdatedAppDevices,
	ScreenResolutionCount:            ScreenResolutionCount,
	UnsupportedScreenResolutions:     UnsupportedScreenResolutions,
	NetworkTypeDistribution:          NetworkTypeDistribution,
	UnsupportedNetworkTypes:          UnsupportedNetworkTypes,
	LocationServiceEnabled:           LocationServiceEnabled,
	LocationAccuracy:                 LocationAccuracy,
	PushNotificationSupport:          PushNotificationSupport,
	PushTokenRegistrationSuccessRate: PushTokenRegistrationSuccessRate,
	StorageSpaceAvailable:            StorageSpaceAvailable,
	InsufficientStorageErrors:        InsufficientStorageErrors,
	BatteryLevel:                     BatteryLevel,
	BatterySaverModeEnabled:          BatterySaverModeEnabled,
	MemoryUsage:                      MemoryUsage,
	MemoryPressureEvents:             MemoryPressureEvents,
	APIVersionSupport:                APIVersionSupport,
	DeprecatedAPICalls:               DeprecatedAPICalls,
	FeatureFlagSupport:               FeatureFlagSupport,
	TLSVersionSupported:              TLSVersionSupported,
	ClientCertificateSupport:         ClientCertificateSupport,
	AccessibilityFeaturesEnabled:     AccessibilityFeaturesEnabled,
	LocalizationSupported:            LocalizationSupported,
	UnlocalizedContent:               UnlocalizedContent,
}
