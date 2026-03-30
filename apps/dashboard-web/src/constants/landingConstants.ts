/**
 * Landing Page Constants
 * Centralized constants for the landing page to eliminate magic numbers
 * and ensure consistent styling across components.
 */

// ============================================================================
// Layout Constants
// ============================================================================

export const LANDING_MAX_WIDTH = '1440px';
export const LANDING_CONTENT_MAX_WIDTH = '1280px';
export const LANDING_SECTION_MAX_WIDTH = '1280px';

export const LANDING_PADDING_X = 'px-4 md:px-10' as const;
export const LANDING_PADDING_Y = 'py-16 md:py-24' as const;

// ============================================================================
// Typography Constants
// ============================================================================

export const HERO_TITLE_SIZE = 'text-4xl md:text-6xl lg:text-7xl' as const;
export const HERO_SUBTITLE_SIZE = 'text-lg md:text-xl' as const;
export const SECTION_TITLE_SIZE = 'text-3xl md:text-4xl' as const;
export const SECTION_SUBTITLE_SIZE = 'text-lg' as const;
export const CARD_TITLE_SIZE = 'text-xl' as const;
export const CARD_SUBTITLE_SIZE = 'text-sm' as const;

// ============================================================================
// Spacing Constants
// ============================================================================

export const SECTION_PADDING_Y_LARGE = 'py-20' as const;
export const SECTION_PADDING_Y_XLARGE = 'py-24' as const;
export const HERO_PADDING_Y = 'py-16 md:py-24' as const;
export const CARD_PADDING = 'p-8' as const;
export const CARD_GAP = 'gap-8' as const;

// ============================================================================
// Color Constants (Tailwind classes)
// ============================================================================

export const PRIMARY_COLOR = 'primary' as const;
export const SURFACE_DARK = '#111a22' as const;
export const SURFACE_DARK_ALT = '#0d141c' as const;
export const BACKGROUND_DARK = '#111a22' as const;

// Shadow classes
export const PRIMARY_SHADOW = 'shadow-primary/25' as const;
export const PRIMARY_SHADOW_LARGE = 'shadow-primary/30' as const;
export const CARD_SHADOW = 'shadow-sm' as const;
export const CARD_SHADOW_HOVER = 'shadow-xl' as const;

// ============================================================================
// Border Constants
// ============================================================================

export const BORDER_LIGHT = 'border-slate-200' as const;
export const BORDER_DARK = 'border-border-dark' as const;
export const BORDER_RADIUS_CARD = 'rounded-2xl' as const;
export const BORDER_RADIUS_BUTTON = 'rounded-lg' as const;

// ============================================================================
// Animation Constants
// ============================================================================

export const TRANSITION_BASE = 'transition-all' as const;
export const TRANSITION_COLOR = 'transition-colors' as const;
export const TRANSITION_TRANSFORM = 'transition-transform' as const;
export const HOVER_SCALE = 'hover:scale-110' as const;
export const HOVER_SHADOW = 'hover:shadow-lg' as const;

// ============================================================================
// Z-Index Constants
// ============================================================================

export const Z_INDEX_TOP_BAR = 60 as const;
export const Z_INDEX_HEADER = 50 as const;
export const Z_INDEX_DROPDOWN = 50 as const;
export const Z_INDEX_MODAL = 100 as const;

// ============================================================================
// Header Constants
// ============================================================================

export const HEADER_HEIGHT = 'h-16' as const;
export const HEADER_TOP_OFFSET = 'top-10' as const;
export const HEADER_BACKDROP = 'backdrop-blur-md' as const;
export const HEADER_BG_LIGHT = 'bg-white/80' as const;
export const HEADER_BG_DARK = 'bg-[#111a22]/80' as const;

// ============================================================================
// Button Constants
// ============================================================================

