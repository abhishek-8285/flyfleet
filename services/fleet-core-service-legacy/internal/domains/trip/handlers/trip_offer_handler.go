package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// OfferTrip godoc
// @Summary Offer a trip to a driver
// @Description Creates a trip offer for a specific driver
// @Tags trip-offers
// @Accept json
// @Produce json
// @Param id path string true "Trip ID"
// @Param body body struct{DriverID string} true "Driver ID"
// @Success 201 {object} models.TripOffer
// @Router /trips/{id}/offer [post]
func (h *TripHandler) OfferTrip(w http.ResponseWriter, r *http.Request) {
	tripID := chi.URLParam(r, "id")
	var req struct {
		DriverID string `json:"driver_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	offer, err := h.service.OfferTrip(r.Context(), tripID, req.DriverID)
	if err != nil {
		h.logger.Error("failed to offer trip", zap.Error(err))
		http.Error(w, "Failed to offer trip", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(offer)
}

// RespondToOffer godoc
// @Summary Respond to trip offer
// @Description Accept or reject a trip offer
// @Tags trip-offers
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param body body struct{DriverID string; Response string; Reason string} true "Response details"
// @Success 200 {object} map[string]string
// @Router /trips/offers/{id}/respond [post]
func (h *TripHandler) RespondToOffer(w http.ResponseWriter, r *http.Request) {
	offerID := chi.URLParam(r, "id")
	var req struct {
		DriverID string `json:"driver_id"`
		Response string `json:"response"` // ACCEPTED, REJECTED
		Reason   string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.RespondToOffer(r.Context(), offerID, req.DriverID, req.Response, req.Reason); err != nil {
		h.logger.Error("failed to respond to offer", zap.Error(err))
		http.Error(w, "Failed to respond to offer", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "Response recorded"})
}

// GetDriverOffers godoc
// @Summary Get driver offers
// @Description Retrieves active trip offers for a driver
// @Tags trip-offers
// @Accept json
// @Produce json
// @Param driver_id query string true "Driver ID"
// @Success 200 {array} models.TripOffer
// @Router /trips/offers [get]
func (h *TripHandler) GetDriverOffers(w http.ResponseWriter, r *http.Request) {
	driverID := r.URL.Query().Get("driver_id")
	if driverID == "" {
		http.Error(w, "driver_id is required", http.StatusBadRequest)
		return
	}

	offers, err := h.service.GetDriverOffers(r.Context(), driverID)
	if err != nil {
		h.logger.Error("failed to get driver offers", zap.Error(err))
		http.Error(w, "Failed to get driver offers", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(offers)
}
