package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/flyfleet/auth-service/internal/models"
	"github.com/google/uuid"
)

// Error definitions
var (
	ErrInvalidCIDR      = errors.New("invalid IP address or CIDR range")
	ErrInvalidAction    = errors.New("action must be ALLOW or BLOCK")
	ErrDuplicateIPRule  = errors.New("IP rule already exists for this CIDR")
	ErrInvalidOrgID     = errors.New("organization ID is required")
	ErrInvalidUserID    = errors.New("user ID is required")
	ErrInvalidReviewID  = errors.New("review ID is required")
	ErrInvalidSessionID = errors.New("session ID is required")
	ErrInvalidStatus    = errors.New("invalid status value")
)

// ValidStatuses for access reviews
var ValidReviewStatuses = map[string]bool{
	"PENDING":     true,
	"IN_PROGRESS": true,
	"COMPLETED":   true,
}

var ValidItemStatuses = map[string]bool{
	"PENDING":  true,
	"APPROVED": true,
	"REVOKED":  true,
}

// IP Rules

func (s *AuthService) ListIPRules(ctx context.Context, orgID uuid.UUID) ([]models.IPRule, error) {
	if orgID == uuid.Nil {
		return nil, ErrInvalidOrgID
	}
	return s.repo.ListIPRules(ctx, orgID)
}

func (s *AuthService) GetIPRuleByID(ctx context.Context, id uuid.UUID) (*models.IPRule, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidReviewID
	}
	return s.repo.GetIPRuleByID(ctx, id)
}

func (s *AuthService) CreateIPRule(ctx context.Context, rule *models.IPRule, adminID uuid.UUID, clientIP string) error {
	// Validate org ID
	if rule.OrgID == uuid.Nil {
		return ErrInvalidOrgID
	}

	// Validate and normalize CIDR format
	normalizedCIDR, err := s.validateAndNormalizeCIDR(rule.CIDR)
	if err != nil {
		return err
	}
	rule.CIDR = normalizedCIDR

	// Validate action
	action := strings.ToUpper(rule.Action)
	if action != "ALLOW" && action != "BLOCK" {
		return ErrInvalidAction
	}
	rule.Action = action

	// Check for duplicates
	existingRules, err := s.repo.ListIPRules(ctx, rule.OrgID)
	if err == nil {
		for _, existing := range existingRules {
			if existing.CIDR == rule.CIDR {
				return ErrDuplicateIPRule
			}
		}
	}

	// Create the rule
	if err := s.repo.CreateIPRule(ctx, rule); err != nil {
		return err
	}

	// Audit log
	s.logSecurityAudit(rule.OrgID, adminID, "ip_rule_create", "ip_rules",
		fmt.Sprintf("Created IP rule: %s (%s)", rule.CIDR, rule.Action), clientIP)

	return nil
}

func (s *AuthService) CreateIPRuleNoAudit(ctx context.Context, rule *models.IPRule) error {
	return s.repo.CreateIPRule(ctx, rule)
}

func (s *AuthService) DeleteIPRule(ctx context.Context, id uuid.UUID, adminID uuid.UUID, clientIP string) error {
	if id == uuid.Nil {
		return ErrInvalidReviewID
	}

	// Get the rule for audit logging
	rule, err := s.repo.GetIPRuleByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteIPRule(ctx, id); err != nil {
		return err
	}

	// Audit log
	s.logSecurityAudit(rule.OrgID, adminID, "ip_rule_delete", "ip_rules",
		fmt.Sprintf("Deleted IP rule: %s (%s)", rule.CIDR, rule.Action), clientIP)

	return nil
}

func (s *AuthService) validateAndNormalizeCIDR(cidr string) (string, error) {
	cidr = strings.TrimSpace(cidr)
	if cidr == "" {
		return "", ErrInvalidCIDR
	}

	// Try parsing as CIDR first
	_, ipNet, err := net.ParseCIDR(cidr)
	if err == nil {
		return ipNet.String(), nil
	}

	// Try as single IP
	ip := net.ParseIP(cidr)
	if ip == nil {
		return "", ErrInvalidCIDR
	}

	// Convert single IP to CIDR
	if ip.To4() != nil {
		return cidr + "/32", nil
	}
	return cidr + "/128", nil
}

// Access Reviews

func (s *AuthService) ListAccessReviews(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.AccessReview, int, error) {
	if orgID == uuid.Nil {
		return nil, 0, ErrInvalidOrgID
	}
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return s.repo.ListAccessReviews(ctx, orgID, page, size)
}

func (s *AuthService) GetAccessReviewByID(ctx context.Context, id uuid.UUID) (*models.AccessReview, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidReviewID
	}

	review, err := s.repo.GetAccessReviewByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Fetch review items
	items, err := s.repo.ListReviewItems(ctx, id)
	if err != nil {
		return review, nil // Return review even if items fail
	}
	review.ReviewItems = items
	return review, nil
}

func (s *AuthService) CreateAccessReview(ctx context.Context, review *models.AccessReview, adminID uuid.UUID, clientIP string) error {
	if review.OrgID == uuid.Nil {
		return ErrInvalidOrgID
	}
	if strings.TrimSpace(review.Title) == "" {
		return errors.New("review title is required")
	}

	if review.Status == "" {
		review.Status = "PENDING"
	}
	if !ValidReviewStatuses[review.Status] {
		return ErrInvalidStatus
	}

	if err := s.repo.CreateAccessReview(ctx, review); err != nil {
		return err
	}

	// Audit log
	s.logSecurityAudit(review.OrgID, adminID, "access_review_create", "access_reviews",
		fmt.Sprintf("Created access review: %s", review.Title), clientIP)

	return nil
}

