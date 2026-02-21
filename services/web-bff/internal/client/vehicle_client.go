package client

import (
	"context"
	"fmt"
	"time"

	"github.com/fleet-platform/shared/pkg/grpcclient"
	pb "github.com/flyfleet/fleet-core-service-new/proto/fleet/v1"
	"github.com/google/uuid"
)

type FleetClient struct {
	client pb.FleetServiceClient
	base   *grpcclient.Client
}

func NewFleetClient(baseURL string) (*FleetClient, error) {
	// e.g. "fleet-core-service:50051"
	baseClient, err := grpcclient.NewClient(baseURL, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to fleet grpc via unified client: %w", err)
	}

	return &FleetClient{
		client: pb.NewFleetServiceClient(baseClient.Conn),
		base:   baseClient,
	}, nil
}

func (c *FleetClient) Close() {
	if c.base != nil {
		c.base.Close()
	}
}

type Vehicle struct {
	ID        uuid.UUID `json:"id"`
	OrgID     uuid.UUID `json:"org_id"`
	VIN       string    `json:"vin"`
	Make      string    `json:"make"`
	Model     string    `json:"model"`
	Year      int       `json:"year"`
	Plate     string    `json:"license_plate"`
	Status    string    `json:"status"`
}

type Driver struct {
	ID            uuid.UUID `json:"id"`
	OrgID         uuid.UUID `json:"org_id"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	Email         string    `json:"email"`
	Phone         string    `json:"phone"`
	LicenseNumber string    `json:"license_number"`
	Status        string    `json:"status"`
}

type VehicleListResponse struct {
	Vehicles []Vehicle `json:"vehicles"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
}

type DriverListResponse struct {
	Drivers []Driver `json:"drivers"`
	Total   int      `json:"total"`
	Page    int      `json:"page"`
	Limit   int      `json:"limit"`
}

type WorkOrder struct {
	ID            uuid.UUID `json:"id"`
	OrgID         uuid.UUID `json:"org_id"`
	VehicleID     uuid.UUID `json:"vehicle_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Type          string    `json:"type"`
	Priority      string    `json:"priority"`
	Status        string    `json:"status"`
	AssignedTo    *string   `json:"assigned_to,omitempty"`
	EstimatedCost float64   `json:"estimated_cost"`
	ActualCost    float64   `json:"actual_cost"`
}

type WorkOrderListResponse struct {
	WorkOrders []WorkOrder `json:"work_orders"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
}

type Inspection struct {
	ID          uuid.UUID `json:"id"`
	OrgID       uuid.UUID `json:"org_id"`
	VehicleID   uuid.UUID `json:"vehicle_id"`
	InspectorID uuid.UUID `json:"inspector_id"`
	Status      string    `json:"status"`
	Notes       string    `json:"notes"`
	Odometer    int64     `json:"odometer"`
}

type InspectionListResponse struct {
	Inspections []Inspection `json:"inspections"`
	Total       int          `json:"total"`
	Page        int          `json:"page"`
	Limit       int          `json:"limit"`
}

type FuelLog struct {
	ID              uuid.UUID `json:"id"`
	OrgID           uuid.UUID `json:"org_id"`
	VehicleID       uuid.UUID `json:"vehicle_id"`
	DriverID        uuid.UUID `json:"driver_id"`
	StationName     string    `json:"station_name"`
	FuelType        string    `json:"fuel_type"`
	Quantity        float64   `json:"quantity"`
	UnitPrice       float64   `json:"unit_price"`
	TotalCost       float64   `json:"total_cost"`
	Odometer        float64   `json:"odometer"`
	TransactionDate string    `json:"transaction_date"`
	Status          string    `json:"status"`
}

type FuelLogListResponse struct {
	Logs  []FuelLog `json:"logs"`
	Total int       `json:"total"`
	Page  int       `json:"page"`
	Limit int       `json:"limit"`
}

type Telemetry struct {
	VehicleID      uuid.UUID         `json:"vehicle_id"`
	OrgID          uuid.UUID         `json:"org_id"`
	DeviceID       string            `json:"device_id"`
	Latitude       float64           `json:"latitude"`
	Longitude      float64           `json:"longitude"`
	Speed          float64           `json:"speed"`
	FuelLevel      float64           `json:"fuel_level"`
	BatteryVoltage float64           `json:"battery_voltage"`
	CoolantTemp    float64           `json:"coolant_temp"`
	Odometer       int64             `json:"odometer"`
	Timestamp      string            `json:"timestamp"`
	Metadata       map[string]string `json:"metadata"`
}

