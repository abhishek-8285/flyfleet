package grpc

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core-service-new/internal/models"
	"github.com/flyfleet/fleet-core-service-new/internal/service"
	pb "github.com/flyfleet/fleet-core-service-new/proto/fleet/v1"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FleetServer struct {
	pb.UnimplementedFleetServiceServer
	fleetService *service.FleetService
}

func NewFleetServer(fleetService *service.FleetService) *FleetServer {
	return &FleetServer{
		fleetService: fleetService,
	}
}

func (s *FleetServer) CreateVehicle(ctx context.Context, req *pb.CreateVehicleRequest) (*pb.CreateVehicleResponse, error) {
	orgID, err := uuid.Parse(req.OrgId)
	if err != nil {
		return nil, err
	}

	v := &models.Vehicle{
		OrgID:        orgID,
		LicensePlate: req.LicensePlate,
		Model:        req.Model,
		Make:         req.Make,
		Year:         req.Year,
		VIN:          req.Vin,
		Status:       "active",
	}

	if err := s.fleetService.CreateVehicle(ctx, v); err != nil {
		return nil, err
	}

	return &pb.CreateVehicleResponse{Vehicle: vehicleModelToProto(v)}, nil
}

func (s *FleetServer) GetVehicle(ctx context.Context, req *pb.GetVehicleRequest) (*pb.GetVehicleResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	v, err := s.fleetService.GetVehicle(ctx, id)
	if err != nil {
		return nil, err
	}

	return &pb.GetVehicleResponse{Vehicle: vehicleModelToProto(v)}, nil
}

func (s *FleetServer) ListVehicles(ctx context.Context, req *pb.ListVehiclesRequest) (*pb.ListVehiclesResponse, error) {
	orgID, err := uuid.Parse(req.OrgId)
	if err != nil {
		return nil, err
	}

	offset := (req.Page - 1) * req.Limit
	if offset < 0 {
		offset = 0
	}

	vehicles, total, err := s.fleetService.ListVehicles(ctx, orgID, req.Limit, offset)
	if err != nil {
		return nil, err
	}

	var protoVehicles []*pb.Vehicle
	for i := range vehicles {
		protoVehicles = append(protoVehicles, vehicleModelToProto(&vehicles[i]))
	}

	return &pb.ListVehiclesResponse{Vehicles: protoVehicles, Total: total}, nil
}

