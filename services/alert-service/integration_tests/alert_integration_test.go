package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/flyfleet/alert-service/internal/handlers"
	"github.com/flyfleet/alert-service/internal/models"
	"github.com/flyfleet/alert-service/internal/providers"
	"github.com/flyfleet/alert-service/internal/repository"
	"github.com/flyfleet/alert-service/internal/rule_engine"
	"github.com/flyfleet/alert-service/internal/service"
)

type AlertServiceTestSuite struct {
	suite.Suite
	db                  *pgxpool.Pool
	router              *gin.Engine
	alertRepo           *repository.AlertRepository
	alertService        *service.AlertService
	alertHandler        *handlers.AlertHandler
	testOrgID           uuid.UUID
	testUserID          uuid.UUID
	logger              *zap.Logger
	natsConn            *nats.Conn
	notificationService *providers.NotificationService
}

func (suite *AlertServiceTestSuite) SetupSuite() {
	// Initialize logger
	logger, _ := zap.NewDevelopment()
	suite.logger = logger

	// Connect to test database
	dbURL := ""
	db, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		suite.T().Fatal("Failed to connect to test database:", err)
	}
	suite.db = db

	// Test database connection
	if err := db.Ping(context.Background()); err != nil {
		suite.T().Fatal("Failed to ping test database:", err)
	}

	// Connect to NATS
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		suite.T().Fatal("Failed to connect to NATS:", err)
	}
	suite.natsConn = nc

	// Initialize components
	suite.alertRepo = repository.NewAlertRepository(db)
	
	// Mock notification service for testing
	mockSMSService := &providers.MockSMSService{}
	mockEmailService := &providers.MockEmailService{}
	mockPushService := &MockPushService{}
	suite.notificationService = providers.NewNotificationService(mockSMSService, mockEmailService, mockPushService, logger)
	
	ruleEngine := rule_engine.NewRuleEngine(suite.alertRepo, logger)
	suite.alertService = service.NewAlertService(suite.alertRepo, ruleEngine, suite.notificationService, logger)
	suite.alertHandler = handlers.NewAlertHandler(suite.alertService, logger)

	// Create test organization and user IDs
	suite.testOrgID = uuid.New()
	suite.testUserID = uuid.New()

	// Setup router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	suite.setupRoutes()
}

func (suite *AlertServiceTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
	if suite.natsConn != nil {
		suite.natsConn.Close()
	}
	if suite.logger != nil {
		suite.logger.Sync()
	}
}

func (suite *AlertServiceTestSuite) setupRoutes() {
	// Health check
	suite.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "alert-service-test",
			"timestamp": time.Now().Unix(),
		})
	})

	// Alert API endpoints
	api := suite.router.Group("/api/v1")
	{
		alerts := api.Group("/alerts")
		{
			// Alert rules management
			alerts.POST("/rules", suite.alertHandler.CreateRule)
			alerts.GET("/rules", suite.alertHandler.ListRules)
			alerts.GET("/rules/:id", suite.alertHandler.GetRule)
			alerts.PUT("/rules/:id", suite.alertHandler.UpdateRule)
			alerts.DELETE("/rules/:id", suite.alertHandler.DeleteRule)
			alerts.POST("/rules/:id/enable", suite.alertHandler.EnableRule)
			alerts.POST("/rules/:id/disable", suite.alertHandler.DisableRule)

			// Alert instances
			alerts.GET("", suite.alertHandler.ListAlerts)
			alerts.GET("/:id", suite.alertHandler.GetAlert)
			alerts.POST("/:id/acknowledge", suite.alertHandler.AcknowledgeAlert)
			alerts.POST("/:id/resolve", suite.alertHandler.ResolveAlert)

			// Alert statistics
			alerts.GET("/stats/summary", suite.alertHandler.GetAlertSummary)
		}
	}
}

func (suite *AlertServiceTestSuite) TestHealthCheck() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "healthy", response["status"])
	assert.Equal(suite.T(), "alert-service-test", response["service"])
}

