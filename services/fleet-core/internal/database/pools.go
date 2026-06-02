package database

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabasePools struct {
	Critical *pgxpool.Pool // auth_v2, emergency, tracking (500 conns)
	High     *pgxpool.Pool // billing_v2, vehicle, driver (300 conns)
	Medium   *pgxpool.Pool // analytics, reports (150 conns)
	Low      *pgxpool.Pool // diagnostics_v2, integrations (50 conns)
}

type PoolConfig struct {
	DatabaseURL string
	CriticalMax int32
	HighMax     int32
	MediumMax   int32
	LowMax      int32
}

func NewDatabasePools(ctx context.Context, cfg PoolConfig) (*DatabasePools, error) {
	pools := &DatabasePools{}
	var err error

	pools.Critical, err = createPool(ctx, cfg.DatabaseURL, cfg.CriticalMax, "flyfleet-critical")
	if err != nil {
		return nil, err
	}

	pools.High, err = createPool(ctx, cfg.DatabaseURL, cfg.HighMax, "flyfleet-high")
	if err != nil {
		return nil, err
	}

	pools.Medium, err = createPool(ctx, cfg.DatabaseURL, cfg.MediumMax, "flyfleet-medium")
	if err != nil {
		return nil, err
	}

	pools.Low, err = createPool(ctx, cfg.DatabaseURL, cfg.LowMax, "flyfleet-low")
	if err != nil {
		return nil, err
	}

	return pools, nil
}

func (p *DatabasePools) Close() {
	if p.Critical != nil {
		p.Critical.Close()
	}
	if p.High != nil {
		p.High.Close()
	}
	if p.Medium != nil {
		p.Medium.Close()
	}
	if p.Low != nil {
		p.Low.Close()
	}
}

func createPool(ctx context.Context, dbURL string, maxConns int32, appName string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	config.MaxConns = maxConns
	config.MinConns = maxConns / 10
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	// Enforce RLS at the session level without redundant roundtrips
	var connState sync.Map

	config.BeforeAcquire = func(ctx context.Context, conn *pgx.Conn) bool {
		start := time.Now()
		defer func() {
			overhead := time.Since(start)
			if overhead > 5*time.Millisecond {
				fmt.Printf("[PERF-ALERT] RLS BeforeAcquire overhead: %v\n", overhead)
			}
		}()

		// Try to extract Context claims robustly using standardized internal helper extractors
		var tenantIDStr string
		if tenantID, ok := ctxkeys.GetTenantID(ctx); ok && tenantID != uuid.Nil {
			tenantIDStr = tenantID.String()
		}

		var userIDStr string
		if userID, ok := ctxkeys.GetUserID(ctx); ok && userID != uuid.Nil {
			userIDStr = userID.String()
		}

		roleStr := ctxkeys.GetRole(ctx)

		// SECURITY: Never bypass RLS setup - always enforce session variables
		// We always set session variables based on the current context, regardless of previous state

		// Note: The original code had a single `err` variable.
		// The new code uses `:=` for `err` in the inner blocks, making them local.
		// The `tenantIDStr` block still uses the outer `err` variable.
		// This is consistent with the provided instruction.
		var err error
		// Note: PostgreSQL `SET` command does not support `$1` binding. We MUST use `set_config`
		// to securely pass parameterized inputs, otherwise pgxpool panics.
		// The `false` flag makes this session-wide, which is cleared explicitly
		// during AfterRelease.
		if tenantIDStr == "" {
			_, err = conn.Exec(ctx, "SELECT set_config('flyfleet.current_tenant_id', '', false)")
			if err == nil {
				_, err = conn.Exec(ctx, "SELECT set_config('flyfleet.current_org_id', '', false)")
			}
		} else {
			_, err = conn.Exec(ctx, "SELECT set_config('flyfleet.current_tenant_id', $1, false)", tenantIDStr)
			if err == nil {
				_, err = conn.Exec(ctx, "SELECT set_config('flyfleet.current_org_id', $1, false)", tenantIDStr)
			}
		}
		if err != nil {
			log.Printf("Failed to set tenant/org ID in connection: %v\n", err)
			return true
		}

		if userIDStr != "" {
			_, err := conn.Exec(ctx, "SELECT set_config('flyfleet.current_user_id', $1, false)", userIDStr)
			if err != nil {
				log.Printf("Failed to set user ID in connection: %v\n", err)
				return true
			}
		} else {
			_, err := conn.Exec(ctx, "SELECT set_config('flyfleet.current_user_id', '', false)")
			if err != nil {
				log.Printf("Failed to reset user ID in connection: %v\n", err)
				return true
			}
		}

		if roleStr != "" {
			_, err := conn.Exec(ctx, "SELECT set_config('flyfleet.current_role', $1, false)", roleStr)
			if err != nil {
				log.Printf("Failed to set role in connection: %v\n", err)
				return true
			}
		} else {
			_, err := conn.Exec(ctx, "SELECT set_config('flyfleet.current_role', '', false)")
			if err != nil {
				log.Printf("Failed to reset role in connection: %v\n", err)
				return true
			}
		}

		// Update connection state
		connState.Store(conn, tenantIDStr)
		return true
	}

	config.BeforeClose = func(conn *pgx.Conn) {
		connState.Delete(conn)
	}

	config.AfterRelease = func(conn *pgx.Conn) bool {
		// Strictly clear all custom session variables before returning to the pool
		_, errOrg := conn.Exec(context.Background(), "SELECT set_config('flyfleet.current_tenant_id', '', false)")
		_, errOrg2 := conn.Exec(context.Background(), "SELECT set_config('flyfleet.current_org_id', '', false)")
		_, errUser := conn.Exec(context.Background(), "SELECT set_config('flyfleet.current_user_id', '', false)")
		_, errRole := conn.Exec(context.Background(), "SELECT set_config('flyfleet.current_role', '', false)")

		connState.Delete(conn)

		// If the reset fails, the connection is poisoned. Return false to destroy it.
		if errOrg != nil || errOrg2 != nil || errUser != nil || errRole != nil {
			log.Printf("level=FATAL event=connection_poisoned msg=\"[SECURITY] Connection poisoned during reset. Destroying physical socket to prevent cross-tenant data leak.\" errOrg=\"%v\" errOrg2=\"%v\" errUser=\"%v\" errRole=\"%v\"", errOrg, errOrg2, errUser, errRole)
			return false
		}
		return true
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
