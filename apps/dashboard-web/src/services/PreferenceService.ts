import apiClient from './apiClient';

export interface UserPreferences {
  sidebar_collapsed?: boolean;
  language?: string;
  theme?: string;
}

class PreferenceService {
  /**
   * Updates user preferences in the backend.
   * This ensures consistency across different devices and sessions.
   */
  async updatePreferences(preferences: UserPreferences): Promise<void> {
    try {
      // Use /auth/profile for self-updates
      await apiClient.put('/auth/profile', preferences);
      console.log('[PreferenceService] Preferences synchronized with backend:', preferences);
    } catch (error) {
      console.error('[PreferenceService] Failed to sync preferences:', error);
      // We don't throw here to avoid blocking UI state changes if backend sync fails
      // The local state remains as the primary source of truth for the current session
    }
  }

  /**
   * Specifically update theme
   */
  async updateTheme(theme: 'light' | 'dark'): Promise<void> {
    return this.updatePreferences({ theme });
  }

  /**
   * Specifically update language
   */
  async updateLanguage(language: string): Promise<void> {
    return this.updatePreferences({ language });
  }
}

export default new PreferenceService();
