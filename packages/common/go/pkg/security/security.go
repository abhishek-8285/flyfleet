package security

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ────────────────────────────────────────────────────
// PII Masking
// ────────────────────────────────────────────────────

var (
	// Improved email regex with Unicode support for internationalized domain names (IDN)
	// Matches standard email formats including Unicode characters in localpart and domain
	emailRe = regexp.MustCompile(`[\p{L}0-9._%+\-]+@[\p{L}0-9.\-]+\.[\p{L}]{2,}`)
	// Improved phone regex: E.164 format (max 15 digits, country code starting 1-9)
	phoneRe = regexp.MustCompile(`\+?[1-9]\d{1,14}`)
)

// MaskPII replaces PII patterns (emails, phone numbers) in a string with masked versions.
func MaskPII(input string) string {
	masked := emailRe.ReplaceAllStringFunc(input, func(email string) string {
		parts := strings.SplitN(email, "@", 2)
		if len(parts) != 2 || len(parts[0]) < 2 {
			return "***@***"
		}
		return string(parts[0][0]) + "***@" + parts[1]
	})

	masked = phoneRe.ReplaceAllStringFunc(masked, func(phone string) string {
		if len(phone) <= 4 {
			return "****"
		}
		return strings.Repeat("*", len(phone)-4) + phone[len(phone)-4:]
	})

	return masked
}

// PIIField wraps a zap.Field to mask PII before logging.
func PIIField(key, value string) zap.Field {
	return zap.String(key, MaskPII(value))
}

// ────────────────────────────────────────────────────
// gRPC Tenant Isolation Interceptor
// ────────────────────────────────────────────────────

// OrgValidator validates that the caller is authorized to access the given organization.
// Implementations should extract the caller's identity from the context and verify
// membership in the organization (or super-admin privileges).
type OrgValidator interface {
	// ValidateOrgAccess checks if the caller (identified by context) can access tenantID.
	// Returns nil if access is granted, or an error if denied.
	ValidateOrgAccess(ctx context.Context, tenantID string) error
}

// tenantValidator wraps an OrgValidator to be used in the interceptor.
type tenantValidator struct {
	validator OrgValidator
}

// TenantUnaryInterceptor enforces tenant_id presence AND validates membership.
// It requires a validator to check that the caller is authorized for the org.
// Extracts tenant_id from "x-org-id" or "x-organization-id" metadata keys.
func TenantUnaryInterceptor(validator OrgValidator) grpc.UnaryServerInterceptor {
	if validator == nil {
		panic("TenantUnaryInterceptor: validator is required (nil provided)")
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		tenantID := extractTenantIDFromMetadata(ctx)
		if tenantID == "" {
			return nil, status.Error(codes.PermissionDenied, "organization context is required")
		}

		// Validate that the caller is authorized for this organization
		if err := validator.ValidateOrgAccess(ctx, tenantID); err != nil {
			return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("not authorized for organization: %v", err))
		}

		// Inject tenant_id into context for downstream use
		ctx = context.WithValue(ctx, tenantIDCtxKey, tenantID)
		return handler(ctx, req)
	}
}

type contextKey string

const tenantIDCtxKey contextKey = "tenant_id"

// GetTenantIDFromCtx retrieves the tenant_id from a gRPC context.
func GetTenantIDFromCtx(ctx context.Context) string {
	if v, ok := ctx.Value(tenantIDCtxKey).(string); ok {
		return v
	}
	return ""
}

func extractTenantIDFromMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	for _, key := range []string{"x-organization-id", "x-org-id"} {
		if vals := md.Get(key); len(vals) > 0 && vals[0] != "" {
			return vals[0]
		}
	}
	return ""
}

// ────────────────────────────────────────────────────
// Immutable Audit Logs (Hash Chaining)
// ────────────────────────────────────────────────────

// AuditEntry represents an immutable, hash-chained audit log entry.
type AuditEntry struct {
	ID            string                 `json:"id"`
	Timestamp     time.Time              `json:"timestamp"`
	TenantID         string                 `json:"tenant_id"`
	UserID        string                 `json:"user_id"`
	Action        string                 `json:"action"`
	Resource      string                 `json:"resource"`
	ResourceID    string                 `json:"resource_id"`
	Data          map[string]interface{} `json:"data,omitempty"`
	PreviousHash  string                 `json:"previous_hash"`
	Hash          string                 `json:"hash"`
}

// AuditLogger maintains hash-chained audit logs.
type AuditLogger struct {
	lastHash string
	logger   *zap.Logger
}

// NewAuditLogger creates a new AuditLogger with the genesis hash.
func NewAuditLogger(logger *zap.Logger) *AuditLogger {
	return &AuditLogger{
		lastHash: "0000000000000000000000000000000000000000000000000000000000000000",
		logger:   logger,
	}
}

// Log creates a new audit entry with hash chaining.
func (al *AuditLogger) Log(tenantID, userID, action, resource, resourceID string, data map[string]interface{}) *AuditEntry {
	entry := &AuditEntry{
		ID:           uuid.New().String(),
		Timestamp:    time.Now().UTC(),
		TenantID:        tenantID,
		UserID:       userID,
		Action:       action,
		Resource:     resource,
		ResourceID:   resourceID,
		Data:         data,
		PreviousHash: al.lastHash,
	}

	// Compute hash of this entry (excluding the hash field itself)
	payload, _ := json.Marshal(map[string]interface{}{
		"id":            entry.ID,
		"timestamp":     entry.Timestamp,
		"tenant_id":        entry.TenantID,
		"user_id":       entry.UserID,
		"action":        entry.Action,
		"resource":      entry.Resource,
		"resource_id":   entry.ResourceID,
		"data":          entry.Data,
		"previous_hash": entry.PreviousHash,
	})
	hash := sha256.Sum256(payload)
	entry.Hash = hex.EncodeToString(hash[:])

	// Update chain
	al.lastHash = entry.Hash

	// Structured log output
	al.logger.Info("AUDIT",
		zap.String("audit_id", entry.ID),
		zap.String("tenant_id", entry.TenantID),
		zap.String("user_id", MaskPII(entry.UserID)),
		zap.String("action", entry.Action),
		zap.String("resource", entry.Resource),
		zap.String("resource_id", entry.ResourceID),
		zap.String("hash", entry.Hash),
		zap.String("prev_hash", entry.PreviousHash),
	)

	return entry
}

// VerifyChain validates that a sequence of audit entries maintains hash integrity.
func VerifyChain(entries []AuditEntry) (bool, int) {
	for i := 1; i < len(entries); i++ {
		if entries[i].PreviousHash != entries[i-1].Hash {
			return false, i
		}
	}
	return true, -1
}
