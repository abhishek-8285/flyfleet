package client

import (
	"context"
	"time"

	"github.com/fleet-platform/shared/pkg/grpcclient"
	pb "github.com/flyfleet/proto/alert/v1"
	"google.golang.org/grpc"
)

type AlertClient struct {
	conn   *grpc.ClientConn
	client pb.AlertServiceClient
}

func NewAlertClient(baseURL string) (*AlertClient, error) {
	baseClient, err := grpcclient.NewClient(baseURL, 5*time.Second)
	if err != nil {
		return nil, err
	}

	return &AlertClient{
		conn:   baseClient.Conn,
		client: pb.NewAlertServiceClient(baseClient.Conn),
	}, nil
}

func (c *AlertClient) Close() error {
	return c.conn.Close()
}

type AlertStats struct {
	TotalAlerts    int
	CriticalAlerts int
	OpenAlerts     int
	NewAlerts      int
}

func (c *AlertClient) GetStats(ctx context.Context, orgID string) (*AlertStats, error) {
	resp, err := c.client.GetAlertStats(ctx, &pb.GetAlertStatsRequest{
		OrgId: orgID,
	})
	if err != nil {
		return &AlertStats{}, err
	}

	return &AlertStats{
		TotalAlerts:    int(resp.TotalAlerts),
		CriticalAlerts: int(resp.CriticalAlerts),
		OpenAlerts:     int(resp.OpenAlerts),
		NewAlerts:      int(resp.OpenAlerts), // Map Open to New for now
	}, nil
}
