package client

import (
	"context"
	"time"

	"github.com/fleet-platform/shared/pkg/grpcclient"
	pb "github.com/flyfleet/proto/admin/v1"
	"google.golang.org/grpc"
)

type SupportClient struct {
	conn   *grpc.ClientConn
	client pb.AdminSupportServiceClient
}

func NewSupportClient(baseURL string) (*SupportClient, error) {
	baseClient, err := grpcclient.NewClient(baseURL, 5*time.Second)
	if err != nil {
		return nil, err
	}

	return &SupportClient{
		conn:   baseClient.Conn,
		client: pb.NewAdminSupportServiceClient(baseClient.Conn),
	}, nil
}

func (c *SupportClient) Close() error {
	return c.conn.Close()
}

func (c *SupportClient) GetDashboardStats(ctx context.Context, token string) (map[string]interface{}, error) {
	resp, err := c.client.GetDashboardStats(ctx, &pb.GetDashboardStatsRequest{})
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	for k, v := range resp.Stats {
		res[k] = v
	}
	return map[string]interface{}{"data": res}, nil
}

func (c *SupportClient) GetOrganizationStats(ctx context.Context, token string) (map[string]interface{}, error) {
	resp, err := c.client.GetOrganizationStats(ctx, &pb.GetOrganizationStatsRequest{})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_orgs":  resp.TotalOrgs,
		"active_orgs": resp.ActiveOrgs,
	}, nil
}
