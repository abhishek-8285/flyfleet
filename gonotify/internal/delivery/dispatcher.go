package delivery

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gonotify/gonotify/internal/provider"
	"github.com/gonotify/gonotify/internal/queue"
	"github.com/gonotify/gonotify/internal/template"
)

type Dispatcher struct {
	db        *pgxpool.Pool
	email     provider.Provider
	templates *template.Repository
	logger    *slog.Logger
}

func NewDispatcher(
	db *pgxpool.Pool,
	email provider.Provider,
	templates *template.Repository,
	logger *slog.Logger,
) *Dispatcher {
	return &Dispatcher{
		db:        db,
		email:     email,
		templates: templates,
		logger:    logger,
	}
}

func (d *Dispatcher) HandleEvent(ctx context.Context, msg queue.Message) error {
	var data map[string]any

	if len(msg.Payload) > 0 {
		if err := json.Unmarshal(msg.Payload, &data); err != nil {
			return fmt.Errorf("unmarshal event payload: %w", err)
		}
	}

	if data == nil {
		data = map[string]any{}
	}

	eventType := msg.Metadata["event_type"]
	userID := msg.Metadata["user_id"]

	data["event"] = eventType
	data["user_id"] = userID

	emailTo, _ := data["email"].(string)

	if emailTo == "" {
		d.logger.WarnContext(
			ctx,
			"event missing data.email; skipping email delivery",
			"event_id", msg.ID,
			"tenant_id", msg.TenantID,
			"event_type", eventType,
		)

		return nil
	}

	subjectTemplate, htmlTemplate, textTemplate, templateID := d.resolveTemplates(
		ctx,
		msg,
		eventType,
	)

	rendered, err := template.Render(template.RenderInput{
		Subject: subjectTemplate,
		HTML:    htmlTemplate,
		Text:    textTemplate,
		Data:    data,
	})
	if err != nil {
		return fmt.Errorf("template render failed: %w", err)
	}

	idempotencyKey := notificationKey(msg.ID, "email", emailTo)

	notificationID, status, err := d.createOrGetNotification(
		ctx,
		msg,
		emailTo,
		idempotencyKey,
		templateID,
		rendered,
		data,
	)
	if err != nil {
		return fmt.Errorf("create or get notification failed: %w", err)
	}

	if isTerminalSuccessStatus(status) {
		d.logger.InfoContext(
			ctx,
			"notification already delivered; skipping duplicate delivery",
			"notification_id", notificationID,
			"event_id", msg.ID,
			"tenant_id", msg.TenantID,
			"status", status,
		)

		return nil
	}

	attempt, err := d.nextAttempt(ctx, notificationID)
	if err != nil {
		d.logger.WarnContext(
			ctx,
			"failed to calculate delivery attempt; defaulting to attempt 1",
			"notification_id", notificationID,
			"error", err,
		)

		attempt = 1
	}

	if err := d.updateNotificationStatus(
		ctx,
		notificationID,
		"sending",
		"",
	); err != nil {
		d.logger.ErrorContext(
			ctx,
			"failed to update notification status to sending",
			"notification_id", notificationID,
			"error", err,
		)
	}

	req := provider.Request{
		TenantID:       msg.TenantID,
		NotificationID: notificationID,
		Channel:        provider.ChannelEmail,
		To: provider.Recipient{
			UserID: userID,
			Email:  emailTo,
		},
		Subject:  rendered.Subject,
		Text:     rendered.Text,
		HTML:     rendered.HTML,
		Data:     data,
		Metadata: msg.Metadata,
	}

	result, sendErr := d.email.Send(ctx, req)

	attemptStatus := "success"
	errText := ""

	if sendErr != nil {
		attemptStatus = "failed"
		errText = sendErr.Error()
	}

	if err := d.recordDeliveryAttempt(
		ctx,
		notificationID,
		attemptStatus,
		attempt,
		req,
		result,
		errText,
	); err != nil {
		d.logger.ErrorContext(
			ctx,
			"failed to record delivery attempt",
			"notification_id", notificationID,
			"error", err,
		)
	}

	if sendErr != nil {
		if err := d.updateNotificationStatus(
			ctx,
			notificationID,
			"failed",
			sendErr.Error(),
		); err != nil {
			d.logger.ErrorContext(
				ctx,
				"failed to update notification status to failed",
				"notification_id", notificationID,
				"error", err,
			)
		}

		return fmt.Errorf("email send failed: %w", sendErr)
	}

	if err := d.updateNotificationStatus(
		ctx,
		notificationID,
		"sent",
		"",
	); err != nil {
		d.logger.ErrorContext(
			ctx,
			"failed to update notification status to sent",
			"notification_id", notificationID,
			"error", err,
		)
	}

	d.logger.InfoContext(
		ctx,
		"notification sent",
		"notification_id", notificationID,
		"event_id", msg.ID,
		"tenant_id", msg.TenantID,
		"channel", "email",
		"to", emailTo,
		"attempt", attempt,
	)

	return nil
}

