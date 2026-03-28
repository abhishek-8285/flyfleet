// jest-dom adds custom jest matchers for asserting on DOM nodes.
// allows you to do things like:
// expect(element).toHaveTextContent(/react/i)
// learn more: https://github.com/testing-library/jest-dom
import '@testing-library/jest-dom';
import { vi } from 'vitest';

// Alias jest to vi for Jest API compatibility during migration
(globalThis as any).jest = vi;

// localStorage mock for jsdom (Vitest doesn't implement localStorage by default)
class LocalStorageMock {
  private store: Record<string, string> = {};

  clear() {
    this.store = {};
  }

  getItem(key: string) {
    return this.store[key] ?? null;
  }

  setItem(key: string, value: string) {
    this.store[key] = String(value);
  }

  removeItem(key: string) {
    delete this.store[key];
  }

  key(index: number) {
    const keys = Object.keys(this.store);
    return keys[index] ?? null;
  }

  get length() {
    return Object.keys(this.store).length;
  }
}

// Replace window.localStorage with mock
Object.defineProperty(window, 'localStorage', {
  value: new LocalStorageMock(),
  writable: true,
});