func (s *FleetServer) UpdateVehicle(ctx context.Context, req *pb.UpdateVehicleRequest) (*pb.UpdateVehicleResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	// Fetch existing
	v, err := s.fleetService.GetVehicle(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Model != nil {
		v.Model = *req.Model
	}
	if req.Make != nil {
		v.Make = *req.Make
	}
	if req.Year != nil {
		v.Year = *req.Year
	}
	if req.Status != nil {
		v.Status = *req.Status
	}

	if err := s.fleetService.UpdateVehicle(ctx, v); err != nil {
		return nil, err
	}

	return &pb.UpdateVehicleResponse{Vehicle: vehicleModelToProto(v)}, nil
}

func (s *FleetServer) DeleteVehicle(ctx context.Context, req *pb.DeleteVehicleRequest) (*pb.DeleteVehicleResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	if err := s.fleetService.DeleteVehicle(ctx, id); err != nil {
		return nil, err
	}

	return &pb.DeleteVehicleResponse{}, nil
}

func (s *FleetServer) AssignDriver(ctx context.Context, req *pb.AssignDriverRequest) (*pb.AssignDriverResponse, error) {
	vid, err := uuid.Parse(req.VehicleId)
	if err != nil {
		return nil, err
	}
	did, err := uuid.Parse(req.DriverId)
	if err != nil {
		return nil, err
	}

	if err := s.fleetService.AssignDriver(ctx, vid, did); err != nil {
		return nil, err
	}

	// Fetch updated vehicle
	v, err := s.fleetService.GetVehicle(ctx, vid)
	if err != nil {
		return nil, err
	}

	return &pb.AssignDriverResponse{Vehicle: vehicleModelToProto(v)}, nil
}

func (s *FleetServer) CreateDriver(ctx context.Context, req *pb.CreateDriverRequest) (*pb.CreateDriverResponse, error) {
	orgID, err := uuid.Parse(req.OrgId)
	if err != nil {
		return nil, err
	}

	d := &models.Driver{
		OrgID:         orgID,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Email:         req.Email,
		Phone:         req.Phone,
		LicenseNumber: req.LicenseNumber,
		LicenseType:   req.LicenseType,
		LicenseExpiry: req.LicenseExpiry.AsTime(),
		Status:        "active",
	}

	if err := s.fleetService.CreateDriver(ctx, d); err != nil {
		return nil, err
	}

	return &pb.CreateDriverResponse{Driver: driverModelToProto(d)}, nil
}

func (s *FleetServer) GetDriver(ctx context.Context, req *pb.GetDriverRequest) (*pb.GetDriverResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	d, err := s.fleetService.GetDriver(ctx, id)
	if err != nil {
		return nil, err
	}

	return &pb.GetDriverResponse{Driver: driverModelToProto(d)}, nil
}

func (s *FleetServer) ListDrivers(ctx context.Context, req *pb.ListDriversRequest) (*pb.ListDriversResponse, error) {
	orgID, err := uuid.Parse(req.OrgId)
	if err != nil {
		return nil, err
	}

	status := ""
	if req.Status != nil {
		status = *req.Status
	}

	offset := (req.Page - 1) * req.Limit
	if offset < 0 {
		offset = 0
	}

	drivers, total, err := s.fleetService.ListDrivers(ctx, orgID, req.Limit, offset, status)
	if err != nil {
		return nil, err
	}

	var protoDrivers []*pb.Driver
	for i := range drivers {
		protoDrivers = append(protoDrivers, driverModelToProto(&drivers[i]))
	}

	return &pb.ListDriversResponse{Drivers: protoDrivers, Total: total}, nil
}

func (s *FleetServer) UpdateDriver(ctx context.Context, req *pb.UpdateDriverRequest) (*pb.UpdateDriverResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	d, err := s.fleetService.GetDriver(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.FirstName != nil {
		d.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		d.LastName = *req.LastName
	}
	if req.Email != nil {
		d.Email = *req.Email
	}
	if req.Phone != nil {
		d.Phone = *req.Phone
	}

	if err := s.fleetService.UpdateDriver(ctx, d); err != nil {
		return nil, err
	}

	return &pb.UpdateDriverResponse{Driver: driverModelToProto(d)}, nil
}

func (s *FleetServer) DeleteDriver(ctx context.Context, req *pb.DeleteDriverRequest) (*pb.DeleteDriverResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	if err := s.fleetService.DeleteDriver(ctx, id); err != nil {
		return nil, err
	}

	return &pb.DeleteDriverResponse{}, nil
}

func (s *FleetServer) UpdateDriverStatus(ctx context.Context, req *pb.UpdateDriverStatusRequest) (*pb.UpdateDriverStatusResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	if err := s.fleetService.UpdateDriverStatus(ctx, id, req.Status); err != nil {
		return nil, err
	}

	d, err := s.fleetService.GetDriver(ctx, id)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateDriverStatusResponse{Driver: driverModelToProto(d)}, nil
}

// Maintenance handlers

func (s *FleetServer) CreateWorkOrder(ctx context.Context, req *pb.CreateWorkOrderRequest) (*pb.CreateWorkOrderResponse, error) {
	orgID, err := uuid.Parse(req.OrgId)
	if err != nil {
		return nil, err
	}
	vID, err := uuid.Parse(req.VehicleId)
	if err != nil {
		return nil, err
	}

	var assignedTo *uuid.UUID
	if req.AssignedTo != nil {
		id, err := uuid.Parse(*req.AssignedTo)
		if err == nil {
			assignedTo = &id
		}
	}

	var scheduledDate *time.Time
	if req.ScheduledDate != nil {
		t := req.ScheduledDate.AsTime()
		scheduledDate = &t
	}

	var vendorID *uuid.UUID
	if req.VendorId != nil {
		id, err := uuid.Parse(*req.VendorId)
		if err == nil {
			vendorID = &id
		}
	}

	wo := &models.WorkOrder{
		OrgID:         orgID,
		VehicleID:     vID,
		Title:         req.Title,
		Description:   req.Description,
		Type:          req.Type,
		Priority:      req.Priority.String(),
		AssignedTo:    assignedTo,
		ScheduledDate: scheduledDate,
		EstimatedCost: req.EstimatedCost,
		VendorID:      vendorID,
	}

	if err := s.fleetService.CreateWorkOrder(ctx, wo); err != nil {
		return nil, err
	}

	return &pb.CreateWorkOrderResponse{WorkOrder: workOrderToProto(wo)}, nil
}

func (s *FleetServer) GetWorkOrder(ctx context.Context, req *pb.GetWorkOrderRequest) (*pb.GetWorkOrderResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	wo, err := s.fleetService.GetWorkOrder(ctx, id)
	if err != nil {
		return nil, err
	}

	return &pb.GetWorkOrderResponse{WorkOrder: workOrderToProto(wo)}, nil
}

func (s *FleetServer) ListWorkOrders(ctx context.Context, req *pb.ListWorkOrdersRequest) (*pb.ListWorkOrdersResponse, error) {
	orgID, err := uuid.Parse(req.OrgId)
	if err != nil {
		return nil, err
	}

	var vehicleID *uuid.UUID
	if req.VehicleId != nil {
		id, err := uuid.Parse(*req.VehicleId)
		if err == nil {
			vehicleID = &id
		}
	}

	status := ""
	if req.Status != nil {
		status = req.Status.String()
	}

	offset := (req.Page - 1) * req.Limit
	if offset < 0 {
		offset = 0
	}

	wos, total, err := s.fleetService.ListWorkOrders(ctx, orgID, req.Limit, offset, vehicleID, status)
	if err != nil {
		return nil, err
	}

	var protoWos []*pb.WorkOrder
	for i := range wos {
		protoWos = append(protoWos, workOrderToProto(&wos[i]))
	}

	return &pb.ListWorkOrdersResponse{WorkOrders: protoWos, Total: total}, nil
}

func (s *FleetServer) UpdateWorkOrder(ctx context.Context, req *pb.UpdateWorkOrderRequest) (*pb.UpdateWorkOrderResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	wo, err := s.fleetService.GetWorkOrder(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		wo.Title = *req.Title
	}
	if req.Description != nil {
		wo.Description = *req.Description
	}
	if req.Priority != nil {
		wo.Priority = req.Priority.String()
	}
	if req.AssignedTo != nil {
		aid, err := uuid.Parse(*req.AssignedTo)
		if err == nil {
			wo.AssignedTo = &aid
		}
	}
	if req.EstimatedCost != nil {
		wo.EstimatedCost = *req.EstimatedCost
	}
	if req.ActualCost != nil {
		wo.ActualCost = *req.ActualCost
	}

	if err := s.fleetService.UpdateWorkOrder(ctx, wo); err != nil {
		return nil, err
	}

	return &pb.UpdateWorkOrderResponse{WorkOrder: workOrderToProto(wo)}, nil
}

func (s *FleetServer) ChangeWorkOrderStatus(ctx context.Context, req *pb.ChangeWorkOrderStatusRequest) (*pb.ChangeWorkOrderStatusResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	if err := s.fleetService.ChangeWorkOrderStatus(ctx, id, req.Status.String()); err != nil {
		return nil, err
	}

	wo, err := s.fleetService.GetWorkOrder(ctx, id)
	if err != nil {
		return nil, err
	}

	return &pb.ChangeWorkOrderStatusResponse{WorkOrder: workOrderToProto(wo)}, nil
}

func (s *FleetServer) CreateInspection(ctx context.Context, req *pb.CreateInspectionRequest) (*pb.CreateInspectionResponse, error) {
	orgID, err := uuid.Parse(req.OrgId)
	if err != nil {
		return nil, err
	}
	vID, err := uuid.Parse(req.VehicleId)
	if err != nil {
		return nil, err
	}
	iID, err := uuid.Parse(req.InspectorId)
	if err != nil {
		return nil, err
	}

	i := &models.Inspection{
		OrgID:       orgID,
		VehicleID:   vID,
		InspectorID: iID,
		Odometer:    req.Odometer,
		Notes:       req.Notes,
		Status:      "completed",
	}

	if err := s.fleetService.CreateInspection(ctx, i); err != nil {
		return nil, err
	}

	return &pb.CreateInspectionResponse{Inspection: inspectionToProto(i)}, nil
}

func (s *FleetServer) ListInspections(ctx context.Context, req *pb.ListInspectionsRequest) (*pb.ListInspectionsResponse, error) {
	orgID, err := uuid.Parse(req.OrgId)
	if err != nil {
		return nil, err
	}

	var vehicleID *uuid.UUID
	if req.VehicleId != nil {
		id, err := uuid.Parse(*req.VehicleId)
		if err == nil {
			vehicleID = &id
		}
	}

	offset := (req.Page - 1) * req.Limit
	if offset < 0 {
		offset = 0
	}

	inspections, total, err := s.fleetService.ListInspections(ctx, orgID, req.Limit, offset, vehicleID)
	if err != nil {
		return nil, err
	}

	var protoInspections []*pb.Inspection
	for i := range inspections {
		protoInspections = append(protoInspections, inspectionToProto(&inspections[i]))
	}

	return &pb.ListInspectionsResponse{Inspections: protoInspections, Total: total}, nil
}

// Fuel handlers

func (s *FleetServer) CreateFuelLog(ctx context.Context, req *pb.CreateFuelLogRequest) (*pb.CreateFuelLogResponse, error) {
	orgID, err := uuid.Parse(req.OrgId)
	if err != nil {
		return nil, err
	}
	vID, err := uuid.Parse(req.VehicleId)
	if err != nil {
		return nil, err
	}
	dID, err := uuid.Parse(req.DriverId)
	if err != nil {
		return nil, err
	}

	fl := &models.FuelLog{
		OrgID:           orgID,
		VehicleID:       vID,
		DriverID:        dID,
		StationName:     req.StationName,
		FuelType:        req.FuelType,
		Quantity:        req.Quantity,
		UnitPrice:       req.UnitPrice,
		TotalCost:       req.TotalCost,
		Odometer:        req.Odometer,
		TransactionDate: req.TransactionDate.AsTime(),
		ReceiptURL:      req.ReceiptUrl,
		Notes:           req.Notes,
	}

	if err := s.fleetService.CreateFuelLog(ctx, fl); err != nil {
		return nil, err
	}

	return &pb.CreateFuelLogResponse{FuelLog: fuelLogToProto(fl)}, nil
}

func (s *FleetServer) ListFuelLogs(ctx context.Context, req *pb.ListFuelLogsRequest) (*pb.ListFuelLogsResponse, error) {
	orgID, err := uuid.Parse(req.OrgId)
	if err != nil {
		return nil, err
	}

	var vehicleID *uuid.UUID
	if req.VehicleId != nil {
		id, err := uuid.Parse(*req.VehicleId)
		if err == nil {
			vehicleID = &id
		}
	}

	var driverID *uuid.UUID
	if req.DriverId != nil {
		id, err := uuid.Parse(*req.DriverId)
		if err == nil {
			driverID = &id
		}
	}

	offset := (req.Page - 1) * req.Limit
	if offset < 0 {
		offset = 0
	}

	logs, total, err := s.fleetService.ListFuelLogs(ctx, orgID, req.Limit, offset, vehicleID, driverID)
	if err != nil {
		return nil, err
	}

	var protoLogs []*pb.FuelLog
	for i := range logs {
		protoLogs = append(protoLogs, fuelLogToProto(&logs[i]))
	}

	return &pb.ListFuelLogsResponse{Logs: protoLogs, Total: total}, nil
}

func (s *FleetServer) GetFuelStats(ctx context.Context, req *pb.GetFuelStatsRequest) (*pb.GetFuelStatsResponse, error) {
	orgID, err := uuid.Parse(req.OrgId)
	if err != nil {
		return nil, err
	}

	var vehicleID *uuid.UUID
	if req.VehicleId != nil {
		id, err := uuid.Parse(*req.VehicleId)
		if err == nil {
			vehicleID = &id
		}
	}

	totalQty, totalCost, avgConsumption, avgCostPerKm, err := s.fleetService.GetFuelStats(ctx, orgID, vehicleID)
	if err != nil {
		return nil, err
	}

	return &pb.GetFuelStatsResponse{
		TotalConsumption: totalQty,
		TotalCost:        totalCost,
		AvgConsumption:   avgConsumption,
		AvgCostPerKm:     avgCostPerKm,
	}, nil
}

// Telemetry handlers

func (s *FleetServer) ReportTelemetry(ctx context.Context, req *pb.ReportTelemetryRequest) (*pb.ReportTelemetryResponse, error) {
	t := protoToTelemetry(req.Telemetry)
	if err := s.fleetService.ReportTelemetry(ctx, t); err != nil {
		return nil, err
	}
	return &pb.ReportTelemetryResponse{Success: true}, nil
}

func (s *FleetServer) GetLatestTelemetry(ctx context.Context, req *pb.GetLatestTelemetryRequest) (*pb.GetLatestTelemetryResponse, error) {
	vID, err := uuid.Parse(req.VehicleId)
	if err != nil {
		return nil, err
	}

	t, err := s.fleetService.GetLatestTelemetry(ctx, vID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return &pb.GetLatestTelemetryResponse{}, nil
	}

	return &pb.GetLatestTelemetryResponse{Telemetry: telemetryToProto(t)}, nil
}

func vehicleModelToProto(v *models.Vehicle) *pb.Vehicle {
	return &pb.Vehicle{
		Id:           v.ID.String(),
		OrgId:        v.OrgID.String(),
		LicensePlate: v.LicensePlate,
		Model:        v.Model,
		Make:         v.Make,
		Year:         v.Year,
		Status:       v.Status,
		Vin:          v.VIN,
		CreatedAt:    timestamppb.New(v.CreatedAt),
		UpdatedAt:    timestamppb.New(v.UpdatedAt),
	}
}

func driverModelToProto(d *models.Driver) *pb.Driver {
	var currentVehicleID *string
	if d.CurrentVehicleID != nil {
		vid := d.CurrentVehicleID.String()
		currentVehicleID = &vid
	}

	return &pb.Driver{
		Id:               d.ID.String(),
		OrgId:            d.OrgID.String(),
		FirstName:        d.FirstName,
		LastName:         d.LastName,
		Email:            d.Email,
		Phone:            d.Phone,
		LicenseNumber:    d.LicenseNumber,
		LicenseType:      d.LicenseType,
		LicenseExpiry:    timestamppb.New(d.LicenseExpiry),
		Status:           d.Status,
		CurrentVehicleId: currentVehicleID,
		Rating:           d.Rating,
		BehaviorScore:    int32(d.BehaviorScore),
		CreatedAt:        timestamppb.New(d.CreatedAt),
		UpdatedAt:        timestamppb.New(d.UpdatedAt),
	}
}

func inspectionToProto(i *models.Inspection) *pb.Inspection {
	return &pb.Inspection{
		Id:          i.ID.String(),
		OrgId:       i.OrgID.String(),
		VehicleId:   i.VehicleID.String(),
		InspectorId: i.InspectorID.String(),
		Status:      i.Status,
		Notes:       i.Notes,
		Odometer:    i.Odometer,
		CreatedAt:   timestamppb.New(i.CreatedAt),
		UpdatedAt:   timestamppb.New(i.UpdatedAt),
	}
}

func workOrderToProto(wo *models.WorkOrder) *pb.WorkOrder {
	var assignedTo *string
	if wo.AssignedTo != nil {
		at := wo.AssignedTo.String()
		assignedTo = &at
	}

	var scheduledDate *timestamppb.Timestamp
	if wo.ScheduledDate != nil {
		scheduledDate = timestamppb.New(*wo.ScheduledDate)
	}

	var completedDate *timestamppb.Timestamp
	if wo.CompletedDate != nil {
		completedDate = timestamppb.New(*wo.CompletedDate)
	}

	var vendorID *string
	if wo.VendorID != nil {
		vid := wo.VendorID.String()
		vendorID = &vid
	}

	status := pb.WorkOrderStatus_WORK_ORDER_STATUS_UNSPECIFIED
	if s, ok := pb.WorkOrderStatus_value["WORK_ORDER_STATUS_"+wo.Status]; ok {
		status = pb.WorkOrderStatus(s)
	}

	priority := pb.WorkOrderPriority_WORK_ORDER_PRIORITY_UNSPECIFIED
	if p, ok := pb.WorkOrderPriority_value["WORK_ORDER_PRIORITY_"+wo.Priority]; ok {
		priority = pb.WorkOrderPriority(p)
	}

	return &pb.WorkOrder{
		Id:             wo.ID.String(),
		OrgId:          wo.OrgID.String(),
		VehicleId:      wo.VehicleID.String(),
		Title:          wo.Title,
		Description:    wo.Description,
		Type:           wo.Type,
		Priority:       priority,
		Status:         status,
		AssignedTo:     assignedTo,
		ScheduledDate:  scheduledDate,
		CompletedDate:  completedDate,
		EstimatedCost:  wo.EstimatedCost,
		ActualCost:     wo.ActualCost,
		VendorId:       vendorID,
		ApprovalStatus: wo.ApprovalStatus,
		CreatedAt:      timestamppb.New(wo.CreatedAt),
		UpdatedAt:      timestamppb.New(wo.UpdatedAt),
	}
}

func fuelLogToProto(fl *models.FuelLog) *pb.FuelLog {
	return &pb.FuelLog{
		Id:              fl.ID.String(),
		OrgId:           fl.OrgID.String(),
		VehicleId:       fl.VehicleID.String(),
		DriverId:        fl.DriverID.String(),
		StationName:     fl.StationName,
		FuelType:        fl.FuelType,
		Quantity:        fl.Quantity,
		UnitPrice:       fl.UnitPrice,
		TotalCost:       fl.TotalCost,
		Odometer:        fl.Odometer,
		TransactionDate: timestamppb.New(fl.TransactionDate),
		Status:          fl.Status,
		ReceiptUrl:      fl.ReceiptURL,
		Notes:           fl.Notes,
		CreatedAt:       timestamppb.New(fl.CreatedAt),
		UpdatedAt:       timestamppb.New(fl.UpdatedAt),
	}
}

func telemetryToProto(t *models.Telemetry) *pb.Telemetry {
	return &pb.Telemetry{
		VehicleId:      t.VehicleID.String(),
		OrgId:          t.OrgID.String(),
		DeviceId:       t.DeviceID,
		Latitude:       t.Latitude,
		Longitude:      t.Longitude,
		Speed:          t.Speed,
		FuelLevel:      t.FuelLevel,
		BatteryVoltage: t.BatteryVoltage,
		CoolantTemp:    t.CoolantTemp,
		Odometer:       t.Odometer,
		Timestamp:      timestamppb.New(t.Timestamp),
		Metadata:       t.Metadata,
	}
}

func protoToTelemetry(p *pb.Telemetry) *models.Telemetry {
	vID, _ := uuid.Parse(p.VehicleId)
	oID, _ := uuid.Parse(p.OrgId)
	return &models.Telemetry{
		VehicleID:      vID,
		OrgID:          oID,
		DeviceID:       p.DeviceId,
		Latitude:       p.Latitude,
		Longitude:      p.Longitude,
		Speed:          p.Speed,
		FuelLevel:      p.FuelLevel,
		BatteryVoltage: p.BatteryVoltage,
		CoolantTemp:    p.CoolantTemp,
		Odometer:       p.Odometer,
		Timestamp:      p.Timestamp.AsTime(),
		Metadata:       p.Metadata,
	}
}