func (suite *AlertServiceTestSuite) TestCreateAlertRule() {
	ruleReq := models.CreateRuleRequest{
		Name:        "Test Overspeed Rule",
		Type:        models.AlertTypeOverspeed,
		Description: "Test rule for overspeed detection",
		Conditions: map[string]interface{}{
			"speed_limit": 80,
			"min_duration": 300,
		},
		Severity: models.SeverityWarning,
		Actions: []models.AlertAction{
			{
				Type:     models.ActionTypeSMS,
				Target:   "+1234567890",
				Template: "overspeed_template",
			},
		},
		Vehicles:  []uuid.UUID{},
		DriverIDs: []uuid.UUID{},
	}

	jsonData, _ := json.Marshal(ruleReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/alerts/rules", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	// Add context values that would come from auth middleware
	req.Header.Set("X-Org-ID", suite.testOrgID.String())
	req.Header.Set("X-User-ID", suite.testUserID.String())

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	
	var createdRule models.AlertRule
	json.Unmarshal(w.Body.Bytes(), &createdRule)
	assert.Equal(suite.T(), ruleReq.Name, createdRule.Name)
	assert.Equal(suite.T(), ruleReq.Type, createdRule.Type)
	assert.Equal(suite.T(), ruleReq.Severity, createdRule.Severity)
	assert.True(suite.T(), createdRule.Enabled)
}

func (suite *AlertServiceTestSuite) TestGetAlertRule() {
	// First create a rule
	ruleReq := models.CreateRuleRequest{
		Name:        "Test Rule for Get",
		Type:        models.AlertTypeGeofenceEntry,
		Description: "Test rule for getting by ID",
		Conditions: map[string]interface{}{
			"geofence_id": uuid.New().String(),
		},
		Severity: models.SeverityInfo,
		Actions: []models.AlertAction{
			{
				Type:     models.ActionTypeEmail,
				Target:   "test@example.com",
				Template: "geofence_template",
			},
		},
		Vehicles:  []uuid.UUID{},
		DriverIDs: []uuid.UUID{},
	}

	createdRule, err := suite.alertService.CreateRule(context.Background(), suite.testOrgID, &ruleReq, suite.testUserID)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), createdRule.ID)

	// Now get the rule
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/alerts/rules/%s", createdRule.ID), nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var retrievedRule models.AlertRule
	json.Unmarshal(w.Body.Bytes(), &retrievedRule)
	assert.Equal(suite.T(), createdRule.ID, retrievedRule.ID)
	assert.Equal(suite.T(), createdRule.Name, retrievedRule.Name)
}

func (suite *AlertServiceTestSuite) TestListAlertRules() {
	// Create multiple rules for testing
	rules := []models.CreateRuleRequest{
		{
			Name:        "Rule 1",
			Type:        models.AlertTypeOverspeed,
			Description: "First test rule",
			Conditions:  map[string]interface{}{"speed_limit": 80},
			Severity:    models.SeverityWarning,
			Actions:     []models.AlertAction{{Type: models.ActionTypeSMS, Target: "+1234567890"}},
			Vehicles:    []uuid.UUID{},
			DriverIDs:   []uuid.UUID{},
		},
		{
			Name:        "Rule 2",
			Type:        models.AlertTypeGeofenceExit,
			Description: "Second test rule",
			Conditions:  map[string]interface{}{"geofence_id": uuid.New().String()},
			Severity:    models.SeverityCritical,
			Actions:     []models.AlertAction{{Type: models.ActionTypeEmail, Target: "test@example.com"}},
			Vehicles:    []uuid.UUID{},
			DriverIDs:   []uuid.UUID{},
		},
	}

	for _, ruleReq := range rules {
		_, err := suite.alertService.CreateRule(context.Background(), suite.testOrgID, &ruleReq, suite.testUserID)
		assert.NoError(suite.T(), err)
	}

	// List rules
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/alerts/rules", nil)
	req.Header.Set("X-Org-ID", suite.testOrgID.String())
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(suite.T(), response, "data")
	assert.Contains(suite.T(), response, "total")
	
	data := response["data"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(data), 2)
}

