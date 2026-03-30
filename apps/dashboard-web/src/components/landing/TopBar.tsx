import branding from '@flyfleet/config';
import { LANDING_PADDING_X, Z_INDEX_TOP_BAR, ICONS, DEFAULT_PHONE_NUMBER, DEFAULT_SALES_BTN } from '../../constants/landingConstants';
import { useUIStrings } from '../../hooks/useUIStrings';

interface TopBarProps {
  landingMetadata?: {
    sales_info?: {
      phone?: string;
    };
  } | null;
}

/**
 * TopBar Component
 * Displays the top info bar with security badge and sales contact
 */
export function TopBar({ landingMetadata }: TopBarProps) {
  const { uiStrings } = useUIStrings('landing');

  return (
    <div
      className={`fixed top-0 w-full z-[${Z_INDEX_TOP_BAR}] bg-[#0f65bd] text-white text-sm font-medium h-10 flex items-center justify-center shadow-md`}
    >
      <div className={`max-w-[${branding.companyName === 'FlyFleet' ? '1440px' : '100%'}] w-full ${LANDING_PADDING_X} flex justify-between items-center`}>
        <div className="flex items-center gap-2 text-xs md:text-sm opacity-90">
          <span className="material-symbols-outlined !text-lg" data-testid="security-icon">
            {ICONS.SECURITY}
          </span>
          <span>
            {uiStrings?.landing?.security_text || 'Operations Grade Security'}
          </span>
        </div>
        <div className="flex items-center gap-6">
          <a
            className="flex items-center gap-2 hover:text-white text-blue-100 transition-colors font-bold group"
            href={`tel:${landingMetadata?.sales_info?.phone || DEFAULT_PHONE_NUMBER}`}
          >
            <span className="material-symbols-outlined !text-lg group-hover:animate-pulse">
              {ICONS.PHONE}
            </span>
            {uiStrings?.landing?.sales_btn || DEFAULT_SALES_BTN}:{' '}
            {landingMetadata?.sales_info?.phone || '+--'}
          </a>
          <div className="h-4 w-px bg-white/20 hidden md:block"></div>
          <div className="hidden md:flex items-center gap-1 cursor-pointer hover:text-white text-blue-100 group relative">
            <span className="material-symbols-outlined !text-lg">
              {ICONS.LANGUAGE}
            </span>
            <span className="text-xs font-bold">
              {uiStrings?.landing?.locale_en || 'US (EN)'}
            </span>
            <span className="material-symbols-outlined !text-sm">
              {ICONS.EXPAND_MORE}
            </span>
            <div className="absolute top-full right-0 mt-2 w-32 bg-white dark:bg-surface-dark rounded-lg shadow-xl border border-slate-200 dark:border-border-dark py-2 hidden group-hover:block text-slate-900 dark:text-white z-50">
              <a
                className="block px-4 py-2 hover:bg-slate-50 dark:hover:bg-background-dark text-xs"
                href="#"
              >
                Global
              </a>
              <a
                className="block px-4 py-2 hover:bg-slate-50 dark:hover:bg-background-dark text-xs"
                href="#"
              >
                Europe (EU)
              </a>
              <a
                className="block px-4 py-2 hover:bg-slate-50 dark:hover:bg-background-dark text-xs font-bold text-primary"
                href="#"
              >
                United States
              </a>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
