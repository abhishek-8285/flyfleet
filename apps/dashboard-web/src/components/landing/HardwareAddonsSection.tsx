import { SECTION_PADDING_Y_LARGE, LANDING_CONTENT_MAX_WIDTH, LANDING_PADDING_X, SECTION_TITLE_SIZE, BORDER_LIGHT, BORDER_DARK } from '../../constants/landingConstants';

interface HardwareAddon {
  id: string;
  name: string;
  price: string;
  icon: string;
  color: string;
  desc?: string;
}

interface HardwareAddonsSectionProps {
  hardwareAddons: HardwareAddon[];
  uiStrings?: Record<string, unknown>;
}

/**
 * Hardware Addons Section Component
 * Displays hardware products available for fleet
 */
export function HardwareAddonsSection({ hardwareAddons, uiStrings }: HardwareAddonsSectionProps) {
  if (!hardwareAddons || hardwareAddons.length === 0) return null;

  return (
    <section className={`${SECTION_PADDING_Y_LARGE} ${LANDING_PADDING_X}`}>
      <div className={`${LANDING_CONTENT_MAX_WIDTH} mx-auto`}>
        <div className="text-center mb-16">
          <h2 className={`${SECTION_TITLE_SIZE} font-bold text-slate-900 dark:text-white mb-4`}>
            Hardware Add-ons
          </h2>
          <p className="text-slate-600 dark:text-text-secondary max-w-2xl mx-auto">
            Enhance your fleet with our certified hardware partners.
          </p>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          {hardwareAddons.map((addon, i) => (
            <div
              key={i}
              className="rounded-2xl bg-white dark:bg-surface-dark border border-slate-200 dark:border-border-dark p-6 hover:shadow-xl transition-all"
            >
              <div className={`w-16 h-16 rounded-xl bg-${addon.color}-500/10 flex items-center justify-center text-${addon.color}-500 mb-4`}>
                <span className="material-symbols-outlined !text-4xl">
                  {addon.icon}
                </span>
              </div>
              <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-2">
                {addon.name}
              </h3>
              <p className="text-2xl font-black text-primary mb-4">{addon.price}</p>
              <p className="text-slate-600 dark:text-text-secondary text-sm mb-6">
                {addon.desc}
              </p>
              <button className="w-full bg-primary hover:bg-primary/90 text-white py-3 rounded-lg text-sm font-bold transition-colors">
                Add to Fleet
              </button>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