func (suite *AlertServiceTestSuite) TestUpdateAlertRule() {
	// Create a rule first
	ruleReq := models.CreateRuleRequest{
		Name:        "Original Rule",
		Type:        models.AlertTypeHarshBraking,
		Description: "Rule to be updated",
		Conditions:  map[string]interface{}{"min_deceleration": -0.8},
		Severity:    models.SeverityWarning,
		Actions:     []models.AlertAction{{Type: models.ActionTypeSMS, Target: "+1234567890"}},
		Vehicles:    []uuid.UUID{},
		DriverIDs:   []uuid.UUID{},
	}

	createdRule, err := suite.alertService.CreateRule(context.Background(), suite.testOrgID, &ruleReq, suite.testUserID)
	assert.NoError(suite.T(), err)

	// Update the rule
	updateReq := models.UpdateRuleRequest{
		Name:        stringPtr("Updated Rule Name"),
		Description: stringPtr("Updated description"),
		Severity:    stringPtr(models.SeverityCritical),
	}

	jsonData, _ := json.Marshal(updateReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/alerts/rules/%s", createdRule.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var updatedRule models.AlertRule
	json.Unmarshal(w.Body.Bytes(), &updatedRule)
	assert.Equal(suite.T(), "Updated Rule Name", updatedRule.Name)
	assert.Equal(suite.T(), "Updated description", updatedRule.Description)
	assert.Equal(suite.T(), models.SeverityCritical, updatedRule.Severity)
}

func (suite *AlertServiceTestSuite) TestDeleteAlertRule() {
	// Create a rule first
	ruleReq := models.CreateRuleRequest{
		Name:        "Rule to Delete",
		Type:        models.AlertTypeIdleTime,
		Description: "Rule that will be deleted",
		Conditions:  map[string]interface{}{"max_idle_time": 1800},
		Severity:    models.SeverityInfo,
		Actions:     []models.AlertAction{{Type: models.ActionTypeEmail, Target: "test@example.com"}},
		Vehicles:    []uuid.UUID{},
		DriverIDs:   []uuid.UUID{},
	}

	createdRule, err := suite.alertService.CreateRule(context.Background(), suite.testOrgID, &ruleReq, suite.testUserID)
	assert.NoError(suite.T(), err)

	// Delete the rule
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/alerts/rules/%s", createdRule.ID), nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNoContent, w.Code)

	// Verify rule is deleted
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/alerts/rules/%s", createdRule.ID), nil)
	suite.router.ServeHTTP(w2, req2)
	assert.Equal(suite.T(), http.StatusNotFound, w2.Code)
}

func (suite *AlertServiceTestSuite) TestEnableDisableRule() {
	// Create a disabled rule
	ruleReq := models.CreateRuleRequest{
		Name:        "Toggle Rule",
		Type:        models.AlertTypeFuelTheft,
		Description: "Rule to test enable/disable",
		Conditions:  map[string]interface{}{"min_fuel_drop": 5.0},
		Severity:    models.SeverityCritical,
		Actions:     []models.AlertAction{{Type: models.ActionTypeSMS, Target: "+1234567890"}},
		Vehicles:    []uuid.UUID{},
		DriverIDs:   []uuid.UUID{},
	}

	createdRule, err := suite.alertService.CreateRule(context.Background(), suite.testOrgID, &ruleReq, suite.testUserID)
	assert.NoError(suite.T(), err)

	// Initially should be enabled
	assert.True(suite.T(), createdRule.Enabled)

	// Disable the rule
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/alerts/rules/%s/disable", createdRule.ID), nil)
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Verify it's disabled
	disabledRule, err := suite.alertService.GetRule(context.Background(), createdRule.ID)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), disabledRule.Enabled)

	// Enable the rule
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/alerts/rules/%s/enable", createdRule.ID), nil)
	suite.router.ServeHTTP(w2, req2)
	assert.Equal(suite.T(), http.StatusOK, w2.Code)

	// Verify it's enabled
	enabledRule, err := suite.alertService.GetRule(context.Background(), createdRule.ID)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), enabledRule.Enabled)
}

