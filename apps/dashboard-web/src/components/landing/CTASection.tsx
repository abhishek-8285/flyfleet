import { SECTION_PADDING_Y_LARGE, LANDING_PADDING_X, LANDING_CONTENT_MAX_WIDTH } from '../../constants/landingConstants';

interface CTASectionProps {
  uiStrings?: Record<string, unknown>;
}

/**
 * CTA Section Component
 * Bottom call-to-action section
 */
export function CTASection({ uiStrings }: CTASectionProps) {
  return (
    <section className={`${SECTION_PADDING_Y_LARGE} ${LANDING_PADDING_X} bg-gradient-to-br from-primary to-primary/80 text-white`}>
      <div className={`${LANDING_CONTENT_MAX_WIDTH} mx-auto text-center`}>
        <h2 className="text-4xl md:text-5xl font-black mb-6">
          Ready to Transform Your Fleet?
        </h2>
        <p className="text-xl text-white/90 mb-10 max-w-2xl mx-auto">
          Join thousands of fleet managers who trust FlyFleet for their operations.
        </p>
        <div className="flex flex-col sm:flex-row gap-4 justify-center">
          <button className="h-12 px-8 rounded-lg bg-white text-primary hover:bg-slate-100 transition-colors text-base font-bold shadow-lg">
            Start Free Trial
          </button>
          <button className="h-12 px-8 rounded-lg border-2 border-white text-white hover:bg-white/10 transition-colors text-base font-bold">
            Schedule Demo
          </button>
        </div>
      </div>
    </section>
  );
}
