package client

import (
	"context"
	"time"

	"github.com/fleet-platform/shared/pkg/grpcclient"
	pb "github.com/flyfleet/proto/billing/v1"
	"google.golang.org/grpc"
)

type BillingClient struct {
	conn   *grpc.ClientConn
	client pb.BillingServiceClient
}

func NewBillingClient(baseURL string) (*BillingClient, error) {
	baseClient, err := grpcclient.NewClient(baseURL, 5*time.Second)
	if err != nil {
		return nil, err
	}

	return &BillingClient{
		conn:   baseClient.Conn,
		client: pb.NewBillingServiceClient(baseClient.Conn),
	}, nil
}

func (c *BillingClient) Close() error {
	return c.conn.Close()
}

type BillingMetrics struct {
	TotalOutstanding float64
	OverdueAmount    float64
	MonthlyRevenue   float64
	RiskScore        float64
}

func (c *BillingClient) GetMetrics(ctx context.Context, orgID string) (*BillingMetrics, error) {
	resp, err := c.client.GetMetrics(ctx, &pb.GetMetricsRequest{
		OrgId: orgID,
	})
	if err != nil {
		return &BillingMetrics{}, err
	}

	return &BillingMetrics{
		MonthlyRevenue: resp.Metrics.TotalRevenue,
		RiskScore:      resp.Metrics.RiskScore,
	}, nil
}