export const BUTTON_HEIGHT = 'h-9' as const;
export const BUTTON_HEIGHT_LARGE = 'h-12' as const;
export const BUTTON_PADDING = 'px-4' as const;
export const BUTTON_PADDING_LARGE = 'px-8' as const;
export const BUTTON_FONT_SIZE = 'text-sm' as const;
export const BUTTON_FONT_SIZE_LARGE = 'text-base' as const;

// Primary button styles
export const BUTTON_PRIMARY = 'bg-primary hover:bg-primary/90 text-white' as const;
export const BUTTON_SECONDARY = 'border border-slate-200 dark:border-border-dark text-slate-900 dark:text-white hover:bg-slate-100 dark:hover:bg-surface-dark' as const;
export const BUTTON_OUTLINE = 'border border-slate-300 dark:border-border-dark bg-white dark:bg-surface-dark hover:bg-slate-50 dark:hover:bg-border-dark' as const;

// ============================================================================
// Mobile Frame Constants (for app showcase)
// ============================================================================

export const MOBILE_FRAME_WIDTH = 'w-[280px]' as const;
export const MOBILE_FRAME_HEIGHT = 'h-[580px]' as const;
export const MOBILE_FRAME_SHADOW = 'shadow-2xl' as const;
export const MOBILE_FRAME_BG = 'bg-[#111a22]' as const;

// ============================================================================
// Grid Constants
// ============================================================================

export const GRID_COLS_SUCCESS = 'grid-cols-1 md:grid-cols-3' as const;
export const GRID_COLS_FEATURES = 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3' as const;
export const GRID_COLS_INTEGRATIONS = 'grid-cols-2 md:grid-cols-4 lg:grid-cols-6' as const;
export const GRID_GAP = 'gap-6' as const;
export const GRID_GAP_LARGE = 'gap-8' as const;
export const GRID_GAP_XLARGE = 'gap-12 md:gap-20' as const;

// ============================================================================
// Section ID Constants
// ============================================================================

export const SECTION_IDS = {
  SUCCESS_STORIES: 'success-stories',
  FEATURES: 'features',
  APPS: 'apps',
  PRICING: 'pricing',
  SECURITY: 'security',
  HARDWARE: 'hardware',
  FAQ: 'faq',
} as const;

// ============================================================================
// Default Values
// ============================================================================

export const DEFAULT_PHONE_NUMBER = '+15551234567' as const;
export const DEFAULT_SALES_BTN = 'Talk to Sales' as const;
export const DEFAULT_LOCALE = 'US (EN)' as const;

// ============================================================================
// Icon Names (Material Symbols)
// ============================================================================

export const ICONS = {
  SECURITY: 'verified_user',
  PHONE: 'call',
  LANGUAGE: 'language',
  EXPAND_MORE: 'expand_more',
  TRUCK: 'local_shipping',
  LIGHT_MODE: 'light_mode',
  DARK_MODE: 'dark_mode',
  ARROW_FORWARD: 'arrow_forward',
  PLAY_CIRCLE: 'play_circle',
  PDF: 'picture_as_pdf',
  SEARCH: 'search',
  SOS: 'sos',
  RADIO_CHECKED: 'radio_button_checked',
  RADIO_UNCHECKED: 'radio_button_unchecked',
  HOME: 'home',
  MAP: 'map',
  PERSON: 'person',
  NAVIGATE: 'navigate',
  DETAILS: 'info',
} as const;

// ============================================================================
// Animation Delays (for staggered animations)
// ============================================================================

export const ANIMATION_DELAY_100 = 'delay-100' as const;
export const ANIMATION_DELAY_200 = 'delay-200' as const;
export const ANIMATION_DELAY_300 = 'delay-300' as const;

// ============================================================================
// Breakpoint Hints (for responsive design)
// ============================================================================

export const BREAKPOINTS = {
  MOBILE: 640,
  TABLET: 768,
  DESKTOP: 1024,
  LARGE_DESKTOP: 1280,
} as const;
