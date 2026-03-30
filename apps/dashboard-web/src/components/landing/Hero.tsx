import { HERO_PADDING_Y, LANDING_CONTENT_MAX_WIDTH, LANDING_PADDING_X, HERO_TITLE_SIZE, HERO_SUBTITLE_SIZE, BUTTON_HEIGHT_LARGE, BUTTON_PADDING_LARGE, BUTTON_PRIMARY, BUTTON_OUTLINE, ICONS, PRIMARY_SHADOW_LARGE } from '../../constants/landingConstants';
import { useUIStrings } from '../../hooks/useUIStrings';

interface HeroProps {
  landingMetadata?: Record<string, unknown> | null;
}

/**
 * Hero Section Component
 * Main hero section with title, subtitle, and CTA buttons
 */
export function Hero({ landingMetadata }: HeroProps) {
  const { uiStrings } = useUIStrings('landing');

  return (
    <section className={`relative w-full ${HERO_PADDING_Y} overflow-hidden mt-10`}>
      <div className="glow-effect"></div>
      <div className={`${LANDING_PADDING_X} ${LANDING_CONTENT_MAX_WIDTH} mx-auto flex flex-col items-center text-center relative z-10`}>
        {/* Security Badge */}
        <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-blue-500/10 border border-blue-500/20 text-blue-500 text-xs font-bold uppercase tracking-wide mb-6">
          <span className="relative flex h-2 w-2">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-500 opacity-75"></span>
            <span className="relative inline-flex rounded-full h-2 w-2 bg-blue-500"></span>
          </span>
          {uiStrings?.landing?.security_badge ||
            'Operations Grade Security & Reliability'}
        </div>

        {/* Hero Title */}
        <h1 className={`${HERO_TITLE_SIZE} font-black leading-tight tracking-tight mb-6 max-w-5xl text-slate-900 dark:text-white`}>
          {uiStrings?.landing?.hero_title ||
            'The Trust Standard for Commercial Fleets'}
        </h1>

        {/* Hero Subtitle */}
        <p className={`${HERO_SUBTITLE_SIZE} text-slate-600 dark:text-text-secondary max-w-2xl mb-10 leading-relaxed`}>
          {uiStrings?.landing?.hero_sub ||
            'Secure, scalable, and compliant fleet management.'}
        </p>

        {/* CTA Buttons */}
        <div className="flex flex-col sm:flex-row gap-4 mb-16">
          <button className={`${BUTTON_HEIGHT_LARGE} ${BUTTON_PADDING_LARGE} rounded-lg ${BUTTON_PRIMARY} transition-all ${PRIMARY_SHADOW_LARGE} flex items-center justify-center gap-2`}>
            {uiStrings?.landing?.demo_btn || 'Request Demo'}
            <span className="material-symbols-outlined !text-lg">
              {ICONS.ARROW_FORWARD}
            </span>
          </button>
          <button className={`${BUTTON_HEIGHT_LARGE} ${BUTTON_PADDING_LARGE} rounded-lg ${BUTTON_OUTLINE} transition-all flex items-center justify-center gap-2 group`}>
            <span className="material-symbols-outlined !text-lg text-primary group-hover:scale-110 transition-transform">
              {ICONS.PLAY_CIRCLE}
            </span>
            {uiStrings?.landing?.tour_btn || 'Watch Product Tour'}
          </button>
        </div>
      </div>
    </section>
  );
}
