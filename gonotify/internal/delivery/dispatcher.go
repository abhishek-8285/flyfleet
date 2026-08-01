package delivery

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gonotify/gonotify/internal/inbox"
	"github.com/gonotify/gonotify/internal/provider"
	"github.com/gonotify/gonotify/internal/queue"
	"github.com/gonotify/gonotify/internal/realtime"
	"github.com/gonotify/gonotify/internal/template"
	"github.com/gonotify/gonotify/internal/user"
)

var e164Regex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

type Dispatcher struct {
	db        *pgxpool.Pool
	email     provider.Provider
	sms       provider.Provider
	templates *template.Repository
	users     *user.Repository
	inbox     *inbox.Repository
	realtime  *realtime.Publisher
	logger    *slog.Logger
}

func NewDispatcher(
	db *pgxpool.Pool,
	email provider.Provider,
	sms provider.Provider,
	templates *template.Repository,
	users *user.Repository,
	inboxRepo *inbox.Repository,
	realtimePublisher *realtime.Publisher,
	logger *slog.Logger,
) *Dispatcher {
	return &Dispatcher{
		db:        db,
		email:     email,
		sms:       sms,
		templates: templates,
		users:     users,
		inbox:     inboxRepo,
		realtime:  realtimePublisher,
		logger:    logger,
	}
}

