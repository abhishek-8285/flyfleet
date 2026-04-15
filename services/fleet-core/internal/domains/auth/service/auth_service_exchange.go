package service

import (
	"context"
	"errors"

	"github.com/flyfleet/fleet-core/internal/domains/auth/models"
	"github.com/google/uuid"
)

func (s *AuthService) SwitchContext(ctx context.Context, requestorID uuid.UUID, targetUserID uuid.UUID, ipAddress, userAgent string) (*models.AuthResponse, error) {
	// 1. Fetch requesting user (must exist)
	requestor, err := s.repo.GetUserByID(ctx, requestorID)
	if err != nil {
		return nil, errors.New("requesting user not found")
	}

	// 2. Fetch target user
	target, err := s.repo.GetUserByID(ctx, targetUserID)
	if err != nil {
		return nil, errors.New("target user not found")
	}

	// 3. Authorization check
	// The requestor can switch only if they are a 'super_admin' or there is an explicit link
	isSuperAdmin := requestor.Role == "super_admin"
	isLinked := s.repo.IsAccountLinked(ctx, requestor.ID, target.ID)

	if !isSuperAdmin && !isLinked {
		// Log unauthorized attempt as a high-severity security event
		_ = s.repo.LogSecurityEvent(ctx, &models.SecurityEvent{
			UserID:    requestor.ID,
			EventType: "UNAUTHORIZED_CONTEXT_SWITCH_ATTEMPT",
			IPAddress: ipAddress,
			UserAgent: userAgent,
			Metadata:  "target_user_id: " + target.ID.String(),
		})
		return nil, errors.New("unauthorized: neither super_admin nor linked to target account")
	}

	// 4. Log the Context Switch Execution with OnBehalfOf fields
	_ = s.repo.LogSecurityEvent(ctx, &models.SecurityEvent{
		UserID:     target.ID,
		EventType:  "CONTEXT_SWITCH_EXECUTED",
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Metadata:   "switched_to_org: " + target.OrgID.String(),
		ActorID:    &requestor.ID,
		OnBehalfOf: &target.ID,
	})

	// 5. Mint tokens acting as the target, but we should attach an actor_id if possible
	// Right now we will issue standard tokens for the target identity. For full implementation,
	// issueTokens or generateAccessToken could accept an optional ActorID claim.
	// For MVP, we fully inherit the target session securely because the Audit Log holds the truth.

	// Issue tokens utilizing existing auth flow for target user
	return s.issueTokens(ctx, target, ipAddress, userAgent)
}
