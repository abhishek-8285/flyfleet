package compliance

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/fleet-platform/shared/pkg/logger"
)

type GDPRCompliance struct {
	logger logger.Logger
	db     Database
	audit  AuditLogger
}

type Database interface {
	Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error)
	Execute(ctx context.Context, query string, args ...interface{}) error
}

type AuditLogger interface {
	Log(ctx context.Context, event AuditEvent) error
}

type UserData struct {
	UserID        uuid.UUID              `json:"user_id"`
	PersonalData  map[string]interface{} `json:"personal_data"`
	Consent       *ConsentRecord         `json:"consent"`
	DataLocations []DataLocation         `json:"data_locations"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

type ConsentRecord struct {
	ConsentID     string    `json:"consent_id"`
	UserID        uuid.UUID `json:"user_id"`
	Purpose       string    `json:"purpose"`
	Granted       bool      `json:"granted"`
	GrantedAt     time.Time `json:"granted_at"`
	WithdrawnAt   *time.Time `json:"withdrawn_at"`
	ConsentText   string    `json:"consent_text"`
	Version       string    `json:"version"`
}

type DataLocation struct {
	SystemName    string    `json:"system_name"`
	DataType      string    `json:"data_type"`
	StorageType   string    `json:"storage_type"`
	RetentionDays int       `json:"retention_days"`
	Encrypted     bool      `json:"encrypted"`
	LastAccessed  time.Time `json:"last_accessed"`
}

type DataSubjectRequest struct {
	RequestID     string    `json:"request_id"`
	UserID        uuid.UUID `json:"user_id"`
	RequestType   string    `json:"request_type"` // access, rectification, erasure, portability
	Status        string    `json:"status"`       // pending, processing, completed, rejected
	SubmittedAt   time.Time `json:"submitted_at"`
	ProcessedAt   *time.Time `json:"processed_at"`
	ResponseData  string    `json:"response_data"`
}

type AuditEvent struct {
	EventID      string                 `json:"event_id"`
	UserID       *uuid.UUID             `json:"user_id"`
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id"`
	Timestamp    time.Time              `json:"timestamp"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
	Metadata     map[string]interface{} `json:"metadata"`
}

const (
	RequestTypeAccess        = "access"
	RequestTypeRectification = "rectification"
	RequestTypeErasure       = "erasure"
	RequestTypePortability   = "portability"
	
	RequestStatusPending     = "pending"
	RequestStatusProcessing  = "processing"
	RequestStatusCompleted   = "completed"
	RequestStatusRejected    = "rejected"
)

func NewGDPRCompliance(logger logger.Logger, db Database, audit AuditLogger) *GDPRCompliance {
	return &GDPRCompliance{
		logger: logger,
		db:     db,
		audit:  audit,
	}
}

func (gc *GDPRCompliance) RecordConsent(ctx context.Context, userID uuid.UUID, purpose, consentText, version string) (*ConsentRecord, error) {
	consent := &ConsentRecord{
		ConsentID:   uuid.New().String(),
		UserID:      userID,
		Purpose:     purpose,
		Granted:     true,
		GrantedAt:   time.Now(),
		ConsentText: consentText,
		Version:     version,
	}
	
	query := `
		INSERT INTO user_consents (consent_id, user_id, purpose, granted, granted_at, consent_text, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	
	err := gc.db.Execute(ctx, query, 
		consent.ConsentID, userID, purpose, true, consent.GrantedAt, consentText, version)
	if err != nil {
		return nil, fmt.Errorf("failed to record consent: %w", err)
	}
	
	// Log audit event
	auditEvent := AuditEvent{
		EventID:      uuid.New().String(),
		UserID:       &userID,
		Action:       "consent_granted",
		ResourceType: "consent",
		ResourceID:   consent.ConsentID,
		Timestamp:    time.Now(),
		Metadata: map[string]interface{}{
			"purpose": purpose,
			"version": version,
		},
	}
	
	if err := gc.audit.Log(ctx, auditEvent); err != nil {
		gc.logger.Error("Failed to log consent audit event", zap.Error(err))
	}
	
	gc.logger.Info("Consent recorded", 
		zap.String("user_id", userID.String()),
		zap.String("purpose", purpose))
	
	return consent, nil
}

func (gc *GDPRCompliance) WithdrawConsent(ctx context.Context, consentID string) error {
	withdrawnAt := time.Now()
	
	query := `UPDATE user_consents SET granted = false, withdrawn_at = $1 WHERE consent_id = $2`
	err := gc.db.Execute(ctx, query, withdrawnAt, consentID)
	if err != nil {
		return fmt.Errorf("failed to withdraw consent: %w", err)
	}
	
	// Log audit event
	auditEvent := AuditEvent{
		EventID:      uuid.New().String(),
		Action:       "consent_withdrawn",
		ResourceType: "consent",
		ResourceID:   consentID,
		Timestamp:    withdrawnAt,
	}
	
	if err := gc.audit.Log(ctx, auditEvent); err != nil {
		gc.logger.Error("Failed to log consent withdrawal audit event", zap.Error(err))
	}
	
	gc.logger.Info("Consent withdrawn", zap.String("consent_id", consentID))
	return nil
}

func (gc *GDPRCompliance) GetDataSubjectRequest(ctx context.Context, requestID string) (*DataSubjectRequest, error) {
	query := `
		SELECT request_id, user_id, request_type, status, submitted_at, processed_at, response_data
		FROM data_subject_requests 
		WHERE request_id = $1`
	
	results, err := gc.db.Query(ctx, query, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get data subject request: %w", err)
	}
	
	if len(results) == 0 {
		return nil, fmt.Errorf("data subject request not found")
	}
	
	row := results[0]
	request := &DataSubjectRequest{
		RequestID:   row["request_id"].(string),
		UserID:      row["user_id"].(uuid.UUID),
		RequestType: row["request_type"].(string),
		Status:      row["status"].(string),
		SubmittedAt: row["submitted_at"].(time.Time),
	}
	
	if row["processed_at"] != nil {
		processedAt := row["processed_at"].(time.Time)
		request.ProcessedAt = &processedAt
	}
	
	if row["response_data"] != nil {
		request.ResponseData = row["response_data"].(string)
	}
	
	return request, nil
}

func (gc *GDPRCompliance) SubmitDataSubjectRequest(ctx context.Context, userID uuid.UUID, requestType string) (*DataSubjectRequest, error) {
	request := &DataSubjectRequest{
		RequestID:   uuid.New().String(),
		UserID:      userID,
		RequestType: requestType,
		Status:      RequestStatusPending,
		SubmittedAt: time.Now(),
	}
	
	query := `
		INSERT INTO data_subject_requests (request_id, user_id, request_type, status, submitted_at)
		VALUES ($1, $2, $3, $4, $5)`
	
	err := gc.db.Execute(ctx, query, 
		request.RequestID, userID, requestType, RequestStatusPending, request.SubmittedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to submit data subject request: %w", err)
	}
	
	// Log audit event
	auditEvent := AuditEvent{
		EventID:      uuid.New().String(),
		UserID:       &userID,
		Action:       "dsr_submitted",
		ResourceType: "data_subject_request",
		ResourceID:   request.RequestID,
		Timestamp:    time.Now(),
		Metadata: map[string]interface{}{
			"request_type": requestType,
		},
	}
	
	if err := gc.audit.Log(ctx, auditEvent); err != nil {
		gc.logger.Error("Failed to log DSR submission audit event", zap.Error(err))
	}
	
	gc.logger.Info("Data subject request submitted",
		zap.String("request_id", request.RequestID),
		zap.String("user_id", userID.String()),
		zap.String("request_type", requestType))
	
	return request, nil
}

func (gc *GDPRCompliance) ProcessDataSubjectRequest(ctx context.Context, requestID string) error {
	request, err := gc.GetDataSubjectRequest(ctx, requestID)
	if err != nil {
		return err
	}
	
	if request.Status != RequestStatusPending {
		return fmt.Errorf("request is not in pending status")
	}
	
	var responseData string
	
	switch request.RequestType {
	case RequestTypeAccess:
		responseData, err = gc.processAccessRequest(ctx, request.UserID)
	case RequestTypeRectification:
		responseData, err = gc.processRectificationRequest(ctx, request.UserID)
	case RequestTypeErasure:
		err = gc.processErasureRequest(ctx, request.UserID)
	case RequestTypePortability:
		responseData, err = gc.processPortabilityRequest(ctx, request.UserID)
	default:
		return fmt.Errorf("unsupported request type: %s", request.RequestType)
	}
	
	if err != nil {
		return fmt.Errorf("failed to process request: %w", err)
	}
	
	// Update request status
	processedAt := time.Now()
	query := `
		UPDATE data_subject_requests 
		SET status = $1, processed_at = $2, response_data = $3 
		WHERE request_id = $4`
	
	err = gc.db.Execute(ctx, query, RequestStatusCompleted, processedAt, responseData, requestID)
	if err != nil {
		return fmt.Errorf("failed to update request status: %w", err)
	}
	
	// Log audit event
	auditEvent := AuditEvent{
		EventID:      uuid.New().String(),
		UserID:       &request.UserID,
		Action:       "dsr_processed",
		ResourceType: "data_subject_request",
		ResourceID:   requestID,
		Timestamp:    processedAt,
		Metadata: map[string]interface{}{
			"request_type": request.RequestType,
			"status":       RequestStatusCompleted,
		},
	}
	
	if err := gc.audit.Log(ctx, auditEvent); err != nil {
		gc.logger.Error("Failed to log DSR processing audit event", zap.Error(err))
	}
	
	gc.logger.Info("Data subject request processed",
		zap.String("request_id", requestID),
		zap.String("request_type", request.RequestType))
	
	return nil
}

func (gc *GDPRCompliance) processAccessRequest(ctx context.Context, userID uuid.UUID) (string, error) {
	// Collect all personal data for the user
	userData, err := gc.collectUserData(ctx, userID)
	if err != nil {
		return "", err
	}
	
	// Format data for export
	formattedData := gc.formatUserData(userData)
	return formattedData, nil
}

func (gc *GDPRCompliance) processRectificationRequest(ctx context.Context, userID uuid.UUID) (string, error) {
	// In a real implementation, this would update user data
	// For now, we'll just log that rectification was requested
	return "Rectification request received and logged", nil
}

func (gc *GDPRCompliance) processErasureRequest(ctx context.Context, userID uuid.UUID) error {
	// Mark user data for deletion (soft delete)
	query := `UPDATE users SET deleted_at = $1 WHERE id = $2`
	err := gc.db.Execute(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to mark user for deletion: %w", err)
	}
	
	// Queue data for permanent deletion
	gc.queueForPermanentDeletion(userID)
	return nil
}

func (gc *GDPRCompliance) processPortabilityRequest(ctx context.Context, userID uuid.UUID) (string, error) {
	// Similar to access request but in machine-readable format
	userData, err := gc.collectUserData(ctx, userID)
	if err != nil {
		return "", err
	}
	
	// Format as JSON for portability
	jsonData := gc.formatUserDataAsJSON(userData)
	return jsonData, nil
}

func (gc *GDPRCompliance) collectUserData(ctx context.Context, userID uuid.UUID) (*UserData, error) {
	// Query all systems for user data
	// This is a simplified implementation
	userData := &UserData{
		UserID: userID,
		PersonalData: make(map[string]interface{}),
		DataLocations: []DataLocation{
			{
				SystemName:    "user_service",
				DataType:      "personal_information",
				StorageType:   "database",
				RetentionDays: 365,
				Encrypted:     true,
				LastAccessed:  time.Now(),
			},
			{
				SystemName:    "trip_service",
				DataType:      "trip_history",
				StorageType:   "database",
				RetentionDays: 1825, // 5 years
				Encrypted:     true,
				LastAccessed:  time.Now(),
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	return userData, nil
}

func (gc *GDPRCompliance) formatUserData(userData *UserData) string {
	// Simplified formatting
	return fmt.Sprintf("User Data for ID: %s\nData Locations: %d systems\nLast Updated: %s",
		userData.UserID.String(),
		len(userData.DataLocations),
		userData.UpdatedAt.Format(time.RFC3339))
}

func (gc *GDPRCompliance) formatUserDataAsJSON(userData *UserData) string {
	// In real implementation, use proper JSON marshaling
	return fmt.Sprintf(`{"user_id":"%s","data_locations":%d}`, 
		userData.UserID.String(), len(userData.DataLocations))
}

func (gc *GDPRCompliance) queueForPermanentDeletion(userID uuid.UUID) {
	// In real implementation, this would queue data for deletion
	// after retention period expires
	gc.logger.Info("User data queued for permanent deletion", 
		zap.String("user_id", userID.String()))
}

func (gc *GDPRCompliance) GetUserDataLocations(ctx context.Context, userID uuid.UUID) ([]DataLocation, error) {
	// Return all locations where user data is stored
	userData, err := gc.collectUserData(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	return userData.DataLocations, nil
}

func (gc *GDPRCompliance) GetUserConsents(ctx context.Context, userID uuid.UUID) ([]*ConsentRecord, error) {
	query := `
		SELECT consent_id, user_id, purpose, granted, granted_at, withdrawn_at, consent_text, version
		FROM user_consents 
		WHERE user_id = $1 
		ORDER BY granted_at DESC`
	
	results, err := gc.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user consents: %w", err)
	}
	
	var consents []*ConsentRecord
	for _, row := range results {
		consent := &ConsentRecord{
			ConsentID:   row["consent_id"].(string),
			UserID:      row["user_id"].(uuid.UUID),
			Purpose:     row["purpose"].(string),
			Granted:     row["granted"].(bool),
			GrantedAt:   row["granted_at"].(time.Time),
			ConsentText: row["consent_text"].(string),
			Version:     row["version"].(string),
		}
		
		if row["withdrawn_at"] != nil {
			withdrawnAt := row["withdrawn_at"].(time.Time)
			consent.WithdrawnAt = &withdrawnAt
		}
		
		consents = append(consents, consent)
	}
	
	return consents, nil
}