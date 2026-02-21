package client

import (
	"context"
	"time"

	"github.com/fleet-platform/shared/pkg/grpcclient"
	pb "github.com/flyfleet/proto/trip/v1"
	"google.golang.org/grpc"
)

type TripClient struct {
	conn   *grpc.ClientConn
	client pb.TripServiceClient
}

func NewTripClient(baseURL string) (*TripClient, error) {
	baseClient, err := grpcclient.NewClient(baseURL, 5*time.Second)
	if err != nil {
		return nil, err
	}

	return &TripClient{
		conn:   baseClient.Conn,
		client: pb.NewTripServiceClient(baseClient.Conn),
	}, nil
}

func (c *TripClient) Close() error {
	return c.conn.Close()
}

type TripStats struct {
	ActiveTrips int
	TotalTrips  int
	Distance    float64
}

func (c *TripClient) GetStats(ctx context.Context, orgID string) (*TripStats, error) {
	resp, err := c.client.GetTripStats(ctx, &pb.GetTripStatsRequest{
		OrgId: orgID,
	})
	if err != nil {
		return &TripStats{}, err
	}

	return &TripStats{
		ActiveTrips: int(resp.ActiveTrips),
		TotalTrips:  int(resp.TotalTrips),
		Distance:    resp.TotalDistance,
	}, nil
}