type createNotificationParams struct {
	Msg             queue.Message
	Channel         string
	ToAddress       *string
	RecipientUserID *string
	TemplateID      *string
	Subject         string
	Body            string
	Payload         map[string]any
	IdempotencyKey  string
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

	channels := parseChannels(msg.Metadata["channels"])

	var errs []error

	for _, channel := range channels {
		switch channel {
		case "email":
			if err := d.handleEmail(ctx, msg, data, eventType, userID); err != nil {
				errs = append(errs, err)
			}

		case "in_app":
			if err := d.handleInApp(ctx, msg, data, eventType, userID); err != nil {
				errs = append(errs, err)
			}

		case "sms":
			if d.sms == nil {
				d.logger.DebugContext(
					ctx,
					"sms channel requested but no sms provider configured; skipping",
					"event_id", msg.ID,
					"tenant_id", msg.TenantID,
				)
				continue
			}

			if err := d.handleSMS(ctx, msg, data, eventType, userID); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (d *Dispatcher) handleEmail(
	ctx context.Context,
	msg queue.Message,
	data map[string]any,
	eventType string,
	userID string,
) error {
	emailTo, _ := data["email"].(string)

	if emailTo == "" {
		d.logger.DebugContext(
			ctx,
			"event missing data.email; skipping email delivery",
			"event_id", msg.ID,
			"tenant_id", msg.TenantID,
			"event_type", eventType,
		)

		return nil
	}

	subjectTemplate, htmlTemplate, textTemplate, templateID := d.resolveEmailTemplates(
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
		return fmt.Errorf("email template render failed: %w", err)
	}

	idempotencyKey := notificationKey(msg.ID, "email", emailTo)

	notificationID, status, err := d.createOrGetNotification(
		ctx,
		createNotificationParams{
			Msg:             msg,
			Channel:         "email",
			ToAddress:       &emailTo,
			RecipientUserID: nil,
			TemplateID:      templateID,
			Subject:         rendered.Subject,
			Body:            rendered.Text,
			Payload:         data,
			IdempotencyKey:  idempotencyKey,
		},
	)
	if err != nil {
		return fmt.Errorf("create or get email notification failed: %w", err)
	}

	if isTerminalSuccessStatus(status) {
		d.logger.InfoContext(
			ctx,
			"email notification already delivered; skipping duplicate delivery",
			"notification_id", notificationID,
			"event_id", msg.ID,
			"tenant_id", msg.TenantID,
			"status", status,
		)

		return nil
	}

	attempt, err := d.nextAttempt(ctx, notificationID)
	if err != nil {
		attempt = 1
	}

	if err := d.updateNotificationStatus(ctx, notificationID, "sending", ""); err != nil {
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
			"failed to record email delivery attempt",
			"notification_id", notificationID,
			"error", err,
		)
	}

	if sendErr != nil {
		if err := d.updateNotificationStatus(ctx, notificationID, "failed", sendErr.Error()); err != nil {
			d.logger.ErrorContext(
				ctx,
				"failed to update notification status to failed",
				"notification_id", notificationID,
				"error", err,
			)
		}

		return fmt.Errorf("email send failed: %w", sendErr)
	}

	if err := d.updateNotificationStatus(ctx, notificationID, "sent", ""); err != nil {
		d.logger.ErrorContext(
			ctx,
			"failed to update notification status to sent",
			"notification_id", notificationID,
			"error", err,
		)
	}

	d.logger.InfoContext(
		ctx,
		"email notification sent",
		"notification_id", notificationID,
		"event_id", msg.ID,
		"tenant_id", msg.TenantID,
		"to", emailTo,
		"attempt", attempt,
	)

	return nil
}

func (d *Dispatcher) handleInApp(
	ctx context.Context,
	msg queue.Message,
	data map[string]any,
	eventType string,
	externalUserID string,
) error {
	if externalUserID == "" {
		d.logger.DebugContext(
			ctx,
			"event missing user_id; skipping in_app delivery",
			"event_id", msg.ID,
			"tenant_id", msg.TenantID,
			"event_type", eventType,
		)

		return nil
	}

	email, _ := data["email"].(string)
	phone, _ := data["phone"].(string)

	internalUserID, err := d.users.EnsureUser(
		ctx,
		msg.TenantID,
		externalUserID,
		email,
		phone,
		data,
	)
	if err != nil {
		return fmt.Errorf("ensure user failed: %w", err)
	}

	tpl, err := d.templates.GetPublishedVersion(
		ctx,
		msg.TenantID,
		eventType,
		"in_app",
		"en",
	)
	if errors.Is(err, template.ErrTemplateNotFound) {
		d.logger.DebugContext(
			ctx,
			"no published in_app template found; skipping in_app delivery",
			"event_id", msg.ID,
			"tenant_id", msg.TenantID,
			"event_type", eventType,
		)

		return nil
	}

	if err != nil {
		return fmt.Errorf("resolve in_app template failed: %w", err)
	}

	rendered, err := template.Render(template.RenderInput{
		Subject: tpl.Subject,
		HTML:    tpl.HTML,
		Text:    tpl.Text,
		Data:    data,
	})
	if err != nil {
		return fmt.Errorf("in_app template render failed: %w", err)
	}

	title := strings.TrimSpace(rendered.Subject)
	if title == "" {
		title = eventType
	}

	body := strings.TrimSpace(rendered.Text)
	if body == "" {
		body = strings.TrimSpace(rendered.HTML)
	}
	if body == "" {
		body = title
	}

	actionURL, _ := data["action_url"].(string)
	imageURL, _ := data["image_url"].(string)
	category, _ := data["category"].(string)

	if category == "" {
		category = "notification"
	}

	notificationIdempotencyKey := notificationKey(msg.ID, "in_app", internalUserID)
	templateID := tpl.TemplateID

	notificationID, status, err := d.createOrGetNotification(
		ctx,
		createNotificationParams{
			Msg:             msg,
			Channel:         "in_app",
			ToAddress:       nil,
			RecipientUserID: &internalUserID,
			TemplateID:      &templateID,
			Subject:         title,
			Body:            body,
			Payload:         data,
			IdempotencyKey:  notificationIdempotencyKey,
		},
	)
	if err != nil {
		return fmt.Errorf("create or get in_app notification failed: %w", err)
	}

	inboxIdempotencyKey := inboxKey(msg.ID, internalUserID)

	itemID, created, err := d.inbox.CreateItemIdempotent(
		ctx,
		inbox.CreateItemParams{
			TenantID:       msg.TenantID,
			UserID:         internalUserID,
			NotificationID: &notificationID,
			Title:          title,
			Body:           body,
			Category:       category,
			ActionURL:      actionURL,
			ImageURL:       imageURL,
			Data:           data,
			IdempotencyKey: inboxIdempotencyKey,
		},
	)
	if err != nil {
		return fmt.Errorf("create inbox item failed: %w", err)
	}

	if !created {
		d.logger.InfoContext(
			ctx,
			"inbox item already exists; skipping duplicate in_app delivery",
			"notification_id", notificationID,
			"inbox_item_id", itemID,
			"event_id", msg.ID,
			"tenant_id", msg.TenantID,
		)

		return nil
	}

	if d.realtime != nil {
		err := d.realtime.Publish(
			ctx,
			msg.TenantID,
			internalUserID,
			realtime.Event{
				Type:   "inbox.item.created",
				ItemID: itemID,
				Data: map[string]any{
					"title":           title,
					"body":            body,
					"category":        category,
					"action_url":      actionURL,
					"image_url":       imageURL,
					"is_read":         false,
					"is_archived":     false,
					"is_pinned":       false,
					"notification_id": notificationID,
				},
			},
		)
		if err != nil {
			d.logger.WarnContext(
				ctx,
				"failed to publish realtime inbox.item.created event",
				"notification_id", notificationID,
				"inbox_item_id", itemID,
				"error", err,
			)
		}
	}

	attempt, err := d.nextAttempt(ctx, notificationID)
	if err != nil {
		attempt = 1
	}

	if err := d.recordDeliveryAttempt(
		ctx,
		notificationID,
		"success",
		attempt,
		map[string]any{
			"channel":  "in_app",
			"user_id":  internalUserID,
			"inbox_id": itemID,
		},
		map[string]any{
			"status": "delivered_to_inbox",
		},
		"",
	); err != nil {
		d.logger.ErrorContext(
			ctx,
			"failed to record in_app delivery attempt",
			"notification_id", notificationID,
			"error", err,
		)
	}

	if !isTerminalSuccessStatus(status) {
		if err := d.updateNotificationStatus(ctx, notificationID, "sent", ""); err != nil {
			d.logger.ErrorContext(
				ctx,
				"failed to update in_app notification status to sent",
				"notification_id", notificationID,
				"error", err,
			)
		}
	}

	d.logger.InfoContext(
		ctx,
		"in_app notification delivered",
		"notification_id", notificationID,
		"inbox_item_id", itemID,
		"event_id", msg.ID,
		"tenant_id", msg.TenantID,
		"user_id", internalUserID,
	)

	return nil
}

func (d *Dispatcher) resolveEmailTemplates(
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
			templateID := resolved.TemplateID
			return resolved.Subject, resolved.HTML, resolved.Text, &templateID
		}
	}

	if err != nil && !errors.Is(err, template.ErrTemplateNotFound) {
		d.logger.ErrorContext(
			ctx,
			"email template resolution failed; falling back to default template",
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
	params createNotificationParams,
) (string, string, error) {
	payload, err := json.Marshal(params.Payload)
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
			recipient_user_id,
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
			$4,
			'transactional',
			'queued',
			$5::uuid,
			$6,
			$7,
			$8,
			$9,
			$10
		)
		ON CONFLICT (tenant_id, idempotency_key) DO NOTHING
		RETURNING id::text, status
		`,
		params.Msg.TenantID,
		params.Msg.ID,
		params.TemplateID,
		params.Channel,
		params.RecipientUserID,
		params.ToAddress,
		params.Subject,
		params.Body,
		payload,
		params.IdempotencyKey,
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
		params.Msg.TenantID,
		params.IdempotencyKey,
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
	request any,
	response any,
	errText string,
) error {
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return err
	}

	responseJSON, err := json.Marshal(response)
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

func notificationKey(
	eventID string,
	channel string,
	recipient string,
) string {
	normalized := strings.ToLower(strings.TrimSpace(recipient))

	raw := fmt.Sprintf(
		"%s:%s:%s",
		eventID,
		channel,
		normalized,
	)

	sum := sha256.Sum256([]byte(raw))

	return hex.EncodeToString(sum[:])
}

func inboxKey(
	eventID string,
	internalUserID string,
) string {
	raw := fmt.Sprintf(
		"%s:in_app:%s",
		eventID,
		internalUserID,
	)

	sum := sha256.Sum256([]byte(raw))

	return hex.EncodeToString(sum[:])
}

func parseChannels(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{"email", "in_app"}
	}

	parts := strings.Split(raw, ",")
	seen := map[string]bool{}
	channels := make([]string, 0)

	for _, part := range parts {
		channel := strings.TrimSpace(part)

		if channel == "" {
			continue
		}

		if !isValidChannel(channel) {
			continue
		}

		if !seen[channel] {
			seen[channel] = true
			channels = append(channels, channel)
		}
	}

	if len(channels) == 0 {
		return []string{"email", "in_app"}
	}

	return channels
}

func isValidChannel(channel string) bool {
	switch channel {
	case "email", "in_app", "sms", "push", "webhook":
		return true
	default:
		return false
	}
}

func isTerminalSuccessStatus(status string) bool {
	switch status {
	case "sent", "delivered", "opened", "clicked":
		return true
	default:
		return false
	}
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

func (d *Dispatcher) handleSMS(
	ctx context.Context,
	msg queue.Message,
	data map[string]any,
	eventType string,
	externalUserID string,
) error {
	phone := normalizePhone(stringValue(data["phone"]))

	if phone == "" {
		d.logger.DebugContext(
			ctx,
			"event missing data.phone; skipping sms delivery",
			"event_id", msg.ID,
			"tenant_id", msg.TenantID,
			"event_type", eventType,
		)

		return nil
	}

	if !e164Regex.MatchString(phone) {
		d.logger.WarnContext(
			ctx,
			"invalid phone number; skipping sms delivery",
			"event_id", msg.ID,
			"tenant_id", msg.TenantID,
			"event_type", eventType,
			"phone", phone,
		)

		return nil
	}

	suppressed, err := d.isSuppressed(ctx, msg.TenantID, "sms", phone)
	if err != nil {
		return fmt.Errorf("sms suppression check failed: %w", err)
	}

	textTemplate := d.resolveSMSTemplate(ctx, msg, eventType)

	rendered, err := template.Render(template.RenderInput{
		Subject: "",
		HTML:    "",
		Text:    textTemplate,
		Data:    data,
	})
	if err != nil {
		return fmt.Errorf("sms template render failed: %w", err)
	}

	smsBody := strings.TrimSpace(rendered.Text)
	if smsBody == "" {
		d.logger.WarnContext(
			ctx,
			"sms template rendered empty body; skipping sms delivery",
			"event_id", msg.ID,
			"tenant_id", msg.TenantID,
			"event_type", eventType,
		)

		return nil
	}

	var recipientUserID *string

	if externalUserID != "" {
		internalUserID, err := d.users.EnsureUser(
			ctx,
			msg.TenantID,
			externalUserID,
			stringValue(data["email"]),
			phone,
			data,
		)
		if err != nil {
			return fmt.Errorf("ensure user for sms failed: %w", err)
		}

		recipientUserID = &internalUserID
	}

	idempotencyKey := notificationKey(msg.ID, "sms", phone)

	notificationID, status, err := d.createOrGetNotification(
		ctx,
		createNotificationParams{
			Msg:             msg,
			Channel:         "sms",
			ToAddress:       &phone,
			RecipientUserID: recipientUserID,
			TemplateID:      nil,
			Subject:         "",
			Body:            smsBody,
			Payload:         data,
			IdempotencyKey:  idempotencyKey,
		},
	)
	if err != nil {
		return fmt.Errorf("create or get sms notification failed: %w", err)
	}

	if isTerminalStatus(status) {
		d.logger.InfoContext(
			ctx,
			"sms notification already terminal; skipping duplicate delivery",
			"notification_id", notificationID,
			"event_id", msg.ID,
			"tenant_id", msg.TenantID,
			"status", status,
		)

		return nil
	}

	if suppressed {
		if err := d.updateNotificationStatus(ctx, notificationID, "suppressed", ""); err != nil {
			d.logger.ErrorContext(
				ctx,
				"failed to update sms notification status to suppressed",
				"notification_id", notificationID,
				"error", err,
			)
		}

		if err := d.recordDeliveryAttempt(
			ctx,
			notificationID,
			"suppressed",
			1,
			map[string]any{
				"channel": "sms",
				"phone":   phone,
			},
			map[string]any{
				"reason": "suppressed_address",
			},
			"",
		); err != nil {
			d.logger.ErrorContext(
				ctx,
				"failed to record suppressed sms attempt",
				"notification_id", notificationID,
				"error", err,
			)
		}

		d.logger.InfoContext(
			ctx,
			"sms suppressed",
			"notification_id", notificationID,
			"event_id", msg.ID,
			"tenant_id", msg.TenantID,
			"phone", phone,
		)

		return nil
	}

	attempt, err := d.nextAttempt(ctx, notificationID)
	if err != nil {
		attempt = 1
	}

	if err := d.updateNotificationStatus(ctx, notificationID, "sending", ""); err != nil {
		d.logger.ErrorContext(
			ctx,
			"failed to update sms notification status to sending",
			"notification_id", notificationID,
			"error", err,
		)
	}

	req := provider.Request{
		TenantID:       msg.TenantID,
		NotificationID: notificationID,
		Channel:        provider.ChannelSMS,
		To: provider.Recipient{
			UserID: externalUserID,
			Phone:  phone,
		},
		Text:     smsBody,
		Data:     data,
		Metadata: msg.Metadata,
	}

	result, sendErr := d.sms.Send(ctx, req)

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
			"failed to record sms delivery attempt",
			"notification_id", notificationID,
			"error", err,
		)
	}

	if sendErr != nil {
		if err := d.updateNotificationStatus(ctx, notificationID, "failed", sendErr.Error()); err != nil {
			d.logger.ErrorContext(
				ctx,
				"failed to update sms notification status to failed",
				"notification_id", notificationID,
				"error", err,
			)
		}

		return fmt.Errorf("sms send failed: %w", sendErr)
	}

	if err := d.updateNotificationStatus(ctx, notificationID, "sent", ""); err != nil {
		d.logger.ErrorContext(
			ctx,
			"failed to update sms notification status to sent",
			"notification_id", notificationID,
			"error", err,
		)
	}

	d.logger.InfoContext(
		ctx,
		"sms notification sent",
		"notification_id", notificationID,
		"event_id", msg.ID,
		"tenant_id", msg.TenantID,
		"phone", phone,
		"attempt", attempt,
	)

	return nil
}

func (d *Dispatcher) isSuppressed(
	ctx context.Context,
	tenantID string,
	channel string,
	address string,
) (bool, error) {
	var id string

	err := d.db.QueryRow(
		ctx,
		`
		SELECT id::text
		FROM suppressions
		WHERE tenant_id = $1::uuid
		  AND channel = $2
		  AND address = $3
		`,
		tenantID,
		channel,
		address,
	).Scan(&id)

	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func (d *Dispatcher) resolveSMSTemplate(
	ctx context.Context,
	msg queue.Message,
	eventType string,
) string {
	locale := "en"

	resolved, err := d.templates.GetPublishedVersion(
		ctx,
		msg.TenantID,
		eventType,
		"sms",
		locale,
	)

	if err == nil {
		if strings.TrimSpace(resolved.Text) != "" {
			return resolved.Text
		}
	}

	if err != nil && !errors.Is(err, template.ErrTemplateNotFound) {
		d.logger.ErrorContext(
			ctx,
			"sms template resolution failed; falling back to default sms template",
			"tenant_id", msg.TenantID,
			"event_type", eventType,
			"error", err,
		)
	}

	return smsTemplatesForEvent(eventType)
}

func normalizePhone(raw string) string {
	phone := strings.TrimSpace(raw)

	if phone == "" {
		return ""
	}

	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")
	phone = strings.ReplaceAll(phone, ".", "")

	return phone
}

func stringValue(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case json.Number:
		return v.String()
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}

func isTerminalStatus(status string) bool {
	if isTerminalSuccessStatus(status) {
		return true
	}

	switch status {
	case "suppressed", "canceled", "expired":
		return true
	default:
		return false
	}
}

func smsTemplatesForEvent(eventType string) string {
	switch eventType {
	case "order.created":
		return "Your order {{.order_id}} has been created. Amount: {{.amount}} {{.currency}}"
	case "payment.failed":
		return "Your payment for order {{.order_id}} failed. Please retry."
	default:
		return "Notification: {{.event}}"
	}
}