func (d *Dispatcher) resolveTemplates(
	ctx context.Context,
	msg queue.Message,
	eventType string,
) (string, string, string, *string) {
	locale := "en"

	resolved, err := d.templates.GetPublishedVersion(
		ctx,
		msg.TenantID,
		eventType,
		"email",
		locale,
	)

	if err == nil {
		if resolved.Subject != "" || resolved.HTML != "" || resolved.Text != "" {
			d.logger.DebugContext(
				ctx,
				"resolved database template",
				"tenant_id", msg.TenantID,
				"event_type", eventType,
				"template_id", resolved.TemplateID,
				"version_id", resolved.VersionID,
			)

			templateID := resolved.TemplateID

			return resolved.Subject, resolved.HTML, resolved.Text, &templateID
		}
	}

	if err != nil && !errors.Is(err, template.ErrTemplateNotFound) {
		d.logger.ErrorContext(
			ctx,
			"template resolution failed; falling back to default template",
			"tenant_id", msg.TenantID,
			"event_type", eventType,
			"error", err,
		)
	}

	subject, html, text := templatesForEvent(eventType)

	return subject, html, text, nil
}

func (d *Dispatcher) createOrGetNotification(
	ctx context.Context,
	msg queue.Message,
	toAddress string,
	idempotencyKey string,
	templateID *string,
	rendered template.RenderOutput,
	data map[string]any,
) (string, string, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return "", "", err
	}

	var notificationID string
	var status string

	err = d.db.QueryRow(
		ctx,
		`
		INSERT INTO notifications (
			tenant_id,
			event_id,
			template_id,
			channel,
			category,
			status,
			to_address,
			subject,
			body,
			payload,
			idempotency_key
		)
		VALUES (
			$1::uuid,
			$2::uuid,
			$3::uuid,
			'email',
			'transactional',
			'queued',
			$4,
			$5,
			$6,
			$7,
			$8
		)
		ON CONFLICT (tenant_id, idempotency_key) DO NOTHING
		RETURNING id::text, status
		`,
		msg.TenantID,
		msg.ID,
		templateID,
		toAddress,
		rendered.Subject,
		rendered.Text,
		payload,
		idempotencyKey,
	).Scan(&notificationID, &status)

	if err == nil {
		return notificationID, status, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return "", "", err
	}

	err = d.db.QueryRow(
		ctx,
		`
		SELECT
			id::text,
			status
		FROM notifications
		WHERE tenant_id = $1::uuid
		  AND idempotency_key = $2
		`,
		msg.TenantID,
		idempotencyKey,
	).Scan(&notificationID, &status)

	if err != nil {
		return "", "", err
	}

	return notificationID, status, nil
}

func (d *Dispatcher) recordDeliveryAttempt(
	ctx context.Context,
	notificationID string,
	status string,
	attempt int,
	req provider.Request,
	result provider.Result,
	errText string,
) error {
	requestJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	responseJSON, err := json.Marshal(result)
	if err != nil {
		return err
	}

	_, err = d.db.Exec(
		ctx,
		`
		INSERT INTO delivery_attempts (
			notification_id,
			provider_config_id,
			status,
			attempt,
			request,
			response,
			error
		)
		VALUES (
			$1::uuid,
			NULL,
			$2,
			$3,
			$4,
			$5,
			NULLIF($6, '')
		)
		`,
		notificationID,
		status,
		attempt,
		requestJSON,
		responseJSON,
		errText,
	)

	return err
}

func (d *Dispatcher) updateNotificationStatus(
	ctx context.Context,
	notificationID string,
	status string,
	errText string,
) error {
	_, err := d.db.Exec(
		ctx,
		`
		UPDATE notifications
		SET
			status = $2,
			error = NULLIF($3, ''),
			sent_at = CASE
				WHEN $2 = 'sent' THEN now()
				ELSE sent_at
			END,
			failed_at = CASE
				WHEN $2 = 'failed' THEN now()
				ELSE failed_at
			END
		WHERE id = $1::uuid
		`,
		notificationID,
		status,
		errText,
	)

	return err
}

func notificationKey(
	eventID string,
	channel string,
	toAddress string,
) string {
	normalized := strings.ToLower(strings.TrimSpace(toAddress))

	raw := fmt.Sprintf(
		"%s:%s:%s",
		eventID,
		channel,
		normalized,
	)

	sum := sha256.Sum256([]byte(raw))

	return hex.EncodeToString(sum[:])
}

func isTerminalSuccessStatus(status string) bool {
	switch status {
	case "sent", "delivered", "opened", "clicked":
		return true
	default:
		return false
	}
}

func (d *Dispatcher) nextAttempt(
	ctx context.Context,
	notificationID string,
) (int, error) {
	var attempt int

	err := d.db.QueryRow(
		ctx,
		`
		SELECT COALESCE(MAX(attempt), 0) + 1
		FROM delivery_attempts
		WHERE notification_id = $1::uuid
		`,
		notificationID,
	).Scan(&attempt)

	if err != nil {
		return 1, err
	}

	return attempt, nil
}

func templatesForEvent(eventType string) (string, string, string) {
	switch eventType {
	case "order.created":
		subject := `Your order {{.order_id}} has been created`

		html := `
<p>Hello,</p>

<p>
  Your order <strong>{{.order_id}}</strong> has been created.
</p>

<p>
  Amount: {{.amount}} {{.currency}}
</p>
`

		text := `
Hello,

Your order {{.order_id}} has been created.

Amount: {{.amount}} {{.currency}}
`

		return subject, html, text

	default:
		subject := `Notification: {{.event}}`

		html := `
<p>
  Event <code>{{.event}}</code> was received.
</p>
`

		text := `
Event {{.event}} was received.
`

		return subject, html, text
	}
}