func (c *FleetClient) ListVehicles(ctx context.Context, orgID string, page, limit int) (*VehicleListResponse, error) {
	var resp *pb.ListVehiclesResponse
	err := c.base.Call(ctx, func(ctx context.Context) error {
		var errCall error
		resp, errCall = c.client.ListVehicles(ctx, &pb.ListVehiclesRequest{
			OrgId: orgID,
			Page:  int32(page),
			Limit: int32(limit),
		})
		return errCall
	})
	if err != nil {
		return nil, err
	}

	var vehicles []Vehicle
	for _, v := range resp.Vehicles {
		vID, _ := uuid.Parse(v.Id)
		oID, _ := uuid.Parse(v.OrgId)
		vehicles = append(vehicles, Vehicle{
			ID:     vID,
			OrgID:  oID,
			VIN:    v.Vin,
			Make:   v.Make,
			Model:  v.Model,
			Year:   int(v.Year),
			Plate:  v.LicensePlate,
			Status: v.Status,
		})
	}

	return &VehicleListResponse{
		Vehicles: vehicles,
		Total:    int(resp.Total),
		Page:     page,
		Limit:    limit,
	}, nil
}

func (c *FleetClient) ListDrivers(ctx context.Context, orgID string, page, limit int) (*DriverListResponse, error) {
	var resp *pb.ListDriversResponse
	err := c.base.Call(ctx, func(ctx context.Context) error {
		var errCall error
		resp, errCall = c.client.ListDrivers(ctx, &pb.ListDriversRequest{
			OrgId: orgID,
			Page:  int32(page),
			Limit: int32(limit),
		})
		return errCall
	})
	if err != nil {
		return nil, err
	}

	var drivers []Driver
	for _, d := range resp.Drivers {
		dID, _ := uuid.Parse(d.Id)
		oID, _ := uuid.Parse(d.OrgId)
		drivers = append(drivers, Driver{
			ID:            dID,
			OrgID:         oID,
			FirstName:     d.FirstName,
			LastName:      d.LastName,
			Email:         d.Email,
			Phone:         d.Phone,
			LicenseNumber: d.LicenseNumber,
			Status:        d.Status,
		})
	}

	return &DriverListResponse{
		Drivers: drivers,
		Total:   int(resp.Total),
		Page:    page,
		Limit:   limit,
	}, nil
}

func (c *FleetClient) ListWorkOrders(ctx context.Context, orgID string, page, limit int) (*WorkOrderListResponse, error) {
	var resp *pb.ListWorkOrdersResponse
	err := c.base.Call(ctx, func(ctx context.Context) error {
		var errCall error
		resp, errCall = c.client.ListWorkOrders(ctx, &pb.ListWorkOrdersRequest{
			OrgId: orgID,
			Page:  int32(page),
			Limit: int32(limit),
		})
		return errCall
	})
	if err != nil {
		return nil, err
	}

	var wos []WorkOrder
	for _, w := range resp.WorkOrders {
		wID, _ := uuid.Parse(w.Id)
		oID, _ := uuid.Parse(w.OrgId)
		vID, _ := uuid.Parse(w.VehicleId)
		wos = append(wos, WorkOrder{
			ID:            wID,
			OrgID:         oID,
			VehicleID:     vID,
			Title:         w.Title,
			Description:   w.Description,
			Type:          w.Type,
			Priority:      w.Priority.String(),
			Status:        w.Status.String(),
			AssignedTo:    w.AssignedTo,
			EstimatedCost: w.EstimatedCost,
			ActualCost:    w.ActualCost,
		})
	}

	return &WorkOrderListResponse{
		WorkOrders: wos,
		Total:      int(resp.Total),
		Page:       page,
		Limit:      limit,
	}, nil
}

func (c *FleetClient) ListInspections(ctx context.Context, orgID string, page, limit int) (*InspectionListResponse, error) {
	var resp *pb.ListInspectionsResponse
	err := c.base.Call(ctx, func(ctx context.Context) error {
		var errCall error
		resp, errCall = c.client.ListInspections(ctx, &pb.ListInspectionsRequest{
			OrgId: orgID,
			Page:  int32(page),
			Limit: int32(limit),
		})
		return errCall
	})
	if err != nil {
		return nil, err
	}

	var inspections []Inspection
	for _, i := range resp.Inspections {
		iID, _ := uuid.Parse(i.Id)
		oID, _ := uuid.Parse(i.OrgId)
		vID, _ := uuid.Parse(i.VehicleId)
		insID, _ := uuid.Parse(i.InspectorId)
		inspections = append(inspections, Inspection{
			ID:          iID,
			OrgID:       oID,
			VehicleID:   vID,
			InspectorID: insID,
			Status:      i.Status,
			Notes:       i.Notes,
			Odometer:    i.Odometer,
		})
	}

	return &InspectionListResponse{
		Inspections: inspections,
		Total:       int(resp.Total),
		Page:        page,
		Limit:       limit,
	}, nil
}