func (suite *AlertServiceTestSuite) TestCreateAlert() {
	// Create a rule first
	ruleReq := models.CreateRuleRequest{
		Name:        "Alert Creation Rule",
		Type:        models.AlertTypeOverspeed,
		Description: "Rule for testing alert creation",
		Conditions:  map[string]interface{}{"speed_limit": 80},
		Severity:    models.SeverityWarning,
		Actions:     []models.AlertAction{{Type: models.ActionTypeSMS, Target: "+1234567890"}},
		Vehicles:    []uuid.UUID{},
		DriverIDs:   []uuid.UUID{},
	}

	rule, err := suite.alertService.CreateRule(context.Background(), suite.testOrgID, &ruleReq, suite.testUserID)
	assert.NoError(suite.T(), err)

	// Create an alert
	alert := &models.Alert{
		ID:        uuid.New().String(),
		OrgID:     suite.testOrgID,
		RuleID:    rule.ID,
		Type:      models.AlertTypeOverspeed,
		Severity:  models.SeverityWarning,
		VehicleID: &[]uuid.UUID{uuid.New()}[0],
		DriverID:  &[]uuid.UUID{uuid.New()}[0],
		Location: &models.Location{
			Latitude:  40.7128,
			Longitude: -74.0060,
			Address:   "New York, NY",
		},
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"speed":     85.5,
			"limit":     80,
			"duration":  350,
			"vehicle_id": "V123",
		},
		Status: models.AlertStatusNew,
	}

	err = suite.alertService.CreateAlert(context.Background(), alert)
	assert.NoError(suite.T(), err)

	// Verify alert was created
	createdAlert, err := suite.alertService.GetAlert(context.Background(), alert.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), alert.ID, createdAlert.ID)
	assert.Equal(suite.T(), alert.Type, createdAlert.Type)
	assert.Equal(suite.T(), alert.Severity, createdAlert.Severity)
	assert.Equal(suite.T(), models.AlertStatusNew, createdAlert.Status)
}

func (suite *AlertServiceTestSuite) TestListAlerts() {
	// Create a rule
	ruleReq := models.CreateRuleRequest{
		Name:        "List Alerts Rule",
		Type:        models.AlertTypeGeofenceEntry,
		Description: "Rule for testing alert listing",
		Conditions:  map[string]interface{}{"geofence_id": uuid.New().String()},
		Severity:    models.SeverityInfo,
		Actions:     []models.AlertAction{{Type: models.ActionTypeEmail, Target: "test@example.com"}},
		Vehicles:    []uuid.UUID{},
		DriverIDs:   []uuid.UUID{},
	}

	rule, err := suite.alertService.CreateRule(context.Background(), suite.testOrgID, &ruleReq, suite.testUserID)
	assert.NoError(suite.T(), err)

	// Create multiple alerts
	alerts := []*models.Alert{
		{
			ID:        uuid.New().String(),
			OrgID:     suite.testOrgID,
			RuleID:    rule.ID,
			Type:      models.AlertTypeGeofenceEntry,
			Severity:  models.SeverityInfo,
			VehicleID: &[]uuid.UUID{uuid.New()}[0],
			DriverID:  &[]uuid.UUID{uuid.New()}[0],
			Location:  &models.Location{Latitude: 40.7128, Longitude: -74.0060, Address: "New York, NY"},
			Timestamp: time.Now().Add(-2 * time.Hour),
			Details:   map[string]interface{}{"event": "entry"},
			Status:    models.AlertStatusNew,
		},
		{
			ID:        uuid.New().String(),
			OrgID:     suite.testOrgID,
			RuleID:    rule.ID,
			Type:      models.AlertTypeGeofenceEntry,
			Severity:  models.SeverityInfo,
			VehicleID: &[]uuid.UUID{uuid.New()}[0],
			DriverID:  &[]uuid.UUID{uuid.New()}[0],
			Location:  &models.Location{Latitude: 40.7589, Longitude: -73.9851, Address: "Manhattan, NY"},
			Timestamp: time.Now().Add(-1 * time.Hour),
			Details:   map[string]interface{}{"event": "entry"},
			Status:    models.AlertStatusAcknowledged,
		},
	}

	for _, alert := range alerts {
		err := suite.alertService.CreateAlert(context.Background(), alert)
		assert.NoError(suite.T(), err)
	}

	// List alerts
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/alerts", nil)
	req.Header.Set("X-Org-ID", suite.testOrgID.String())
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(suite.T(), response, "data")
	assert.Contains(suite.T(), response, "total")
	
	data := response["data"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(data), 2)
}

