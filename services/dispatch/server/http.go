package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/flyfleet/services/dispatch"
)

type HTTPServer struct {
	dispatcher *dispatch.Dispatcher
	server     *http.Server
}

func NewHTTPServer(dispatcher *dispatch.Dispatcher, addr string) *HTTPServer {
	return &HTTPServer{
		dispatcher: dispatcher,
		server: &http.Server{
			Addr:         addr,
			Handler:      http.DefaultServeMux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
	}
}

func (s *HTTPServer) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/dispatch", s.handleDispatch)
	mux.HandleFunc("/offer", s.handleOfferRide)
	mux.HandleFunc("/nearby", s.handleFindNearbyDrivers)
	mux.HandleFunc("/result", s.handleGetDispatchResult)
	mux.HandleFunc("/health", s.handleHealth)

	s.server.Handler = mux

	log.Info().Str("addr", s.server.Addr).Msg("dispatch HTTP server starting")
	return s.server.ListenAndServe()
}

func (s *HTTPServer) Stop() error {
	return s.server.Shutdown(nil)
}

func (s *HTTPServer) handleDispatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dispatch.RideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	idempotencyKey := r.Header.Get("Idempotency-Key")
	if idempotencyKey != "" {
		r = r.WithContext(dispatch.WithIdempotencyKey(r.Context(), idempotencyKey))
	}

	result, err := s.dispatcher.Dispatch(r.Context(), &req)
	if err != nil {
		log.Error().Err(err).Str("ride_id", req.RideID).Msg("dispatch failed")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *HTTPServer) handleOfferRide(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RideID     string  `json:"ride_id"`
		DriverID   string  `json:"driver_id"`
		RiderID    string  `json:"rider_id"`
		PickupLat  float64 `json:"pickup_lat"`
		PickupLng  float64 `json:"pickup_lng"`
		TimeoutSec int     `json:"timeout_sec"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	rideReq := &dispatch.RideRequest{
		RideID: req.RideID,
		RiderID: req.RiderID,
		Lat:     req.PickupLat,
		Lng:     req.PickupLng,
	}

	accepted, err := s.dispatcher.OfferRide(r.Context(), rideReq, req.DriverID, req.TimeoutSec)
	if err != nil {
		log.Error().Err(err).Str("ride_id", req.RideID).Msg("offer ride failed")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"accepted": accepted,
		"ride_id":  req.RideID,
		"driver_id": req.DriverID,
	})
}

func (s *HTTPServer) handleFindNearbyDrivers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	lat, _ := strconv.ParseFloat(r.URL.Query().Get("pickup_lat"), 64)
	lng, _ := strconv.ParseFloat(r.URL.Query().Get("pickup_lng"), 64)
	radius, _ := strconv.ParseFloat(r.URL.Query().Get("search_radius_km"), 64)

	req := &dispatch.RideRequest{
		RideID:         r.URL.Query().Get("ride_id"),
		RiderID:        r.URL.Query().Get("rider_id"),
		ZoneID:         r.URL.Query().Get("zone_id"),
		Lat:            lat,
		Lng:            lng,
		SearchRadiusKm: radius,
	}

	drivers, err := s.dispatcher.FindNearbyDrivers(r.Context(), req)
	if err != nil {
		log.Error().Err(err).Msg("find nearby drivers failed")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"driver_ids": drivers,
		"total_found": len(drivers),
	})
}

func (s *HTTPServer) handleGetDispatchResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rideID := r.URL.Query().Get("ride_id")
	if rideID == "" {
		http.Error(w, "ride_id is required", http.StatusBadRequest)
		return
	}

	result, err := s.dispatcher.GetDispatchResult(r.Context(), rideID)
	if err != nil {
		log.Error().Err(err).Str("ride_id", rideID).Msg("get dispatch result failed")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}