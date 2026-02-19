package async

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// Task Types
const (
	TypeWebhookDeliver     = "webhook:deliver"
	TypeReportGenerate     = "report:generate"
	TypeMaintenanceReminder = "maintenance:reminder"
	TypeUsageAggregation   = "billing:usage_aggregation"
	TypeMonthlyInvoice     = "billing:monthly_invoice"
	TypeAlertEscalation    = "alert:escalation"
	TypeSLAAggregation     = "monitoring:sla_aggregation"
)

// WebhookPayload defines the payload for webhook delivery tasks
type WebhookPayload struct {
	WebhookID uuid.UUID              `json:"webhook_id"`
	OrgID     uuid.UUID              `json:"org_id"`
	Event     string                 `json:"event"`
	Payload   map[string]interface{} `json:"payload"`
}

// ReportGeneratePayload defines the payload for report generation tasks
type ReportGeneratePayload struct {
	ExecutionID uuid.UUID `json:"execution_id"`
	ReportID    uuid.UUID `json:"report_id"`
	OrgID       uuid.UUID `json:"org_id"`
}

// MaintenanceReminderPayload defines the payload for maintenance reminder tasks
type MaintenanceReminderPayload struct {
	OrgID      uuid.UUID `json:"org_id"`
	ScheduleID uuid.UUID `json:"schedule_id"`
}

// UsageAggregationPayload defines the payload for aggregating consumption events
type UsageAggregationPayload struct {
	OrgID     uuid.UUID `json:"org_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// MonthlyInvoicePayload defines the payload for generating monthly invoices
type MonthlyInvoicePayload struct {
	OrgID uuid.UUID `json:"org_id"`
	Month time.Month `json:"month"`
	Year  int        `json:"year"`
}

// AlertEscalationPayload defines the payload for alert escalation tasks
type AlertEscalationPayload struct {
	OrgID   uuid.UUID `json:"org_id"`
	AlertID string    `json:"alert_id"`
}

// SLAAggregationPayload defines the payload for SLA metric aggregation
type SLAAggregationPayload struct {
	OrgID     uuid.UUID `json:"org_id"`
	Metric    string    `json:"metric"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// QueueClient provides a standardized way to enqueue asynchronous tasks
type QueueClient struct {
	client *asynq.Client
}

func NewQueueClient(redisAddr string) *QueueClient {
	return &QueueClient{
		client: asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr}),
	}
}

func (qc *QueueClient) EnqueueWebhook(ctx context.Context, payload WebhookPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	task := asynq.NewTask(TypeWebhookDeliver, data)
	_, err = qc.client.EnqueueContext(ctx, task, asynq.MaxRetry(5), asynq.Timeout(1*time.Minute))
	return err
}

func (qc *QueueClient) EnqueueReport(ctx context.Context, payload ReportGeneratePayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	task := asynq.NewTask(TypeReportGenerate, data)
	_, err = qc.client.EnqueueContext(ctx, task, asynq.MaxRetry(3), asynq.Timeout(30*time.Minute))
	return err
}

func (qc *QueueClient) Close() error {
	return qc.client.Close()
}