func (s *AuthService) UpdateAccessReviewStatus(ctx context.Context, id uuid.UUID, status string, adminID uuid.UUID, clientIP string) error {
	if id == uuid.Nil {
		return ErrInvalidReviewID
	}
	if !ValidReviewStatuses[status] {
		return ErrInvalidStatus
	}

	// Get review for audit
	review, err := s.repo.GetAccessReviewByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.UpdateAccessReviewStatus(ctx, id, status); err != nil {
		return err
	}

	// Audit log
	s.logSecurityAudit(review.OrgID, adminID, "access_review_update", "access_reviews",
		fmt.Sprintf("Updated review status: %s -> %s", review.Status, status), clientIP)

	return nil
}

func (s *AuthService) UpdateReviewItem(ctx context.Context, itemID uuid.UUID, status string, reviewerID uuid.UUID, notes string, clientIP string) error {
	if itemID == uuid.Nil {
		return errors.New("item ID is required")
	}
	if !ValidItemStatuses[status] {
		return ErrInvalidStatus
	}
	if reviewerID == uuid.Nil {
		return ErrInvalidUserID
	}

	if err := s.repo.UpdateReviewItem(ctx, itemID, status, reviewerID, notes); err != nil {
		return err
	}

	// Audit log - note: would need to fetch item to get org_id
	// For now, log with reviewer's context
	s.logSecurityAudit(uuid.Nil, reviewerID, "review_item_update", "review_items",
		fmt.Sprintf("Updated review item %s: %s", itemID, status), clientIP)

	return nil
}

// Sessions - Extended for Admin management

func (s *AuthService) ListAllSessions(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.AuthToken, int, error) {
	if orgID == uuid.Nil {
		return nil, 0, ErrInvalidOrgID
	}
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return s.repo.ListAllSessions(ctx, orgID, page, size)
}

func (s *AuthService) RevokeSession(ctx context.Context, sessionID uuid.UUID, adminID uuid.UUID, clientIP string) error {
	if sessionID == uuid.Nil {
		return ErrInvalidSessionID
	}

	if err := s.repo.RevokeSession(ctx, sessionID); err != nil {
		return err
	}

	// Audit log
	s.logSecurityAudit(uuid.Nil, adminID, "session_revoke", "sessions",
		fmt.Sprintf("Revoked session: %s", sessionID), clientIP)

	return nil
}

func (s *AuthService) RevokeSessionNoAudit(ctx context.Context, sessionID uuid.UUID) error {
	return s.repo.RevokeSession(ctx, sessionID)
}

func (s *AuthService) RevokeAllSessionsForOrg(ctx context.Context, orgID uuid.UUID, adminID uuid.UUID, clientIP string) error {
	if orgID == uuid.Nil {
		return ErrInvalidOrgID
	}

	if err := s.repo.RevokeAllSessionsForOrg(ctx, orgID); err != nil {
		return err
	}

	// Audit log
	s.logSecurityAudit(orgID, adminID, "sessions_revoke_all", "sessions",
		"Revoked all sessions for organization", clientIP)

	return nil
}

func (s *AuthService) ListLoginEvents(ctx context.Context, orgID uuid.UUID, page, size int, status string) ([]models.LoginEvent, int, error) {
	if orgID == uuid.Nil {
		return nil, 0, ErrInvalidOrgID
	}
	filter := make(map[string]any)
	if status != "" {
		filter["status"] = status
	}
	return s.repo.ListLoginEvents(ctx, orgID, page, size, filter)
}

func (s *AuthService) ListThreatAlerts(ctx context.Context, orgID uuid.UUID, page, size int, status string) ([]models.ThreatAlert, int, error) {
	if orgID == uuid.Nil {
		return nil, 0, ErrInvalidOrgID
	}
	filter := make(map[string]any)
	if status != "" {
		filter["status"] = status
	}
	return s.repo.ListThreatAlerts(ctx, orgID, page, size, filter)
}

func (s *AuthService) UpdateThreatAlertStatus(ctx context.Context, alertID uuid.UUID, status string) error {
	if alertID == uuid.Nil {
		return errors.New("alert ID is required")
	}
	return s.repo.UpdateThreatAlertStatus(ctx, alertID, status)
}

func (s *AuthService) ListMFAUsers(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.MFAUser, int, error) {
	if orgID == uuid.Nil {
		return nil, 0, ErrInvalidOrgID
	}
	return s.repo.ListMFAUsers(ctx, orgID, page, size)
}

// logSecurityAudit creates an audit log entry for security actions
func (s *AuthService) logSecurityAudit(orgID, userID uuid.UUID, action, resource, details, clientIP string) {
	log := &models.AuditLog{
		ID:        uuid.New(),
		OrgID:     orgID,
		UserID:    userID,
		Action:    action,
		Resource:  resource,
		Severity:  "info",
		Status:    "success",
		Details:   details,
		IPAddress: clientIP,
		CreatedAt: time.Now(),
	}

	// Fire and forget - don't fail the main operation if audit logging fails
	go func() {
		_ = s.repo.CreateAuditLog(context.Background(), log)
	}()
}
