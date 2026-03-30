import { SECTION_PADDING_Y_LARGE, LANDING_CONTENT_MAX_WIDTH, LANDING_PADDING_X, SECTION_TITLE_SIZE, SECTION_SUBTITLE_SIZE, GRID_COLS_FEATURES, CARD_PADDING, BORDER_LIGHT, BORDER_DARK, TRANSITION_BASE, GRID_GAP } from '../../constants/landingConstants';
import { useUIStrings } from '../../hooks/useUIStrings';

interface Feature {
  id: string;
  title: string;
  desc: string;
  icon: string;
  color: string;
}

interface FeaturesProps {
  features: Feature[];
  landingMetadata?: Record<string, unknown>;
}

/**
 * Features Section Component
 * Displays platform features in a grid layout
 */
export function Features({ features, landingMetadata }: FeaturesProps) {
  const { uiStrings } = useUIStrings('landing');

  return (
    <section className={`${SECTION_PADDING_Y_LARGE} ${LANDING_PADDING_X}`} id="features">
      <div className={`${LANDING_CONTENT_MAX_WIDTH} mx-auto`}>
        <div className="text-center mb-16 max-w-3xl mx-auto">
          <h2 className={`${SECTION_TITLE_SIZE} font-bold text-slate-900 dark:text-white mb-4`}>
            Complete Fleet Control
          </h2>
          <p className={`${SECTION_SUBTITLE_SIZE} text-slate-600 dark:text-text-secondary`}>
            A comprehensive suite designed for data-driven decisions.
          </p>
        </div>
        <div className={`grid ${GRID_COLS_FEATURES} ${GRID_GAP}`}>
          {features.map((feature, idx) => (
            <div
              key={idx}
              className={`group ${CARD_PADDING} rounded-xl border border-solid ${BORDER_LIGHT} dark:${BORDER_DARK} bg-white dark:bg-surface-dark hover:border-primary/50 ${TRANSITION_BASE} hover:shadow-lg hover:shadow-primary/5`}
            >
              <div
                className={`w-12 h-12 rounded-lg bg-${feature.color}-500/10 flex items-center justify-center text-${feature.color}-500 mb-4`}
              >
                <span className="material-symbols-outlined !text-3xl">
                  {feature.icon}
                </span>
              </div>
              <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-2">
                {feature.title}
              </h3>
              <p className="text-slate-600 dark:text-text-secondary text-sm">
                {feature.desc}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
