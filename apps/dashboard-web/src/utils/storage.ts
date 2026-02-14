/**
 * Storage utility for persisting user preferences
 * Supports localStorage with fallback mechanisms
 */

const STORAGE_KEYS = {
  SIDEBAR_COLLAPSED: 'flyfleet_sidebar_collapsed',
  THEME: 'flyfleet_theme',
  LANGUAGE: 'flyfleet_language'
} as const;

type StorageKey = keyof typeof STORAGE_KEYS;

/**
 * Get item from localStorage with fallback
 */
export const getItem = <T>(key: StorageKey, fallback: T): T => {
  try {
    const stored = localStorage.getItem(STORAGE_KEYS[key]);
    if (stored === null) return fallback;
    
    // Handle boolean values specially
    if (typeof fallback === 'boolean') {
      return (stored === 'true') as T;
    }
    
    return JSON.parse(stored) as T;
  } catch (error) {
    console.warn(`Failed to get item from storage for key: ${key}`, error);
    return fallback;
  }
};

/**
 * Set item in localStorage
 */
export const setItem = <T>(key: StorageKey, value: T): void => {
  try {
    // Store booleans as strings for consistency
    const serializedValue = typeof value === 'boolean' ? String(value) : JSON.stringify(value);
    localStorage.setItem(STORAGE_KEYS[key], serializedValue);
  } catch (error) {
    console.warn(`Failed to set item in storage for key: ${key}`, error);
  }
};

/**
 * Remove item from localStorage
 */
export const removeItem = (key: StorageKey): void => {
  try {
    localStorage.removeItem(STORAGE_KEYS[key]);
  } catch (error) {
    console.warn(`Failed to remove item from storage for key: ${key}`, error);
  }
};

/**
 * Clear all FlyFleet storage items
 */
export const clearStorage = (): void => {
  try {
    Object.values(STORAGE_KEYS).forEach(key => {
      localStorage.removeItem(key);
    });
  } catch (error) {
    console.warn('Failed to clear storage', error);
  }
};

/**
 * Specific utilities for sidebar state
 */
export const getSidebarCollapsed = (): boolean => {
  return getItem('SIDEBAR_COLLAPSED', false);
};

export const setSidebarCollapsed = (collapsed: boolean): void => {
  setItem('SIDEBAR_COLLAPSED', collapsed);
};

/**
 * Check if localStorage is available
 */
export const isLocalStorageAvailable = (): boolean => {
  try {
    const testKey = '__flyfleet_storage_test__';
    localStorage.setItem(testKey, 'test');
    localStorage.removeItem(testKey);
    return true;
  } catch (error) {
    return false;
  }
};