import { SECTION_PADDING_Y_LARGE, LANDING_CONTENT_MAX_WIDTH, LANDING_PADDING_X, SECTION_TITLE_SIZE, SECTION_SUBTITLE_SIZE, GRID_COLS_SUCCESS, CARD_PADDING, BORDER_LIGHT, BORDER_DARK, TRANSITION_BASE } from '../../constants/landingConstants';
import { useUIStrings } from '../../hooks/useUIStrings';

interface SuccessStory {
  id: string;
  title: string;
  value: string;
  quote: string;
  company: string;
  fleet: string;
  icon: string;
  color: string;
  badge: string;
  badgeColor: string;
}

interface SuccessStoriesProps {
  successStories: SuccessStory[];
  landingMetadata?: Record<string, unknown>;
}

/**
 * SuccessStories Section Component
 * Displays customer success stories and case studies
 */
export function SuccessStories({ successStories, landingMetadata }: SuccessStoriesProps) {
  const { uiStrings } = useUIStrings('landing');

  return (
    <section className={`${SECTION_PADDING_Y_LARGE} ${LANDING_PADDING_X}`} id="success-stories">
      <div className={`${LANDING_CONTENT_MAX_WIDTH} mx-auto`}>
        <div className="mb-16 md:text-center max-w-3xl mx-auto">
          <h2 className={`${SECTION_TITLE_SIZE} font-bold text-slate-900 dark:text-white mb-4`}>
            {uiStrings?.landing?.success_title ||
              'Commercial Success Stories'}
          </h2>
          <p className={`${SECTION_SUBTITLE_SIZE} text-slate-600 dark:text-text-secondary`}>
            {uiStrings?.landing?.success_sub ||
              'Real results from companies scaling their operations with FlyFleet.'}
          </p>
        </div>
        <div className={`grid ${GRID_COLS_SUCCESS} gap-8`}>
          {successStories.map((story, i) => (
            <div
              key={i}
              className={`relative group ${CARD_PADDING} rounded-2xl bg-white dark:bg-surface-dark border border-solid ${BORDER_LIGHT} dark:${BORDER_DARK} ${TRANSITION_BASE} hover:shadow-xl`}
            >
              <div className={`text-${story.color}-500 mb-6`}>
                <span className="material-symbols-outlined !text-5xl">
                  {story.icon}
                </span>
              </div>
              <div className="text-4xl md:text-5xl font-black text-slate-900 dark:text-white mb-2">
                {story.value}
              </div>
              <h3 className="text-lg font-bold text-slate-700 dark:text-slate-300 mb-6">
                {story.title}
              </h3>
              <p className="text-slate-600 dark:text-text-secondary mb-4 text-sm leading-relaxed">
                "{story.quote}"
              </p>
              <a
                className="inline-flex items-center gap-2 text-primary text-xs font-bold hover:underline mb-6"
                href="#"
              >
                <span className="material-symbols-outlined text-sm">
                  picture_as_pdf
                </span>
                Download PDF Case Study
              </a>
              <div className="flex items-center gap-3 border-t border-slate-100 dark:border-border-dark pt-6">
                <div
                  className={`w-10 h-10 rounded-full bg-${story.badgeColor}-100 text-${story.badgeColor}-600 flex items-center justify-center font-bold text-xs`}
                >
                  {story.badge}
                </div>
                <div>
                  <p className="text-sm font-bold text-slate-900 dark:text-white">
                    {story.company}
                  </p>
                  <p className="text-xs text-slate-500">{story.fleet}</p>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
