package connectors

import "context"

type ERPType string

const (
	SAP        ERPType = "SAP"
	Salesforce ERPType = "SALESFORCE"
	Oracle     ERPType = "ORACLE"
	Dynamics   ERPType = "DYNAMICS"
)

type SyncResult struct {
	Success bool
	Message string
	Data    interface{}
}

type ERPConnector interface {
	PushOrder(ctx context.Context, orderID string, data interface{}) (*SyncResult, error)
	PullCustomer(ctx context.Context, externalID string) (*SyncResult, error)
	GetStatus(ctx context.Context) (string, error)
}
