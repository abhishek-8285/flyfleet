import { SECTION_PADDING_Y_LARGE, LANDING_CONTENT_MAX_WIDTH, LANDING_PADDING_X, SECTION_TITLE_SIZE, SECTION_SUBTITLE_SIZE, GRID_COLS_FEATURES, CARD_PADDING, BORDER_LIGHT, BORDER_DARK, TRANSITION_BASE, GRID_GAP } from '../../constants/landingConstants';
import { useUIStrings } from '../../hooks/useUIStrings';

interface SecurityItem {
  id: string;
  title: string;
  desc: string;
  icon: string;
  color: string;
}

interface SecurityProps {
  securityItems: SecurityItem[];
  landingMetadata?: Record<string, unknown>;
}

/**
 * Security Section Component
 * Displays security features and compliance information
 */
export function Security({ securityItems, landingMetadata }: SecurityProps) {
  const { uiStrings } = useUIStrings('landing');

  return (
    <section className={`${SECTION_PADDING_Y_LARGE} ${LANDING_PADDING_X}`} id="security">
      <div className={`${LANDING_CONTENT_MAX_WIDTH} mx-auto`}>
        <div className="text-center mb-16 max-w-3xl mx-auto">
          <h2 className={`${SECTION_TITLE_SIZE} font-bold text-slate-900 dark:text-white mb-4`}>
            Enterprise-Grade Security
          </h2>
          <p className={`${SECTION_SUBTITLE_SIZE} text-slate-600 dark:text-text-secondary`}>
            Bank-level security to protect your fleet data.
          </p>
        </div>
        <div className={`grid ${GRID_COLS_FEATURES} ${GRID_GAP}`}>
          {securityItems.map((item, idx) => (
            <div
              key={idx}
              className={`${CARD_PADDING} rounded-xl border border-solid ${BORDER_LIGHT} dark:${BORDER_DARK} bg-white dark:bg-surface-dark ${TRANSITION_BASE} hover:shadow-lg`}
            >
              <div
                className={`w-12 h-12 rounded-lg bg-${item.color}-500/10 flex items-center justify-center text-${item.color}-500 mb-4`}
              >
                <span className="material-symbols-outlined !text-3xl">
                  {item.icon}
                </span>
              </div>
              <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-2">
                {item.title}
              </h3>
              <p className="text-slate-600 dark:text-text-secondary text-sm">
                {item.desc}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
