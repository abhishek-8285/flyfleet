package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fleet-platform/shared/pkg/grpcclient"
	pb "github.com/flyfleet/auth-service/proto/auth/v1"
	"github.com/google/uuid"
)

type AuthClient struct {
	client pb.AuthServiceClient
	base   *grpcclient.Client
}

func NewAuthClient(baseURL string) (*AuthClient, error) {
	// e.g. "auth-service:50051"
	baseClient, err := grpcclient.NewClient(baseURL, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth grpc via unified client: %w", err)
	}

	return &AuthClient{
		client: pb.NewAuthServiceClient(baseClient.Conn),
		base:   baseClient,
	}, nil
}

func (c *AuthClient) Close() {
	if c.base != nil {
		c.base.Close()
	}
}

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	OrgID     string    `json:"org_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

type RegisterResponse struct {
	User User `json:"user"`
}

func (c *AuthClient) Login(ctx context.Context, email, password, ip, userAgent string) (*LoginResponse, error) {
	var resp *pb.LoginResponse
	err := c.base.Call(ctx, func(ctx context.Context) error {
		var errCall error
		resp, errCall = c.client.Login(ctx, &pb.LoginRequest{
			Email:     email,
			Password:  password,
			Ip:        ip,
			UserAgent: userAgent,
		})
		return errCall
	})
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		User: User{
			ID:        resp.User.Id,
			Email:     resp.User.Email,
			Name:      resp.User.Name,
			Role:      resp.User.Role,
			OrgID:     resp.User.OrgId,
			CreatedAt: resp.User.CreatedAt.AsTime(),
			UpdatedAt: resp.User.UpdatedAt.AsTime(),
		},
	}, nil
}

func (c *AuthClient) Register(ctx context.Context, email, password, name, ip, userAgent string) (*RegisterResponse, error) {
	var resp *pb.RegisterResponse
	err := c.base.Call(ctx, func(ctx context.Context) error {
		var errCall error
		resp, errCall = c.client.Register(ctx, &pb.RegisterRequest{
			Email:     email,
			Password:  password,
			Name:      name,
			Ip:        ip,
			UserAgent: userAgent,
		})
		return errCall
	})
	if err != nil {
		return nil, err
	}

	return &RegisterResponse{
		User: User{
			ID:        resp.User.Id,
			Email:     resp.User.Email,
			Name:      resp.User.Name,
			Role:      resp.User.Role,
			OrgID:     resp.User.OrgId,
			CreatedAt: resp.User.CreatedAt.AsTime(),
			UpdatedAt: resp.User.UpdatedAt.AsTime(),
		},
	}, nil
}

type Organization struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Domain       string    `json:"domain"`
	DefaultTheme string    `json:"default_theme"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (c *AuthClient) GetOrganization(ctx context.Context, id, token string) (*Organization, error) {
	var resp *pb.GetOrganizationResponse
	err := c.base.Call(ctx, func(ctx context.Context) error {
		var errCall error
		resp, errCall = c.client.GetOrganization(ctx, &pb.GetOrganizationRequest{Id: id})
		return errCall
	})
	if err != nil {
		return nil, err
	}

	oID, _ := uuid.Parse(resp.Organization.Id)
	return &Organization{
		ID:           oID,
		Name:         resp.Organization.Name,
		Domain:       resp.Organization.Domain,
		DefaultTheme: resp.Organization.DefaultTheme,
		Status:       resp.Organization.Status,
		CreatedAt:    resp.Organization.CreatedAt.AsTime(),
		UpdatedAt:    resp.Organization.UpdatedAt.AsTime(),
	}, nil
}

