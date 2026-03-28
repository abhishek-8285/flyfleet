import { sanitizeHTML, sanitizeURL } from './sanitize';

describe('sanitizeHTML', () => {
  it('should allow safe HTML tags', () => {
    const input = '<p>Hello <strong>World</strong></p>';
    const result = sanitizeHTML(input);
    expect(result).toBe(input);
  });

  it('should remove script tags', () => {
    const input = '<script>alert("XSS")</script><p>Safe</p>';
    const result = sanitizeHTML(input);
    expect(result).not.toContain('<script>');
    expect(result).toContain('<p>Safe</p>');
  });

  it('should remove event handlers', () => {
    const input = '<img src=x onerror=alert(1)>';
    const result = sanitizeHTML(input);
    expect(result).not.toContain('onerror');
  });

  it('should remove javascript: URLs', () => {
    const input = '<a href="javascript:alert(1)">Click</a>';
    const result = sanitizeHTML(input);
    expect(result).not.toContain('javascript:');
  });

  it('should handle null or undefined gracefully', () => {
    expect(sanitizeHTML(null as any)).toBe('');
    expect(sanitizeHTML(undefined as any)).toBe('');
    expect(sanitizeHTML(123 as any)).toBe('');
  });
});

describe('sanitizeURL', () => {
  it('should allow https URLs', () => {
    expect(sanitizeURL('https://example.com')).toBe('https://example.com');
  });

  it('should allow http URLs', () => {
    expect(sanitizeURL('http://example.com')).toBe('http://example.com');
  });

  it('should allow mailto URLs', () => {
    expect(sanitizeURL('mailto:test@example.com')).toBe('mailto:test@example.com');
  });

  it('should allow tel URLs', () => {
    expect(sanitizeURL('tel:+1234567890')).toBe('tel:+1234567890');
  });

  it('should block javascript: URLs', () => {
    expect(sanitizeURL('javascript:alert(1)')).toBe('');
  });

  it('should block data: URLs', () => {
    expect(sanitizeURL('data:text/html,<script>alert(1)</script>')).toBe('');
  });

  it('should block URLs with spaces', () => {
    expect(sanitizeURL('http://example.com /path')).toBe('');
  });

  it('should handle null/undefined/non-string', () => {
    expect(sanitizeURL(null as any)).toBe('');
    expect(sanitizeURL(undefined as any)).toBe('');
    expect(sanitizeURL(123 as any)).toBe('');
  });
});
