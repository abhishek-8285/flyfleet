package grpc

import (
	"context"

	"github.com/flyfleet/fleet-core/internal/domains/reports/models"
	"github.com/flyfleet/fleet-core/internal/domains/reports/service"
	pb "github.com/flyfleet/proto/reports/v1"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ReportsServer struct {
	pb.UnimplementedReportsServiceServer
	service *service.ReportService
	logger  *zap.Logger
}

func NewReportsServer(service *service.ReportService, logger *zap.Logger) *ReportsServer {
	return &ReportsServer{
		service: service,
		logger:  logger,
	}
}

func (s *ReportsServer) CreateReport(ctx context.Context, req *pb.CreateReportRequest) (*pb.CreateReportResponse, error) {
	// tenantID should come from context
	tenantID, _ := uuid.Parse("00000000-0000-0000-0000-000000000001") // Placeholder

	report := &models.ScheduledReport{
		OrganizationID: tenantID,
		Name:           "Generated Report", // Missing in proto, using default
		ReportType:     req.Type,
		Format:         req.Format,
		CronExpression: "0 0 * * *", // Default daily
		Recipients:     []string{},
	}

	id, err := s.service.CreateReport(ctx, report)
	if err != nil {
		return nil, err
	}
	report.ID = id // Assign ID back to object for mapping

	return &pb.CreateReportResponse{
		Report: s.mapReportToProto(report),
	}, nil
}

func (s *ReportsServer) GetReport(ctx context.Context, req *pb.GetReportRequest) (*pb.GetReportResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	report, err := s.service.GetReport(ctx, id)
	if err != nil {
		return nil, err
	}
	return &pb.GetReportResponse{
		Report: s.mapReportToProto(report),
	}, nil
}

func (s *ReportsServer) ListReports(ctx context.Context, req *pb.ListReportsRequest) (*pb.ListReportsResponse, error) {
	reports, total, err := s.service.ListReports(ctx, int(req.Offset), int(req.Limit))
	if err != nil {
		return nil, err
	}

	var pbReports []*pb.Report
	for _, r := range reports {
		pbReports = append(pbReports, s.mapReportToProto(&r))
	}

	return &pb.ListReportsResponse{
		Reports: pbReports,
		Total:   int32(total),
	}, nil
}

func (s *ReportsServer) DownloadReport(ctx context.Context, req *pb.DownloadReportRequest) (*pb.DownloadReportResponse, error) {
	// Placeholder logic as download URL generation logic resides in service/handlers
	return &pb.DownloadReportResponse{
		DownloadUrl:   "https://flyfleet-reports.s3.amazonaws.com/report.pdf",
		ExpirySeconds: 3600,
	}, nil
}

func (s *ReportsServer) mapReportToProto(r *models.ScheduledReport) *pb.Report {
	return &pb.Report{
		Id:        r.ID.String(),
		TenantId:     r.OrganizationID.String(),
		Type:      r.ReportType,
		Format:    r.Format,
		Status:    r.Status,
		CreatedAt: timestamppb.New(r.CreatedAt),
		CreatedBy: r.CreatedBy.String(),
	}
}
