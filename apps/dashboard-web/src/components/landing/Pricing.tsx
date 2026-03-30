import { useState } from 'react';
import { SECTION_PADDING_Y_LARGE, LANDING_CONTENT_MAX_WIDTH, LANDING_PADDING_X, SECTION_TITLE_SIZE, SECTION_SUBTITLE_SIZE, GRID_COLS_FEATURES, CARD_PADDING, BORDER_LIGHT, BORDER_DARK, TRANSITION_BASE, BUTTON_PRIMARY, BUTTON_OUTLINE } from '../../constants/landingConstants';
import { useUIStrings } from '../../hooks/useUIStrings';

interface PricingTier {
  id: string;
  name: string;
  price: number;
  period: string;
  features: string[];
  cta: string;
  popular?: boolean;
  icon: string;
}

interface PricingProps {
  pricing: PricingTier[];
  landingMetadata?: Record<string, unknown>;
}

/**
 * Pricing Section Component
 * Displays pricing tiers with monthly/yearly toggle
 */
export function Pricing({ pricing, landingMetadata }: PricingProps) {
  const { uiStrings } = useUIStrings('landing');
  const [isYearly, setIsYearly] = useState(false);

  const togglePricing = () => setIsYearly((prev) => !prev);

  return (
    <section className={`${SECTION_PADDING_Y_LARGE} ${LANDING_PADDING_X}`} id="pricing">
      <div className={`${LANDING_CONTENT_MAX_WIDTH} mx-auto`}>
        <div className="text-center mb-16 max-w-3xl mx-auto">
          <h2 className={`${SECTION_TITLE_SIZE} font-bold text-slate-900 dark:text-white mb-4`}>
            Simple, Transparent Pricing
          </h2>
          <p className={`${SECTION_SUBTITLE_SIZE} text-slate-600 dark:text-text-secondary mb-8`}>
            Choose the plan that fits your fleet size and needs.
          </p>

          {/* Pricing Toggle */}
          <div className="flex items-center justify-center gap-4 mb-8">
            <span className={`text-sm font-bold ${!isYearly ? 'text-primary' : 'text-slate-500'}`}>
              Monthly
            </span>
            <button
              onClick={togglePricing}
              className="w-14 h-8 bg-slate-200 dark:bg-slate-700 rounded-full relative transition-colors"
              aria-label="Toggle pricing period"
            >
              <div
                className={`absolute top-1 w-6 h-6 bg-primary rounded-full transition-transform ${
                  isYearly ? 'translate-x-7' : 'translate-x-1'
                }`}
              />
            </button>
            <span className={`text-sm font-bold ${isYearly ? 'text-primary' : 'text-slate-500'}`}>
              Yearly <span className="text-green-500 text-xs">(Save 20%)</span>
            </span>
          </div>
        </div>

        <div className={`grid ${GRID_COLS_FEATURES} gap-8`}>
          {pricing.map((tier, idx) => (
            <div
              key={idx}
              className={`${CARD_PADDING} rounded-2xl border border-solid ${BORDER_LIGHT} dark:${BORDER_DARK} bg-white dark:bg-surface-dark ${TRANSITION_BASE} hover:shadow-xl relative ${
                tier.popular ? 'border-primary border-2' : ''
              }`}
            >
              {tier.popular && (
                <div className="absolute -top-4 left-1/2 transform -translate-x-1/2 bg-primary text-white text-xs font-bold px-4 py-1 rounded-full">
                  Most Popular
                </div>
              )}
              <div className="text-center mb-6">
                <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-2">
                  {tier.name}
                </h3>
                <div className="text-4xl font-black text-slate-900 dark:text-white mb-2">
                  ${isYearly ? (tier.price * 0.8).toFixed(0) : tier.price}
                  <span className="text-lg font-medium text-slate-500">/{tier.period}</span>
                </div>
              </div>
              <ul className="space-y-3 mb-8">
                {tier.features.map((feature, i) => (
                  <li key={i} className="flex items-center gap-2 text-sm text-slate-600 dark:text-text-secondary">
                    <span className="material-symbols-outlined text-green-500 text-sm">
                      check_circle
                    </span>
                    {feature}
                  </li>
                ))}
              </ul>
              <button className={`w-full ${BUTTON_PRIMARY} py-3 rounded-lg text-sm font-bold transition-all`}>
                {tier.cta}
              </button>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
