import DOMPurify from 'dompurify';

/**
 * Sanitizes HTML string to prevent XSS attacks.
 * Uses DOMPurify with a strict configuration.
 *
 * @param dirty - The raw HTML string that may contain unsafe content
 * @returns Sanitized HTML safe for rendering
 */
export const sanitizeHTML = (dirty: string): string => {
  if (typeof dirty !== 'string') {
    return '';
  }

  return DOMPurify.sanitize(dirty, {
    ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'a', 'p', 'br', 'ul', 'ol', 'li', 'span', 'div', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'blockquote', 'code', 'pre'],
    ALLOWED_ATTR: ['href', 'target', 'rel', 'class', 'id', 'style'],
    ALLOW_DATA_ATTR: false,
    KEEP_CONTENT: false,
  });
};

/**
 * Sanitizes a URL to prevent javascript: and data: XSS vectors
 *
 * @param url - The URL to sanitize
 * @returns Sanitized URL or empty string if unsafe
 */
export const sanitizeURL = (url: string): string => {
  if (typeof url !== 'string') {
    return '';
  }

  // Basic validation: only allow http, https, mailto, tel
  const allowedProtocols = /^(https?:|mailto:|tel:)/i;
  if (!allowedProtocols.test(url)) {
    return '';
  }

  // Additional check: no embedded spaces or control chars
  if (/\s/.test(url)) {
    return '';
  }

  return url;
};