func (suite *AlertServiceTestSuite) TestAcknowledgeAlert() {
	// Create a rule and alert
	ruleReq := models.CreateRuleRequest{
		Name:        "Acknowledge Test Rule",
		Type:        models.AlertTypeHarshAcceleration,
		Description: "Rule for testing acknowledgment",
		Conditions:  map[string]interface{}{"min_acceleration": 0.5},
		Severity:    models.SeverityWarning,
		Actions:     []models.AlertAction{{Type: models.ActionTypeSMS, Target: "+1234567890"}},
		Vehicles:    []uuid.UUID{},
		DriverIDs:   []uuid.UUID{},
	}

	rule, err := suite.alertService.CreateRule(context.Background(), suite.testOrgID, &ruleReq, suite.testUserID)
	assert.NoError(suite.T(), err)

	alert := &models.Alert{
		ID:        uuid.New().String(),
		OrgID:     suite.testOrgID,
		RuleID:    rule.ID,
		Type:      models.AlertTypeHarshAcceleration,
		Severity:  models.SeverityWarning,
		VehicleID: &[]uuid.UUID{uuid.New()}[0],
		DriverID:  &[]uuid.UUID{uuid.New()}[0],
		Location:  &models.Location{Latitude: 40.7128, Longitude: -74.0060, Address: "New York, NY"},
		Timestamp: time.Now(),
		Details:   map[string]interface{}{"acceleration": 0.6},
		Status:    models.AlertStatusNew,
	}

	err = suite.alertService.CreateAlert(context.Background(), alert)
	assert.NoError(suite.T(), err)

	// Acknowledge the alert
	ackReq := map[string]string{"notes": "Reviewed and acknowledged"}
	jsonData, _ := json.Marshal(ackReq)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/alerts/%s/acknowledge", alert.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", suite.testUserID.String())
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Verify alert status was updated
	acknowledgedAlert, err := suite.alertService.GetAlert(context.Background(), alert.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), models.AlertStatusAcknowledged, acknowledgedAlert.Status)
	assert.Equal(suite.T(), suite.testUserID, *acknowledgedAlert.AcknowledgedBy)
}

func (suite *AlertServiceTestSuite) TestResolveAlert() {
	// Create a rule and alert
	ruleReq := models.CreateRuleRequest{
		Name:        "Resolve Test Rule",
		Type:        models.AlertTypeLowBattery,
		Description: "Rule for testing resolution",
		Conditions:  map[string]interface{}{"min_battery_level": 20.0},
		Severity:    models.SeverityCritical,
		Actions:     []models.AlertAction{{Type: models.ActionTypeEmail, Target: ""}},
		Vehicles:    []uuid.UUID{},
		DriverIDs:   []uuid.UUID{},
	}

	rule, err := suite.alertService.CreateRule(context.Background(), suite.testOrgID, &ruleReq, suite.testUserID)
	assert.NoError(suite.T(), err)

	alert := &models.Alert{
		ID:        uuid.New().String(),
		OrgID:     suite.testOrgID,
		RuleID:    rule.ID,
		Type:      models.AlertTypeLowBattery,
		Severity:  models.SeverityCritical,
		VehicleID: &[]uuid.UUID{uuid.New()}[0],
		DriverID:  &[]uuid.UUID{uuid.New()}[0],
		Location:  &models.Location{Latitude: 40.7128, Longitude: -74.0060, Address: "New York, NY"},
		Timestamp: time.Now(),
		Details:   map[string]interface{}{"battery_level": 15.5},
		Status:    models.AlertStatusAcknowledged,
	}

	err = suite.alertService.CreateAlert(context.Background(), alert)
	assert.NoError(suite.T(), err)

	// Resolve the alert
	resolveReq := map[string]string{"notes": "Issue resolved by charging battery"}
	jsonData, _ := json.Marshal(resolveReq)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/alerts/%s/resolve", alert.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", suite.testUserID.String())
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Verify alert status was updated
	resolvedAlert, err := suite.alertService.GetAlert(context.Background(), alert.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), models.AlertStatusResolved, resolvedAlert.Status)
	assert.Equal(suite.T(), suite.testUserID, *resolvedAlert.ResolvedBy)
}

