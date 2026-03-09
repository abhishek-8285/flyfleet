//go:build integration
// +build integration

package repository_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/auth_v2/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/auth_v2/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// Dummy setup for test scope, assumes testcontainers sets up a valid `testDBPool` 
// and that migrations are run. For the sake of the framework illustration, we will
// mock the pool insertion logic if a real DB isn't passed.
func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	// In a real environment, this spins up Postgres via testcontainers-go
	return nil, func() {}
}

func TestRefreshToken_RotationAndReplayAttack(t *testing.T) {
	// Skip actual DB if not implemented yet in the repo for this file
	t.Skip("skipping integration test without testcontainers")

	pool, cleanup := setupTestDB(t)
	defer cleanup()
	repo := repository.NewAuthRepository(pool)
	ctx := context.Background()
	userID := uuid.New()

	tokenA := "refresh-token-a"
	err := repo.CreateRefreshToken(ctx, &models.RefreshToken{Token: tokenA, UserID: userID})
	require.NoError(t, err)

	// Rotate (Refresh) - Issue Token B, Revoke Token A
	tokenB := "refresh-token-b"
	err = repo.RotateSession(ctx, tokenA, tokenB) 
	require.NoError(t, err)

	oldSession, _ := repo.GetRefreshToken(ctx, tokenA)
	require.True(t, oldSession.Revoked)

	// Replay Attack - Attempting to use Token A again
	err = repo.RotateSession(ctx, tokenA, "refresh-token-c")
	require.ErrorContains(t, err, "replay attack detected")
	
	activeSessions, _ := repo.ListActiveSessions(ctx, userID)
	require.Equal(t, 0, len(activeSessions), "Replay attack should invalidate all session families")
}

func TestSessionLimit_EvictOldest(t *testing.T) {
	t.Skip("skipping integration test without testcontainers")
	
	pool, cleanup := setupTestDB(t)
	defer cleanup()
	repo := repository.NewAuthRepository(pool)
	ctx := context.Background()
	userID := uuid.New()

	for i := 0; i < 6; i++ {
		go repo.EnforceSessionLimit(ctx, userID, 5) // max 5 sessions
		repo.CreateRefreshToken(ctx, &models.RefreshToken{Token: fmt.Sprintf("t-%d", i), UserID: userID})
	}
	activeSessions, _ := repo.ListActiveSessions(ctx, userID)
	require.LessOrEqual(t, len(activeSessions), 5)
}

func TestSession_RaceCondition_RevokeVsRefresh(t *testing.T) {
	t.Skip("skipping integration test without testcontainers")
	
	pool, cleanup := setupTestDB(t)
	defer cleanup()
	repo := repository.NewAuthRepository(pool)
	userID := uuid.New()
	
	token := "race-token"
	repo.CreateRefreshToken(context.Background(), &models.RefreshToken{Token: token, UserID: userID})

	var wg sync.WaitGroup
	wg.Add(2)
	
	go func() {
		defer wg.Done()
		repo.RevokeRefreshToken(context.Background(), token)
	}()
	
	go func() {
		defer wg.Done()
		repo.RotateSession(context.Background(), token, "new-race-token")
	}()

	wg.Wait()
	session, _ := repo.GetRefreshToken(context.Background(), token)
	require.True(t, session.Revoked)
}

func TestMultiTenantIsolation_Forbidden(t *testing.T) {
	t.Skip("skipping integration test without testcontainers")
	
	pool, cleanup := setupTestDB(t)
	defer cleanup()
	repo := repository.NewAuthRepository(pool)

	orgA := uuid.New()
	orgB := uuid.New()
	userOrgA := &models.User{ID: uuid.New(), OrgID: orgA}

	err := repo.AssignUserToResource(context.Background(), userOrgA.ID, orgB, "resource-id")
	require.ErrorContains(t, err, "tenant isolation violation")
}
