package server

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/flyfleet/proto/dispatch/v1"
	"github.com/flyfleet/services/dispatch"
)

type GRPCServer struct {
	pb.UnimplementedDispatchServiceServer
	dispatcher *dispatch.Dispatcher
	server     *grpc.Server
}

func NewGRPCServer(dispatcher *dispatch.Dispatcher) *GRPCServer {
	return &GRPCServer{
		dispatcher: dispatcher,
		server:     grpc.NewServer(),
	}
}

func (s *GRPCServer) Start(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	pb.RegisterDispatchServiceServer(s.server, s)

	return s.server.Serve(lis)
}

func (s *GRPCServer) Stop() {
	s.server.GracefulStop()
}

func (s *GRPCServer) Dispatch(ctx context.Context, req *pb.DispatchRequest) (*pb.DispatchResponse, error) {
	rideReq := &dispatch.RideRequest{
		RideID:          req.RideId,
		RiderID:         req.RiderId,
		ZoneID:          req.ZoneId,
		Lat:             req.PickupLat,
		Lng:             req.PickupLng,
		SearchRadiusKm:  req.SearchRadiusKm,
		SurgeMultiplier: req.SurgeMultiplier,
		DriverDistances: req.DriverDistances,
		DriverRatings:   req.DriverRatings,
	}

	result, err := s.dispatcher.Dispatch(ctx, rideReq)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DispatchResponse{
		DispatchId:        result.DispatchID,
		RideId:            result.RideID,
		DriverId:          result.DriverID,
		State:             string(result.State),
		Score:             result.Score,
		Variant:           result.Variant,
		Strategy:          pb.DispatchStrategy(result.Strategy),
		Attempt:           int32(result.Attempt),
		OfferType:         result.OfferType,
		TotalCandidates:   int32(result.TotalCandidates),
		DispatchLatencyMs: result.DispatchLatencyMs,
	}, nil
}

func (s *GRPCServer) OfferRide(ctx context.Context, req *pb.OfferRideRequest) (*pb.OfferRideResponse, error) {
	rideReq := &dispatch.RideRequest{
		RideID: req.RideId,
		RiderID: req.RiderId,
		Lat:     req.PickupLat,
		Lng:     req.PickupLng,
	}

	accepted, err := s.dispatcher.OfferRide(ctx, rideReq, req.DriverId, int(req.TimeoutSeconds))
	if err != nil {
		return nil, err
	}

	return &pb.OfferRideResponse{
		Accepted: accepted,
		DriverId: req.DriverId,
		RideId:   req.RideId,
	}, nil
}

func (s *GRPCServer) FindNearbyDrivers(ctx context.Context, req *pb.FindNearbyDriversRequest) (*pb.FindNearbyDriversResponse, error) {
	rideReq := &dispatch.RideRequest{
		RideID:         req.RideId,
		RiderID:        req.RiderId,
		ZoneID:         req.ZoneId,
		Lat:            req.PickupLat,
		Lng:            req.PickupLng,
		SearchRadiusKm: req.SearchRadiusKm,
	}

	drivers, err := s.dispatcher.FindNearbyDrivers(ctx, rideReq)
	if err != nil {
		return nil, err
	}

	return &pb.FindNearbyDriversResponse{
		DriverIds: drivers,
		TotalFound: int32(len(drivers)),
	}, nil
}

func (s *GRPCServer) GetDispatchResult(ctx context.Context, req *pb.GetDispatchResultRequest) (*pb.GetDispatchResultResponse, error) {
	result, err := s.dispatcher.GetDispatchResult(ctx, req.RideId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	var dispatchedAt *timestamppb.Timestamp
	if result.DispatchedAt != nil {
		dispatchedAt = timestamppb.New(*result.DispatchedAt)
	}

	return &pb.GetDispatchResultResponse{
		Result: &pb.DispatchResponse{
			DispatchId:        result.DispatchID,
			RideId:            result.RideID,
			DriverId:          result.DriverID,
			State:             string(result.State),
			Score:             result.Score,
			Variant:           result.Variant,
			Strategy:          pb.DispatchStrategy(result.Strategy),
			Attempt:           int32(result.Attempt),
			OfferType:         result.OfferType,
			TotalCandidates:   int32(result.TotalCandidates),
			DispatchLatencyMs: result.DispatchLatencyMs,
		},
		DispatchedAt: dispatchedAt,
	}, nil
}