func (suite *AlertServiceTestSuite) TestGetAlertSummary() {
	// Create some alerts for summary testing
	ruleReq := models.CreateRuleRequest{
		Name:        "Summary Test Rule",
		Type:        models.AlertTypeOverspeed,
		Description: "Rule for testing summary",
		Conditions:  map[string]interface{}{"speed_limit": 80},
		Severity:    models.SeverityWarning,
		Actions:     []models.AlertAction{{Type: models.ActionTypeSMS, Target: "+1234567890"}},
		Vehicles:    []uuid.UUID{},
		DriverIDs:   []uuid.UUID{},
	}

	rule, err := suite.alertService.CreateRule(context.Background(), suite.testOrgID, &ruleReq, suite.testUserID)
	assert.NoError(suite.T(), err)

	// Create alerts with different statuses and severities
	alerts := []*models.Alert{
		{
			ID:        uuid.New().String(),
			OrgID:     suite.testOrgID,
			RuleID:    rule.ID,
			Type:      models.AlertTypeOverspeed,
			Severity:  models.SeverityWarning,
			VehicleID: &[]uuid.UUID{uuid.New()}[0],
			DriverID:  &[]uuid.UUID{uuid.New()}[0],
			Location:  &models.Location{Latitude: 40.7128, Longitude: -74.0060, Address: "New York, NY"},
			Timestamp: time.Now().Add(-24 * time.Hour),
			Details:   map[string]interface{}{"speed": 85.0},
			Status:    models.AlertStatusNew,
		},
		{
			ID:        uuid.New().String(),
			OrgID:     suite.testOrgID,
			RuleID:    rule.ID,
			Type:      models.AlertTypeOverspeed,
			Severity:  models.SeverityCritical,
			VehicleID: &[]uuid.UUID{uuid.New()}[0],
			DriverID:  &[]uuid.UUID{uuid.New()}[0],
			Location:  &models.Location{Latitude: 40.7589, Longitude: -73.9851, Address: "Manhattan, NY"},
			Timestamp: time.Now().Add(-12 * time.Hour),
			Details:   map[string]interface{}{"speed": 95.0},
			Status:    models.AlertStatusAcknowledged,
		},
		{
			ID:        uuid.New().String(),
			OrgID:     suite.testOrgID,
			RuleID:    rule.ID,
			Type:      models.AlertTypeOverspeed,
			Severity:  models.SeverityInfo,
			VehicleID: &[]uuid.UUID{uuid.New()}[0],
			DriverID:  &[]uuid.UUID{uuid.New()}[0],
			Location:  &models.Location{Latitude: 40.7505, Longitude: -73.9934, Address: "Brooklyn, NY"},
			Timestamp: time.Now().Add(-6 * time.Hour),
			Details:   map[string]interface{}{"speed": 82.0},
			Status:    models.AlertStatusResolved,
		},
	}

	for _, alert := range alerts {
		err := suite.alertService.CreateAlert(context.Background(), alert)
		assert.NoError(suite.T(), err)
	}

	// Get alert summary
	fromDate := time.Now().Add(-48 * time.Hour).Format(time.RFC3339)
	toDate := time.Now().Format(time.RFC3339)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/alerts/stats/summary?from_date=%s&to_date=%s", fromDate, toDate), nil)
	req.Header.Set("X-Org-ID", suite.testOrgID.String())
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var summary models.AlertSummary
	json.Unmarshal(w.Body.Bytes(), &summary)
	assert.Equal(suite.T(), 3, summary.TotalAlerts)
	assert.Equal(suite.T(), 1, summary.NewAlerts)
	assert.Equal(suite.T(), 1, summary.Acknowledged)
	assert.Equal(suite.T(), 1, summary.Resolved)
	assert.Equal(suite.T(), 1, summary.BySeverity[models.SeverityCritical])
	assert.Equal(suite.T(), 1, summary.BySeverity[models.SeverityWarning])
	assert.Equal(suite.T(), 1, summary.BySeverity[models.SeverityInfo])
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

// Mock services for testing
type MockSMSService struct{}

func (m *MockSMSService) Send(to, message string) error {
	return nil
}

type MockEmailService struct{}

func (m *MockEmailService) Send(to, subject, body string) error {
	return nil
}

type MockPushService struct{}

func (m *MockPushService) SendPush(token, title, body string, data map[string]interface{}) error {
	return nil
}

func TestAlertServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AlertServiceTestSuite))
}