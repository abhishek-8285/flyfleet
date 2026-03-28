package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flyfleet/fleet-core-service/internal/pkg/ctxkeys"
	pb "github.com/flyfleet/proto/compliance/v1"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockComplianceService struct {
	mock.Mock
}

func (m *MockComplianceService) RecordDutyStatus(ctx context.Context, req *pb.RecordDutyStatusRequest) (*pb.RecordDutyStatusResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.RecordDutyStatusResponse), args.Error(1)
}

func (m *MockComplianceService) GetCurrentDutyStatus(ctx context.Context, driverID string) (*pb.DutyStatusLog, error) {
	args := m.Called(ctx, driverID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.DutyStatusLog), args.Error(1)
}

func newTestHandlers() (*Handlers, *MockComplianceService) {
	logger, _ := zap.NewDevelopment()
	svc := new(MockComplianceService)
	return NewHandlers(svc, logger), svc
}

func newTestRouter(handler *Handlers) http.Handler {
	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	return r
}

func TestHandlers_RecordDutyStatus(t *testing.T) {
	handler, mockSvc := newTestHandlers()
	router := newTestRouter(handler)

	userID := uuid.New().String()
	otherDriverID := uuid.New().String()

	makeReq := func(driverID, body string, userID, role string) *http.Request {
		req := httptest.NewRequest("POST", "/compliance/duty-status/record", strings.NewReader(body))
		ctx := req.Context()
		if userID != "" {
			ctx = context.WithValue(ctx, ctxkeys.UserIDKey, userID)
		}
		if role != "" {
			ctx = context.WithValue(ctx, ctxkeys.RoleKey, role)
		}
		return req.WithContext(ctx)
	}

	validReqBody := func(driver, vehicle string) string {
		// Protobuf JSON uses snake_case field names: driver_id, vehicle_id, status as number
		return `{"driver_id":"` + driver + `","vehicle_id":"` + vehicle + `","status":1}`
	}

	vehicleID := uuid.New().String()

	t.Run("Unauthorized if no user ID in context", func(t *testing.T) {
		req := makeReq(uuid.New().String(), validReqBody(uuid.New().String(), vehicleID), "", "")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Forbidden if user records for another driver", func(t *testing.T) {
		req := makeReq(otherDriverID, validReqBody(otherDriverID, vehicleID), userID, "")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("OK if user records for self", func(t *testing.T) {
		req := makeReq(userID, validReqBody(userID, vehicleID), userID, "")
		mockSvc.On("RecordDutyStatus", mock.Anything, mock.Anything).Return(&pb.RecordDutyStatusResponse{}, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("OK for admin recording for any driver", func(t *testing.T) {
		adminID := uuid.New().String()
		req := makeReq(otherDriverID, validReqBody(otherDriverID, vehicleID), adminID, "fleet_manager")
		mockSvc.On("RecordDutyStatus", mock.Anything, mock.Anything).Return(&pb.RecordDutyStatusResponse{}, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("BadRequest on malformed JSON", func(t *testing.T) {
		req := makeReq(userID, `{invalid}`, userID, "")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestHandlers_GetCurrentDutyStatus(t *testing.T) {
	handler, mockSvc := newTestHandlers()
	router := newTestRouter(handler)

	userID := uuid.New().String()
	driverID := uuid.New().String()

	makeReq := func(driverID, userID, role string) *http.Request {
		req := httptest.NewRequest("GET", "/compliance/duty-status/current/"+driverID, nil)
		ctx := req.Context()
		if userID != "" {
			ctx = context.WithValue(ctx, ctxkeys.UserIDKey, userID)
		}
		if role != "" {
			ctx = context.WithValue(ctx, ctxkeys.RoleKey, role)
		}
		return req.WithContext(ctx)
	}

	t.Run("Unauthorized if no user ID", func(t *testing.T) {
		req := makeReq(driverID, "", "")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Forbidden if viewing another driver's status", func(t *testing.T) {
		req := makeReq(driverID, userID, "")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("OK if viewing own status", func(t *testing.T) {
		req := makeReq(userID, userID, "")
		mockSvc.On("GetCurrentDutyStatus", mock.Anything, userID).Return(&pb.DutyStatusLog{}, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("OK for admin viewing any driver", func(t *testing.T) {
		adminID := uuid.New().String()
		req := makeReq(driverID, adminID, "org_admin")
		mockSvc.On("GetCurrentDutyStatus", mock.Anything, driverID).Return(&pb.DutyStatusLog{}, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
