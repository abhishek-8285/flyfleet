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
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gonotify/gonotify/internal/consent"
	"github.com/gonotify/gonotify/internal/inbox"
	"github.com/gonotify/gonotify/internal/observability"
	"github.com/gonotify/gonotify/internal/preference"
	"github.com/gonotify/gonotify/internal/provider"
	"github.com/gonotify/gonotify/internal/queue"
	"github.com/gonotify/gonotify/internal/realtime"
	"github.com/gonotify/gonotify/internal/suppression"
	"github.com/gonotify/gonotify/internal/template"
	"github.com/gonotify/gonotify/internal/user"
)

var e164Regex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

type Dispatcher struct {
	db          *pgxpool.Pool
	selector    channelProviderSelector
	templates   *template.Repository
	users       *user.Repository
	inbox       *inbox.Repository
	realtime    *realtime.Publisher
	preferences *preference.Repository
	consent     consent.Repository
	suppressions *suppression.Repository
	logger      *slog.Logger
}

func NewDispatcher(
	db *pgxpool.Pool,
	selector channelProviderSelector,
	templates *template.Repository,
	users *user.Repository,
	inboxRepo *inbox.Repository,
	realtimePublisher *realtime.Publisher,
	preferences *preference.Repository,
	consentRepo consent.Repository,
	suppressionsRepo *suppression.Repository,
	logger *slog.Logger,
) *Dispatcher {
	if selector == nil {
		panic("dispatcher requires a channelProviderSelector")
	}
	return &Dispatcher{
		db:           db,
		selector:     selector,
		templates:    templates,
		users:        users,
		inbox:        inboxRepo,
		realtime:     realtimePublisher,
		preferences:  preferences,
		consent:      consentRepo,
		suppressions: suppressionsRepo,
		logger:       logger,
	}
}

// checkConsent enforces that marketing and digest categories require an
// explicit opt-in consent log. For transactional and other categories, consent
// is not required and the method returns true.
func (d *Dispatcher) checkConsent(
	ctx context.Context,
	tenantID string,
	internalUserID *string,
	category string,
	channel string,
) (bool, error) {
	if d.consent == nil {
		return true, nil
	}

	if category != "marketing" && category != "digest" {
		return true, nil
	}

	if internalUserID == nil {
		return false, nil
	}

	return d.consent.HasActiveConsent(
		ctx,
		tenantID,
		*internalUserID,
		category,
		channel,
	)
}