func (c *FleetClient) ListFuelLogs(ctx context.Context, orgID string, page, limit int) (*FuelLogListResponse, error) {
	var resp *pb.ListFuelLogsResponse
	err := c.base.Call(ctx, func(ctx context.Context) error {
		var errCall error
		resp, errCall = c.client.ListFuelLogs(ctx, &pb.ListFuelLogsRequest{
			OrgId: orgID,
			Page:  int32(page),
			Limit: int32(limit),
		})
		return errCall
	})
	if err != nil {
		return nil, err
	}

	var logs []FuelLog
	for _, fl := range resp.Logs {
		id, _ := uuid.Parse(fl.Id)
		oid, _ := uuid.Parse(fl.OrgId)
		vid, _ := uuid.Parse(fl.VehicleId)
		did, _ := uuid.Parse(fl.DriverId)
		logs = append(logs, FuelLog{
			ID:              id,
			OrgID:           oid,
			VehicleID:       vid,
			DriverID:        did,
			StationName:     fl.StationName,
			FuelType:        fl.FuelType,
			Quantity:        fl.Quantity,
			UnitPrice:       fl.UnitPrice,
			TotalCost:       fl.TotalCost,
			Odometer:        fl.Odometer,
			TransactionDate: fl.TransactionDate.AsTime().String(),
			Status:          fl.Status,
		})
	}

	return &FuelLogListResponse{
		Logs:  logs,
		Total: int(resp.Total),
		Page:  page,
		Limit: limit,
	}, nil
}

func (c *FleetClient) GetFuelStats(ctx context.Context, orgID string, vehicleID *string) (map[string]interface{}, error) {
	var resp *pb.GetFuelStatsResponse
	err := c.base.Call(ctx, func(ctx context.Context) error {
		var errCall error
		resp, errCall = c.client.GetFuelStats(ctx, &pb.GetFuelStatsRequest{
			OrgId:     orgID,
			VehicleId: vehicleID,
		})
		return errCall
	})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_consumption": resp.TotalConsumption,
		"total_cost":        resp.TotalCost,
		"avg_consumption":   resp.AvgConsumption,
		"avg_cost_per_km":   resp.AvgCostPerKm,
	}, nil
}

func (c *FleetClient) GetLatestTelemetry(ctx context.Context, vehicleID string) (*Telemetry, error) {
	var resp *pb.GetLatestTelemetryResponse
	err := c.base.Call(ctx, func(ctx context.Context) error {
		var errCall error
		resp, errCall = c.client.GetLatestTelemetry(ctx, &pb.GetLatestTelemetryRequest{
			VehicleId: vehicleID,
		})
		return errCall
	})
	if err != nil {
		return nil, err
	}

	if resp.Telemetry == nil {
		return nil, nil
	}

	t := resp.Telemetry
	vID, _ := uuid.Parse(t.VehicleId)
	oID, _ := uuid.Parse(t.OrgId)

	return &Telemetry{
		VehicleID:      vID,
		OrgID:          oID,
		DeviceID:       t.DeviceId,
		Latitude:       t.Latitude,
		Longitude:      t.Longitude,
		Speed:          t.Speed,
		FuelLevel:      t.FuelLevel,
		BatteryVoltage: t.BatteryVoltage,
		CoolantTemp:    t.CoolantTemp,
		Odometer:       t.Odometer,
		Timestamp:      t.Timestamp.AsTime().String(),
		Metadata:       t.Metadata,
	}, nil
}

type VehicleStats struct {
	TotalVehicles  int     `json:"total_vehicles"`
	ActiveVehicles int     `json:"active_vehicles"`
	TotalDistance  float64 `json:"total_distance"`
	TotalFuel      float64 `json:"total_fuel"`
}

func (c *FleetClient) GetStats(ctx context.Context, token string) (*VehicleStats, error) {
	return &VehicleStats{
		TotalVehicles:  0,
		ActiveVehicles: 0,
	}, nil
}
