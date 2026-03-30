import { SECTION_PADDING_Y_LARGE, LANDING_CONTENT_MAX_WIDTH, LANDING_PADDING_X, SECTION_TITLE_SIZE, BORDER_LIGHT, BORDER_DARK } from '../../constants/landingConstants';

interface Integration {
  id: string;
  name: string;
  icon: string;
  color: string;
}

interface IntegrationsSectionProps {
  integrations: Integration[];
  uiStrings?: Record<string, unknown>;
}

/**
 * Integrations Section Component
 * Displays third-party integrations grid
 */
export function IntegrationsSection({ integrations, uiStrings }: IntegrationsSectionProps) {
  if (!integrations || integrations.length === 0) return null;

  return (
    <section className={`${SECTION_PADDING_Y_LARGE} ${LANDING_PADDING_X} bg-slate-50 dark:bg-[#0d141c]`}>
      <div className={`${LANDING_CONTENT_MAX_WIDTH} mx-auto`}>
        <div className="flex flex-col md:flex-row justify-between items-end mb-12 gap-6">
          <div>
            <h2 className={`${SECTION_TITLE_SIZE} font-bold text-slate-900 dark:text-white mb-4`}>
              Deep Integrations
            </h2>
            <p className="text-slate-600 dark:text-text-secondary max-w-xl">
              Seamlessly connect FlyFleet into your existing tech stack.
            </p>
          </div>
          <a
            className="flex items-center gap-2 text-primary font-bold hover:underline"
            href="#"
          >
            View API Documentation{' '}
            <span className="material-symbols-outlined text-sm">
              arrow_forward
            </span>
          </a>
        </div>
        <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
          {integrations.map((item, index) => (
            <div
              key={index}
              className="bg-white dark:bg-surface-dark p-6 rounded-xl border border-slate-200 dark:border-border-dark flex flex-col items-center justify-center gap-3 hover:border-primary transition-colors group"
            >
              <span className={`material-symbols-outlined text-4xl text-slate-400 group-hover:text-${item.color}`}>
                {item.icon}
              </span>
              <span className="text-xs font-bold text-slate-700 dark:text-slate-300">
                {item.name}
              </span>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
