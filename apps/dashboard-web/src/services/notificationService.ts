import axios from 'axios';
import { GlobalNotificationDefault, NotificationLog, NotificationSettings, PreferenceAudit, ResendRequest, UserNotificationPreference } from '../types/notification';

const API_BASE = '/api/v1/admin/notifications';
const PREFS_BASE = '/api/v1/admin/preferences';

export const notificationService = {
  getLogs: async (params: { page?: number; limit?: number; status?: string; channel?: string; search?: string }) => {
    const response = await axios.get<{ data: NotificationLog[]; total: number; page: number; limit: number }>(API_BASE, { params });
    return response.data;
  },

  getLogDetail: async (id: string) => {
    const response = await axios.get<NotificationLog>(`${API_BASE}/${id}`);
    return response.data;
  },

  resendNotification: async (id: string, request: ResendRequest) => {
    const response = await axios.post<{ new_notification_id: string }>(`${API_BASE}/${id}/resend`, request);
    return response.data;
  },

  getSettings: async () => {
    const response = await axios.get<NotificationSettings[]>(`${API_BASE}/settings`);
    return response.data;
  },

  updateSettings: async (settings: NotificationSettings[]) => {
    const response = await axios.put<{ status: string }>(`${API_BASE}/settings`, settings);
    return response.data;
  },

  getPreferences: async (params: { page?: number; limit?: number; search?: string }) => {
    const response = await axios.get<{ items: UserNotificationPreference[]; total: number; page: number; limit: number }>(PREFS_BASE, { params });
    return response.data;
  },

  getUserPreference: async (userId: string) => {
    const response = await axios.get<UserNotificationPreference>(`${PREFS_BASE}/${userId}`);
    return response.data;
  },

  updateUserPreference: async (userId: string, data: { email_enabled: boolean; sms_enabled: boolean; push_enabled: boolean; in_app_enabled: boolean; reason: string }) => {
    const response = await axios.put<{ status: string }>(`${PREFS_BASE}/${userId}`, data);
    return response.data;
  },

  getGlobalDefaults: async () => {
    const response = await axios.get<GlobalNotificationDefault>(`${PREFS_BASE}/global`);
    return response.data;
  },

  updateGlobalDefaults: async (data: Omit<GlobalNotificationDefault, 'updated_at'>) => {
    const response = await axios.put<{ status: string }>(`${PREFS_BASE}/global`, data);
    return response.data;
  },

  getPreferenceAudit: async (userId: string) => {
    const response = await axios.get<PreferenceAudit[]>(`${PREFS_BASE}/${userId}/audit`);
    return response.data;
  }
};
