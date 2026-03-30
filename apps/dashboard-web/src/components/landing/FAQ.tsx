import { useState } from 'react';
import { SECTION_PADDING_Y_LARGE, LANDING_CONTENT_MAX_WIDTH, LANDING_PADDING_X, SECTION_TITLE_SIZE, SECTION_SUBTITLE_SIZE, BORDER_LIGHT, BORDER_DARK, TRANSITION_BASE } from '../../constants/landingConstants';
import { useUIStrings } from '../../hooks/useUIStrings';

interface FAQ {
  id: string;
  question: string;
  answer: string;
}

interface FAQProps {
  faqs: FAQ[];
  landingMetadata?: Record<string, unknown>;
}

/**
 * FAQ Section Component
 * Displays frequently asked questions in an accordion
 */
export function FAQ({ faqs, landingMetadata }: FAQProps) {
  const { uiStrings } = useUIStrings('landing');
  const [openIndex, setOpenIndex] = useState<number | null>(null);

  const toggleFAQ = (index: number) => {
    setOpenIndex(openIndex === index ? null : index);
  };

  return (
    <section className={`${SECTION_PADDING_Y_LARGE} ${LANDING_PADDING_X}`} id="faq">
      <div className={`${LANDING_CONTENT_MAX_WIDTH} mx-auto`}>
        <div className="text-center mb-16 max-w-3xl mx-auto">
          <h2 className={`${SECTION_TITLE_SIZE} font-bold text-slate-900 dark:text-white mb-4`}>
            Frequently Asked Questions
          </h2>
          <p className={`${SECTION_SUBTITLE_SIZE} text-slate-600 dark:text-text-secondary`}>
            Everything you need to know about FlyFleet.
          </p>
        </div>
        <div className="max-w-3xl mx-auto space-y-4">
          {faqs.map((faq, index) => (
            <div
              key={index}
              className={`rounded-xl border border-solid ${BORDER_LIGHT} dark:${BORDER_DARK} bg-white dark:bg-surface-dark ${TRANSITION_BASE}`}
            >
              <button
                onClick={() => toggleFAQ(index)}
                className="w-full px-6 py-4 flex items-center justify-between text-left"
              >
                <span className="text-sm font-bold text-slate-900 dark:text-white">
                  {faq.question}
                </span>
                <span className="material-symbols-outlined text-slate-500 transition-transform">
                  {openIndex === index ? 'expand_less' : 'expand_more'}
                </span>
              </button>
              {openIndex === index && (
                <div className="px-6 pb-4 text-sm text-slate-600 dark:text-text-secondary">
                  {faq.answer}
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
