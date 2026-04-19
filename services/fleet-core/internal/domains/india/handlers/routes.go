package handlers

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts all India integration endpoints under /api/v1/india
func (h *IndiaIntegrationHandlers) RegisterRoutes(r chi.Router) {
	// FASTag routes
	r.Route("/fastag", func(r chi.Router) {
		r.Post("/transactions", h.CreateFASTagTransaction)
		r.Get("/transactions", h.ListFASTagTransactions)
		r.Get("/transactions/{id}", h.GetFASTagTransaction)
		r.Put("/transactions/{id}", h.UpdateFASTagTransaction)
		r.Delete("/transactions/{id}", h.DeleteFASTagTransaction)
		r.Post("/sync", h.SyncFASTag)
		r.Post("/webhook", h.HandleFASTagWebhook)
	})

	// E-Way Bill routes
	r.Route("/ewaybill", func(r chi.Router) {
		r.Post("/", h.CreateEWayBill)
		r.Get("/", h.ListEWayBills)
		r.Get("/{id}", h.GetEWayBill)
		r.Get("/number/{number}", h.GetEWayBillByNumber)
		r.Put("/{id}", h.UpdateEWayBill)
		r.Delete("/{id}", h.DeleteEWayBill)
		r.Post("/validate", h.ValidateEWayBill)
		r.Post("/webhook", h.HandleEWayBillWebhook)
	})

	// Vehicle Verification (VAHAN) routes
	r.Route("/vahan", func(r chi.Router) {
		r.Post("/register-internal", h.CreateVehicleVerification) // Renamed for clarity
		r.Post("/verify", h.VerifyVehicle)
		r.Get("/verifications", h.ListVehicleVerifications)
		r.Get("/verifications/{id}", h.GetVehicleVerification)
		r.Get("/verifications/reg/{reg_number}", h.GetVehicleVerificationByRegNumber)
		r.Put("/verifications/{id}", h.UpdateVehicleVerification)
		r.Delete("/verifications/{id}", h.DeleteVehicleVerification)
		r.Post("/verify/bulk", h.BulkVerifyVahan)
	})

	// Driver Verification (SARATHI) routes
	r.Route("/sarathi", func(r chi.Router) {
		r.Post("/verify", h.CreateDriverVerification)
		r.Get("/verifications", h.ListDriverVerifications)
		r.Get("/verifications/{id}", h.GetDriverVerification)
		r.Get("/verifications/license/{license_number}", h.GetDriverVerificationByLicenseNumber)
		r.Put("/verifications/{id}", h.UpdateDriverVerification)
		r.Delete("/verifications/{id}", h.DeleteDriverVerification)
		r.Post("/verify/bulk", h.BulkVerifySarathi)
	})

	// Local Permit routes
	r.Route("/permits", func(r chi.Router) {
		r.Post("/", h.CreateLocalPermit)
		r.Get("/", h.ListLocalPermits)
		r.Get("/{id}", h.GetLocalPermit)
		r.Get("/number/{permit_number}", h.GetLocalPermitByNumber)
		r.Put("/{id}", h.UpdateLocalPermit)
		r.Delete("/{id}", h.DeleteLocalPermit)
	})

	// GST Compliance routes
	r.Route("/gst", func(r chi.Router) {
		r.Post("/compliance", h.CreateGSTCompliance)
		r.Get("/compliance", h.ListGSTCompliances)
		r.Get("/compliance/{id}", h.GetGSTCompliance)
		r.Get("/compliance/gstin/{gstin}", h.GetGSTComplianceByGSTIN)
		r.Put("/compliance/{id}", h.UpdateGSTCompliance)
		r.Delete("/compliance/{id}", h.DeleteGSTCompliance)
		r.Post("/webhook", h.HandleGSTWebhook)
	})

	// Integration Config routes
	r.Route("/config", func(r chi.Router) {
		r.Post("/", h.CreateIntegrationConfig)
		r.Get("/", h.ListIntegrationConfigs)
		r.Get("/{id}", h.GetIntegrationConfig)
		r.Get("/org/{tenant_id}/service/{service_type}", h.GetIntegrationConfigByServiceType)
		r.Put("/{id}", h.UpdateIntegrationConfig)
		r.Delete("/{id}", h.DeleteIntegrationConfig)
	})

	// Health check
	r.Get("/health", h.HealthCheck)
}
