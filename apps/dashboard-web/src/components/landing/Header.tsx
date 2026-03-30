import { useNavigate } from 'react-router-dom';
import { LANDING_PADDING_X, HEADER_HEIGHT, HEADER_TOP_OFFSET, HEADER_BACKDROP, HEADER_BG_LIGHT, HEADER_BG_DARK, BORDER_LIGHT, BORDER_DARK, Z_INDEX_HEADER, ICONS } from '../../constants/landingConstants';
import { useUIStrings } from '../../hooks/useUIStrings';

interface HeaderProps {
  isDark: boolean;
  onToggleTheme: () => void;
}

/**
 * Header Component
 * Main navigation header with logo, nav links, and action buttons
 */
export function Header({ isDark, onToggleTheme }: HeaderProps) {
  const navigate = useNavigate();
  const { uiStrings } = useUIStrings('landing');

  const navLinks = [
    { href: '#success-stories', label: uiStrings?.landing?.success_stories || 'Success Stories' },
    { href: '#features', label: uiStrings?.landing?.platform || 'Platform' },
    { href: '#apps', label: uiStrings?.landing?.apps || 'Apps' },
    { href: '#pricing', label: uiStrings?.landing?.pricing || 'Pricing' },
    { href: '#security', label: uiStrings?.landing?.security || 'Security' },
  ];

  return (
    <header className={`fixed ${HEADER_TOP_OFFSET} z-[${Z_INDEX_HEADER}] w-full border-b border-solid ${BORDER_LIGHT} dark:${BORDER_DARK} ${HEADER_BG_LIGHT} dark:${HEADER_BG_DARK} ${HEADER_BACKDROP} transition-all`}>
      <div className={`${LANDING_PADDING_X} flex h-[${HEADER_HEIGHT}] items-center justify-between max-w-[1440px] mx-auto`}>
        <div className="flex items-center gap-4 text-slate-900 dark:text-white">
          <div className="size-8 flex items-center justify-center text-primary">
            <span className="material-symbols-outlined !text-3xl">
              {ICONS.TRUCK}
            </span>
          </div>
          <h2 className="text-lg font-bold leading-tight tracking-[-0.015em]">
            FlyFleet
          </h2>
        </div>
        
        <nav className="hidden md:flex flex-1 justify-center gap-8">
          {navLinks.map((link) => (
            <a
              key={link.href}
              className="text-slate-600 dark:text-text-secondary text-sm font-medium hover:text-primary transition-colors"
              href={link.href}
            >
              {link.label}
            </a>
          ))}
        </nav>
        
        <div className="flex gap-4 items-center">
          <button
            onClick={onToggleTheme}
            className="p-2 rounded-full hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-600 dark:text-slate-300 transition-colors"
            aria-label="Toggle theme"
          >
            <span className="material-symbols-outlined">
              {isDark ? ICONS.LIGHT_MODE : ICONS.DARK_MODE}
            </span>
          </button>
          <button
            onClick={() => navigate('/login')}
            className="hidden sm:flex h-9 items-center justify-center rounded-lg px-4 bg-transparent border border-slate-200 dark:border-border-dark text-slate-900 dark:text-white text-sm font-bold hover:bg-slate-100 dark:hover:bg-surface-dark transition-colors"
          >
            <span className="truncate">
              {uiStrings?.landing?.login_btn || 'Login'}
            </span>
          </button>
          <button className="flex h-9 items-center justify-center rounded-lg px-4 bg-primary hover:bg-primary/90 text-white text-sm font-bold transition-colors shadow-lg shadow-primary/25">
            <span className="truncate">
              {uiStrings?.landing?.trial_btn || 'Start Free Trial'}
            </span>
          </button>
        </div>
      </div>
    </header>
  );
}