func (c *AuthClient) ListOrganizations(ctx context.Context, page, limit int, token string) ([]Organization, int, error) {
	var resp *pb.ListOrganizationsResponse
	err := c.base.Call(ctx, func(ctx context.Context) error {
		var errCall error
		resp, errCall = c.client.ListOrganizations(ctx, &pb.ListOrganizationsRequest{
			Page:  int32(page),
			Limit: int32(limit),
		})
		return errCall
	})
	if err != nil {
		return nil, 0, err
	}

	var orgs []Organization
	for _, o := range resp.Organizations {
		oID, _ := uuid.Parse(o.Id)
		orgs = append(orgs, Organization{
			ID:           oID,
			Name:         o.Name,
			Domain:       o.Domain,
			DefaultTheme: o.DefaultTheme,
			Status:       o.Status,
			CreatedAt:    o.CreatedAt.AsTime(),
			UpdatedAt:    o.UpdatedAt.AsTime(),
		})
	}

	return orgs, int(resp.Total), nil
}

type UserStats struct {
	TotalUsers    int `json:"total_users"`
	ActiveUsers   int `json:"active_users"`
	TotalDrivers  int `json:"total_drivers"`
	TotalOrgs     int `json:"total_orgs"`
	ActiveOrgs    int `json:"active_orgs"`
	PendingOrgs   int `json:"pending_orgs"`
	InactiveOrgs  int `json:"inactive_orgs"`
}

func (c *AuthClient) GetStats(ctx context.Context, token string) (*UserStats, error) {
	var resp *pb.GetStatsResponse
	err := c.base.Call(ctx, func(ctx context.Context) error {
		var errCall error
		resp, errCall = c.client.GetStats(ctx, &pb.GetStatsRequest{})
		return errCall
	})
	if err != nil {
		return &UserStats{}, err
	}

	return &UserStats{
		TotalUsers:   int(resp.TotalUsers),
		ActiveUsers:  int(resp.ActiveUsers),
		TotalDrivers: int(resp.TotalDrivers),
		TotalOrgs:    int(resp.TotalOrgs),
		ActiveOrgs:   int(resp.ActiveOrgs),
		PendingOrgs:  int(resp.PendingOrgs),
		InactiveOrgs: int(resp.InactiveOrgs),
	}, nil
}

func (c *AuthClient) GetSecurityTelemetry(ctx context.Context, token string) (map[string]any, error) {
	var resp *pb.GetSecurityTelemetryResponse
	err := c.base.Call(ctx, func(ctx context.Context) error {
		var errCall error
		resp, errCall = c.client.GetSecurityTelemetry(ctx, &pb.GetSecurityTelemetryRequest{})
		return errCall
	})
	if err != nil {
		return nil, err
	}
	return resp.Telemetry.AsMap(), nil
}

type AuditLog struct {
	ID        uuid.UUID      `json:"id"`
	Action    string         `json:"action"`
	Entity    string         `json:"entity"`
	EntityID  string         `json:"entity_id"`
	ActorID   uuid.UUID      `json:"actor_id"`
	CreatedAt time.Time      `json:"created_at"`
	Metadata  map[string]any `json:"metadata"`
}

func (c *AuthClient) GetAuditLogs(ctx context.Context, page, size int, token string) ([]AuditLog, error) {
	var resp *pb.GetAuditLogsResponse
	err := c.base.Call(ctx, func(ctx context.Context) error {
		var errCall error
		resp, errCall = c.client.GetAuditLogs(ctx, &pb.GetAuditLogsRequest{
			Page:  int32(page),
			Limit: int32(size),
		})
		return errCall
	})
	if err != nil {
		return nil, err
	}

	var logs []AuditLog
	for _, l := range resp.Logs {
		logID, _ := uuid.Parse(l.Id)
		actID, _ := uuid.Parse(l.ActorId)

		var metadata map[string]any
		if l.Metadata != nil {
			metaMap := l.Metadata.AsMap()
			if raw, ok := metaMap["raw"].(string); ok {
				json.Unmarshal([]byte(raw), &metadata)
			}
		}

		logs = append(logs, AuditLog{
			ID:        logID,
			Action:    l.Action,
			Entity:    l.Entity,
			EntityID:  l.EntityId,
			ActorID:   actID,
			CreatedAt: l.CreatedAt.AsTime(),
			Metadata:  metadata,
		})
	}

	return logs, nil
}