type createNotificationParams struct {
	Msg             queue.Message
	Channel         string
	Category        string
	ToAddress       *string
	RecipientUserID *string
	TemplateID      *string
	Subject         string
	Body            string
	Payload         map[string]any
	IdempotencyKey  string
	Status          string
	ScheduledAt     *time.Time
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

func eventCategory(msg queue.Message, data map[string]any) string {
	if category, ok := msg.Metadata["category"]; ok && category != "" {
		return category
	}

	if category, ok := data["category"].(string); ok && category != "" {
		return category
	}

	return "transactional"
}

func (d *Dispatcher) isChannelAllowedForUser(
	ctx context.Context,
	tenantID string,
	userID *string,
	category string,
	channel string,
) (bool, error) {
	if userID == nil {
		if category == "marketing" || category == "digest" {
			return false, nil
		}
		return true, nil
	}

	if d.preferences == nil {
		return true, nil
	}

	d.logger.DebugContext(
		ctx,
		"preference check",
		"tenant_id", tenantID,
		"user_id", *userID,
		"category", category,
		"channel", channel,
	)

	return d.preferences.IsAllowed(
		ctx,
		tenantID,
		*userID,
		category,
		channel,
	)
}

func (d *Dispatcher) suppressNotification(
	ctx context.Context,
	params createNotificationParams,
	reason string,
) error {
	notificationID, status, _, err := d.createOrGetNotification(ctx, params)
	if err != nil {
		return err
	}

	if isTerminalStatus(status) {
		return nil
	}

	if err := d.updateNotificationStatus(ctx, notificationID, "suppressed", ""); err != nil {
		d.logger.ErrorContext(
			ctx,
			"failed to update notification status to suppressed",
			"notification_id", notificationID,
			"error", err,
		)
	}

	if err := d.setSuppressionReason(ctx, notificationID, reason); err != nil {
		d.logger.ErrorContext(
			ctx,
			"failed to record suppression reason",
			"notification_id", notificationID,
			"reason", reason,
			"error", err,
		)
	}

	if err := d.recordDeliveryAttempt(
		ctx,
		notificationID,
		"",
		"suppressed",
		1,
		map[string]any{
			"channel": params.Channel,
			"reason":  reason,
		},
		map[string]any{
			"suppressed": true,
			"reason":     reason,
		},
		"",
	); err != nil {
		d.logger.ErrorContext(
			ctx,
			"failed to record suppressed delivery attempt",
			"notification_id", notificationID,
			"error", err,
		)
	}

	d.logger.InfoContext(
		ctx,
		"notification suppressed",
		"notification_id", notificationID,
		"event_id", params.Msg.ID,
		"tenant_id", params.Msg.TenantID,
		"channel", params.Channel,
		"reason", reason,
	)

	return nil
}

// resolveDeferment determines whether delivery should be deferred instead of
// sent immediately. It returns a scheduled-at time and a reason string when the
// notification should be deferred (either because of an explicit future
// scheduled_at on the message or because the user is currently within their
// quiet-hours window). A nil result means the notification should be delivered
// now.
func (d *Dispatcher) resolveDeferment(
	ctx context.Context,
	msg queue.Message,
	internalUserID *string,
	category string,
	now time.Time,
) (*time.Time, string) {
	var scheduledAt *time.Time
	reason := ""

	if msg.ScheduledAt != nil && !msg.ScheduledAt.IsZero() && msg.ScheduledAt.After(now) {
		scheduledAt = msg.ScheduledAt
		reason = "scheduled_delivery"
	}

	if internalUserID != nil && d.preferences != nil {
		qh, err := d.preferences.GetQuietHours(ctx, *internalUserID)
		if err == nil && qh.Enabled && qh.IsActive(now) {
			quietEnd := qh.End(now)
			if scheduledAt == nil || quietEnd.After(*scheduledAt) {
				scheduledAt = &quietEnd
				reason = "quiet_hours"
			}
		}
	}

	return scheduledAt, reason
}

// scheduleNotification records a notification as "scheduled" for later delivery
// without sending it immediately. It is used when delivery is deferred due to a
// future scheduled_at or active quiet-hours window.
func (d *Dispatcher) scheduleNotification(
	ctx context.Context,
	params createNotificationParams,
	reason string,
) error {
	notificationID, status, created, err := d.createOrGetNotification(ctx, params)
	if err != nil {
		return err
	}

	if !created {
		if status != "queued" && status != "sending" {
			d.logger.DebugContext(
				ctx,
				"notification already scheduled or terminal; skipping",
				"notification_id", notificationID,
				"status", status,
			)
		}

		return nil
	}

	if err := d.recordDeliveryAttempt(
		ctx,
		notificationID,
		"",
		"scheduled",
		1,
		map[string]any{
			"channel":      params.Channel,
			"reason":       reason,
			"scheduled_at": params.ScheduledAt,
		},
		map[string]any{
			"deferral_reason": reason,
			"scheduled_at":    params.ScheduledAt,
		},
		"",
	); err != nil {
		d.logger.ErrorContext(
			ctx,
			"failed to record scheduled delivery attempt",
			"notification_id", notificationID,
			"error", err,
		)
	}

	d.logger.InfoContext(
		ctx,
		"notification scheduled for later delivery",
		"notification_id", notificationID,
		"event_id", params.Msg.ID,
		"tenant_id", params.Msg.TenantID,
		"channel", params.Channel,
		"reason", reason,
		"scheduled_at", params.ScheduledAt,
	)

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

	category := eventCategory(msg, data)

	var internalUserID *string

	if userID != "" {
		resolvedUserID, err := d.users.EnsureUser(
			ctx,
			msg.TenantID,
			userID,
			emailTo,
			stringValue(data["phone"]),
			data,
		)
		if err != nil {
			return fmt.Errorf("ensure user for email failed: %w", err)
		}

		internalUserID = &resolvedUserID
	}

	allowed, err := d.isChannelAllowedForUser(
		ctx,
		msg.TenantID,
		internalUserID,
		category,
		"email",
	)
	if err != nil {
		return fmt.Errorf("email preference check failed: %w", err)
	}

	if !allowed {
		return d.suppressNotification(
			ctx,
			createNotificationParams{
				Msg:             msg,
				Channel:         "email",
				Category:        category,
				ToAddress:       &emailTo,
				RecipientUserID: internalUserID,
				TemplateID:      nil,
				Subject:         "",
				Body:            "",
				Payload:         data,
				IdempotencyKey:  notificationKey(msg.ID, "email", emailTo),
				Status:          "queued",
			},
			"preference_disabled",
		)
	}

	hasConsent, err := d.checkConsent(ctx, msg.TenantID, internalUserID, category, "email")
	if err != nil {
		return fmt.Errorf("email consent check failed: %w", err)
	}

	if !hasConsent {
		return d.suppressNotification(
			ctx,
			createNotificationParams{
				Msg:             msg,
				Channel:         "email",
				Category:        category,
				ToAddress:       &emailTo,
				RecipientUserID: internalUserID,
				TemplateID:      nil,
				Subject:         "",
				Body:            "",
				Payload:         data,
				IdempotencyKey:  notificationKey(msg.ID, "email", emailTo),
				Status:          "queued",
			},
			"missing_consent",
		)
	}

	suppressed, err := d.suppressions.IsSuppressed(ctx, msg.TenantID, "email", emailTo)
	if err != nil {
		return fmt.Errorf("email suppression check failed: %w", err)
	}

	if suppressed {
		return d.suppressNotification(
			ctx,
			createNotificationParams{
				Msg:             msg,
				Channel:         "email",
				Category:        category,
				ToAddress:       &emailTo,
				RecipientUserID: internalUserID,
				TemplateID:      nil,
				Subject:         "",
				Body:            "",
				Payload:         data,
				IdempotencyKey:  notificationKey(msg.ID, "email", emailTo),
				Status:          "queued",
			},
			"suppressed_address",
		)
	}

	scheduledAt, deferReason := d.resolveDeferment(ctx, msg, internalUserID, category, time.Now().UTC())
	if scheduledAt != nil {
		return d.scheduleNotification(
			ctx,
			createNotificationParams{
				Msg:             msg,
				Channel:         "email",
				Category:        category,
				ToAddress:       &emailTo,
				RecipientUserID: internalUserID,
				TemplateID:      nil,
				Subject:         "",
				Body:            "",
				Payload:         data,
				IdempotencyKey:  notificationKey(msg.ID, "email", emailTo),
				Status:          "scheduled",
				ScheduledAt:     scheduledAt,
			},
			deferReason,
		)
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

	notificationID, status, _, err := d.createOrGetNotification(
		ctx,
		createNotificationParams{
			Msg:             msg,
			Channel:         "email",
			Category:        category,
			ToAddress:       &emailTo,
			RecipientUserID: internalUserID,
			TemplateID:      templateID,
			Subject:         rendered.Subject,
			Body:            rendered.Text,
			Payload:         data,
			IdempotencyKey:  idempotencyKey,
			Status:          "queued",
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

	_, sendErr := d.trySend(ctx, notificationID, attempt, req)

	if sendErr != nil {
		if errors.Is(sendErr, ErrNoProviderForChannel) {
			d.logger.DebugContext(
				ctx,
				"no email provider configured; skipping",
				"notification_id", notificationID,
				"event_id", msg.ID,
				"tenant_id", msg.TenantID,
			)
			return nil
		}

		if retryErr := d.handleSendFailure(ctx, notificationID, msg.TenantID, "email", sendErr); retryErr != nil {
			d.logger.ErrorContext(
				ctx,
				"failed to handle send failure",
				"notification_id", notificationID,
				"error", retryErr,
			)

			return fmt.Errorf("email send failed: %w", sendErr)
		}

		return nil
	}

	if err := d.updateNotificationStatus(ctx, notificationID, "sent", ""); err != nil {
		d.logger.ErrorContext(
			ctx,
			"failed to update notification status to sent",
			"notification_id", notificationID,
			"error", err,
		)
	}

	observability.NotificationsProcessed.WithLabelValues(msg.TenantID, "email", "sent").Inc()

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

	category := eventCategory(msg, data)

	allowed, err := d.isChannelAllowedForUser(
		ctx,
		msg.TenantID,
		&internalUserID,
		category,
		"in_app",
	)
	if err != nil {
		return fmt.Errorf("in_app preference check failed: %w", err)
	}

	if !allowed {
		return d.suppressNotification(
			ctx,
			createNotificationParams{
				Msg:             msg,
				Channel:         "in_app",
				Category:        category,
				ToAddress:       nil,
				RecipientUserID: &internalUserID,
				TemplateID:      nil,
				Subject:         "",
				Body:            "",
				Payload:         data,
				IdempotencyKey:  notificationKey(msg.ID, "in_app", internalUserID),
				Status:          "queued",
			},
			"preference_disabled",
		)
	}

	hasConsent, err := d.checkConsent(ctx, msg.TenantID, &internalUserID, category, "in_app")
	if err != nil {
		return fmt.Errorf("in_app consent check failed: %w", err)
	}

	if !hasConsent {
		return d.suppressNotification(
			ctx,
			createNotificationParams{
				Msg:             msg,
				Channel:         "in_app",
				Category:        category,
				ToAddress:       nil,
				RecipientUserID: &internalUserID,
				TemplateID:      nil,
				Subject:         "",
				Body:            "",
				Payload:         data,
				IdempotencyKey:  notificationKey(msg.ID, "in_app", internalUserID),
				Status:          "queued",
			},
			"missing_consent",
		)
	}

	scheduledAt, deferReason := d.resolveDeferment(
		ctx,
		msg,
		&internalUserID,
		category,
		time.Now().UTC(),
	)
	if scheduledAt != nil {
		return d.scheduleNotification(
			ctx,
			createNotificationParams{
				Msg:             msg,
				Channel:         "in_app",
				Category:        category,
				ToAddress:       nil,
				RecipientUserID: &internalUserID,
				TemplateID:      nil,
				Subject:         "",
				Body:            "",
				Payload:         data,
				IdempotencyKey:  notificationKey(msg.ID, "in_app", internalUserID),
				Status:          "scheduled",
				ScheduledAt:     scheduledAt,
			},
			deferReason,
		)
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
	inboxCategory, _ := data["category"].(string)

	if inboxCategory == "" {
		inboxCategory = "notification"
	}

	notificationIdempotencyKey := notificationKey(msg.ID, "in_app", internalUserID)
	templateID := tpl.TemplateID

	notificationID, status, _, err := d.createOrGetNotification(
		ctx,
		createNotificationParams{
			Msg:             msg,
			Channel:         "in_app",
			Category:        category,
			ToAddress:       nil,
			RecipientUserID: &internalUserID,
			TemplateID:      &templateID,
			Subject:         title,
			Body:            body,
			Payload:         data,
			IdempotencyKey:  notificationIdempotencyKey,
			Status:          "queued",
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
			Category:       inboxCategory,
			ActionURL:      actionURL,
			ImageURL:       imageURL,
			Data:           data,
			IdempotencyKey: inboxIdempotencyKey,
		},
	)
	if err != nil {
		if retryErr := d.handleSendFailure(ctx, notificationID, msg.TenantID, "in_app", err); retryErr != nil {
			d.logger.ErrorContext(
				ctx,
				"failed to handle in_app send failure",
				"notification_id", notificationID,
				"error", retryErr,
			)

			return fmt.Errorf("create inbox item failed: %w", err)
		}

		return nil
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
		"",
		"success",
		attempt,
		map[string]any{
			"channel": "in_app",
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
) (notificationID string, status string, created bool, err error) {
	payload, err := json.Marshal(params.Payload)
	if err != nil {
		return "", "", false, err
	}

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
			idempotency_key,
			scheduled_at
		)
		VALUES (
			$1::uuid,
			$2::uuid,
			$3::uuid,
			$4,
			$5,
			$6,
			$7::uuid,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13
		)
		ON CONFLICT (tenant_id, idempotency_key) DO NOTHING
		RETURNING id::text, status
		`,
		params.Msg.TenantID,
		params.Msg.ID,
		params.TemplateID,
		params.Channel,
		params.Category,
		params.Status,
		params.RecipientUserID,
		params.ToAddress,
		params.Subject,
		params.Body,
		payload,
		params.IdempotencyKey,
		params.ScheduledAt,
	).Scan(&notificationID, &status)

	if err == nil {
		return notificationID, status, true, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return "", "", false, err
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
		return "", "", false, err
	}

	return notificationID, status, false, nil
}

// trySend selects providers for the request's channel via the routing
// engine and tries each one in order of priority/weight. It records a
// delivery attempt for every provider tried and returns the result of
// the first successful send.
func (d *Dispatcher) trySend(
	ctx context.Context,
	notificationID string,
	attempt int,
	req provider.Request,
) (provider.Result, error) {
	selected, err := d.selector.ProvidersForChannel(ctx, req.TenantID, string(req.Channel))
	if err != nil {
		return provider.Result{}, fmt.Errorf("select %s providers: %w", req.Channel, err)
	}

	if len(selected) == 0 {
		return provider.Result{}, ErrNoProviderForChannel
	}

	var lastErr error
	var lastResult provider.Result

	for _, sp := range selected {
		d.logger.DebugContext(ctx, "trying provider",
			"notification_id", notificationID,
			"provider", sp.Provider.Name(),
			"channel", req.Channel,
			"config_id", sp.ConfigID,
		)

		result, sendErr := sp.Provider.Send(ctx, req)

		var attemptStatus string
		var errText string
		if sendErr != nil {
			attemptStatus = "failed"
			errText = sendErr.Error()
			lastErr = sendErr
		} else {
			attemptStatus = "success"
		}

		lastResult = result

		if recordErr := d.recordDeliveryAttempt(
			ctx,
			notificationID,
			sp.ConfigID,
			attemptStatus,
			attempt,
			req,
			result,
			errText,
		); recordErr != nil {
			d.logger.WarnContext(ctx, "failed to record delivery attempt",
				"notification_id", notificationID,
				"error", recordErr,
			)
		}

		if sendErr == nil {
			d.logger.InfoContext(ctx, "provider send succeeded",
				"notification_id", notificationID,
				"provider", sp.Provider.Name(),
				"channel", req.Channel,
			)
			return result, nil
		}

		d.logger.WarnContext(ctx, "provider send failed, trying next",
			"notification_id", notificationID,
			"provider", sp.Provider.Name(),
			"channel", req.Channel,
			"error", sendErr,
		)
	}

	return lastResult, fmt.Errorf("all %s providers failed: %w", req.Channel, lastErr)
}

// handleSendFailure classifies a provider send error and either schedules
// a retry, moves the notification to dead_letter, or marks it as permanently
// failed (non-retryable). Returns nil on success; non-nil if the DB update
// fails (signalling NATS to redeliver).
func (d *Dispatcher) handleSendFailure(
	ctx context.Context,
	notificationID string,
	tenantID string,
	channel string,
	sendErr error,
) error {
	var attemptCount int
	var maxAttempts int
	var retryPolicy string

	err := d.db.QueryRow(ctx, `
		SELECT attempt_count, max_attempts, retry_policy
		FROM notifications
		WHERE id = $1::uuid
	`, notificationID).Scan(&attemptCount, &maxAttempts, &retryPolicy)
	if err != nil {
		return fmt.Errorf("load notification for retry: %w", err)
	}

	if !isRetryableError(sendErr) {
		d.logger.InfoContext(ctx, "notification failed permanently (non-retryable)",
			"notification_id", notificationID,
			"error", sendErr,
		)

		observability.NotificationsProcessed.WithLabelValues(tenantID, channel, "failed").Inc()

		_, err := d.db.Exec(ctx, `
			UPDATE notifications
			SET
				status = 'failed',
				error = $2,
				failed_at = now(),
				next_retry_at = NULL
			WHERE id = $1::uuid
		`, notificationID, sendErr.Error())

		return err
	}

	attemptCount++

	if attemptCount >= maxAttempts {
		d.logger.InfoContext(ctx, "notification exhausted retries; moving to dead_letter",
			"notification_id", notificationID,
			"attempt_count", attemptCount,
			"max_attempts", maxAttempts,
		)

		observability.DeadLetterMoved.WithLabelValues(tenantID, channel).Inc()
		observability.DeadLetterPending.WithLabelValues(tenantID).Inc()

		_, err := d.db.Exec(ctx, `
			UPDATE notifications
			SET
				status = 'dead_letter',
				attempt_count = $2,
				error = $3,
				failed_at = now(),
				next_retry_at = NULL
			WHERE id = $1::uuid
		`, notificationID, attemptCount, sendErr.Error())

		return err
	}

	observability.RetriesScheduled.WithLabelValues(tenantID, channel).Inc()

	delay := RetryDelay(attemptCount, retryPolicy)

	_, err = d.db.Exec(ctx, `
		UPDATE notifications
		SET
			status = 'retry_pending',
			attempt_count = $2,
			error = $3,
			next_retry_at = now() + $4::interval
		WHERE id = $1::uuid
	`, notificationID, attemptCount, sendErr.Error(), delay.String())

	if err != nil {
		return err
	}

	d.logger.InfoContext(ctx, "notification scheduled for retry",
		"notification_id", notificationID,
		"attempt_count", attemptCount,
		"max_attempts", maxAttempts,
		"delay", delay.String(),
	)

	return nil
}

func (d *Dispatcher) recordDeliveryAttempt(
	ctx context.Context,
	notificationID string,
	providerConfigID string,
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
			NULLIF($2::text, '')::uuid,
			$3,
			$4,
			$5,
			$6,
			NULLIF($7, '')
		)
		`,
		notificationID,
		providerConfigID,
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

func (d *Dispatcher) setSuppressionReason(
	ctx context.Context,
	notificationID string,
	reason string,
) error {
	_, err := d.db.Exec(
		ctx,
		`UPDATE notifications SET suppress_reason = $2 WHERE id = $1::uuid`,
		notificationID,
		reason,
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

	suppressed, err := d.suppressions.IsSuppressed(ctx, msg.TenantID, "sms", phone)
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

	category := eventCategory(msg, data)

	allowed, err := d.isChannelAllowedForUser(
		ctx,
		msg.TenantID,
		recipientUserID,
		category,
		"sms",
	)
	if err != nil {
		return fmt.Errorf("sms preference check failed: %w", err)
	}

	if !allowed {
		return d.suppressNotification(
			ctx,
			createNotificationParams{
				Msg:             msg,
				Channel:         "sms",
				Category:        category,
				ToAddress:       &phone,
				RecipientUserID: recipientUserID,
				TemplateID:      nil,
				Subject:         "",
				Body:            "",
				Payload:         data,
				IdempotencyKey:  notificationKey(msg.ID, "sms", phone),
				Status:          "queued",
			},
			"preference_disabled",
		)
	}

	hasConsent, err := d.checkConsent(ctx, msg.TenantID, recipientUserID, category, "sms")
	if err != nil {
		return fmt.Errorf("sms consent check failed: %w", err)
	}

	if !hasConsent {
		return d.suppressNotification(
			ctx,
			createNotificationParams{
				Msg:             msg,
				Channel:         "sms",
				Category:        category,
				ToAddress:       &phone,
				RecipientUserID: recipientUserID,
				TemplateID:      nil,
				Subject:         "",
				Body:            "",
				Payload:         data,
				IdempotencyKey:  notificationKey(msg.ID, "sms", phone),
				Status:          "queued",
			},
			"missing_consent",
		)
	}

	if !suppressed {
		scheduledAt, deferReason := d.resolveDeferment(
			ctx,
			msg,
			recipientUserID,
			category,
			time.Now().UTC(),
		)
		if scheduledAt != nil {
			return d.scheduleNotification(
				ctx,
				createNotificationParams{
					Msg:             msg,
					Channel:         "sms",
					Category:        category,
					ToAddress:       &phone,
					RecipientUserID: recipientUserID,
					TemplateID:      nil,
					Subject:         "",
					Body:            "",
					Payload:         data,
					IdempotencyKey:  notificationKey(msg.ID, "sms", phone),
					Status:          "scheduled",
					ScheduledAt:     scheduledAt,
				},
				deferReason,
			)
		}
	}

	idempotencyKey := notificationKey(msg.ID, "sms", phone)

	notificationID, status, _, err := d.createOrGetNotification(
		ctx,
		createNotificationParams{
			Msg:             msg,
			Channel:         "sms",
			Category:        category,
			ToAddress:       &phone,
			RecipientUserID: recipientUserID,
			TemplateID:      nil,
			Subject:         "",
			Body:            smsBody,
			Payload:         data,
			IdempotencyKey:  idempotencyKey,
			Status:          "queued",
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
			"",
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

	_, sendErr := d.trySend(ctx, notificationID, attempt, req)

	if sendErr != nil {
		if errors.Is(sendErr, ErrNoProviderForChannel) {
			d.logger.DebugContext(
				ctx,
				"no sms provider configured; skipping",
				"notification_id", notificationID,
				"event_id", msg.ID,
				"tenant_id", msg.TenantID,
				"phone", phone,
			)
			return nil
		}

		if retryErr := d.handleSendFailure(ctx, notificationID, msg.TenantID, "sms", sendErr); retryErr != nil {
			d.logger.ErrorContext(
				ctx,
				"failed to handle send failure",
				"notification_id", notificationID,
				"error", retryErr,
			)

			return fmt.Errorf("sms send failed: %w", sendErr)
		}

		return nil
	}

	if err := d.updateNotificationStatus(ctx, notificationID, "sent", ""); err != nil {
		d.logger.ErrorContext(
			ctx,
			"failed to update sms notification status to sent",
			"notification_id", notificationID,
			"error", err,
		)
	}

	observability.NotificationsProcessed.WithLabelValues(msg.TenantID, "sms", "sent").Inc()

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
