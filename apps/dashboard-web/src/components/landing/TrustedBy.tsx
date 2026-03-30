import { LANDING_CONTENT_MAX_WIDTH, LANDING_PADDING_X, BORDER_LIGHT, BORDER_DARK } from '../../constants/landingConstants';
import { useUIStrings } from '../../hooks/useUIStrings';

interface TrustedBrand {
  id: string;
  name: string;
  icon: string;
}

interface TrustedByProps {
  trustedBy: TrustedBrand[];
  landingMetadata?: Record<string, unknown>;
}

/**
 * TrustedBy Section Component
 * Displays brand logos of companies using the platform
 */
export function TrustedBy({ trustedBy, landingMetadata }: TrustedByProps) {
  const { uiStrings } = useUIStrings('landing');

  return (
    <section className={`border-y border-solid ${BORDER_LIGHT} dark:${BORDER_DARK} bg-slate-50 dark:bg-[#0d141c] py-10`}>
      <div className={`${LANDING_PADDING_X} ${LANDING_CONTENT_MAX_WIDTH} mx-auto`}>
        <p className="text-center text-xs font-bold uppercase tracking-widest text-slate-500 dark:text-slate-400 mb-8">
          {uiStrings?.landing?.trusted_title ||
            'Trusted by Global Logistics Leaders'}
        </p>
        <div className="flex flex-wrap justify-center items-center gap-12 md:gap-20 opacity-60 grayscale hover:grayscale-0 transition-all duration-500">
          {trustedBy.map((brand, idx) => (
            <div
              key={idx}
              className="flex items-center gap-2 font-black text-xl text-slate-800 dark:text-white"
            >
              <span className="material-symbols-outlined text-3xl">
                {brand.icon}
              </span>{' '}
              {brand.name}
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
