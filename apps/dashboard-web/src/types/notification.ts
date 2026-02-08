export type NotificationChannel = 'EMAIL' | 'SMS' | 'PUSH' | 'IN_APP';
export type NotificationStatus = 'QUEUED' | 'SENT' | 'FAILED' | 'READ' | 'BOUNCED';

export interface NotificationEvent {
  id: string;
  notification_id: string;
  event_type: string;
  message: string;
  created_at: string;
}

export interface NotificationLog {
  id: string;
  correlation_id: string;
  user_id?: string;
  user_email: string;
  channel: NotificationChannel;
  type: string;
  subject: string;
  body?: string;
  status: NotificationStatus;
  provider?: string;
  provider_response?: string;
  retry_count: number;
  max_retries: number;
  metadata?: Record<string, any>;
  created_at: string;
  updated_at: string;
  items?: NotificationEvent[];
}

export interface NotificationSettings {
  id: string;
  channel: NotificationChannel;
  enabled: boolean;
  max_retries: number;
  priority: string;
  retention_months: number;
  updated_at: string;
}

export interface ResendRequest {
  reason: string;
  high_priority: boolean;
}
export interface UserNotificationPreference {
  user_id: string;
  user_name?: string;
  user_email?: string;
  email_enabled: boolean;
  sms_enabled: boolean;
  push_enabled: boolean;
  in_app_enabled: boolean;
  last_modified_at: string;
}

export interface PreferenceAudit {
  id: string;
  user_id: string;
  changed_by: string;
  old_values: any;
  new_values: any;
  change_reason: string;
  created_at: string;
}

export interface GlobalNotificationDefault {
  email_enabled: boolean;
  sms_enabled: boolean;
  push_enabled: boolean;
  in_app_enabled: boolean;
  updated_at: string;
}
