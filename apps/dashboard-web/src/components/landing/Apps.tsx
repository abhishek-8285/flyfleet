import { SECTION_PADDING_Y_XLARGE, LANDING_CONTENT_MAX_WIDTH, LANDING_PADDING_X, SECTION_TITLE_SIZE, SECTION_SUBTITLE_SIZE, MOBILE_FRAME_WIDTH, MOBILE_FRAME_HEIGHT, MOBILE_FRAME_SHADOW, MOBILE_FRAME_BG, SURFACE_DARK, TRANSITION_TRANSFORM, ANIMATION_DELAY_100 } from '../../constants/landingConstants';
import { useUIStrings } from '../../hooks/useUIStrings';

interface AppsProps {
  landingMetadata?: Record<string, unknown> | null;
  systemMeta?: Record<string, unknown> | null;
  marketingMeta?: Record<string, unknown> | null;
}

interface UserMetadata {
  full_name?: string;
}

/**
 * Apps Section Component
 * Showcases the mobile app ecosystem with phone mockups
 */
export function Apps({ landingMetadata, systemMeta, marketingMeta }: AppsProps) {
  const { uiStrings } = useUIStrings('landing');
  
  const userName = (landingMetadata?.user_metadata as UserMetadata | undefined)?.full_name;

  return (
    <section
      className={`${SECTION_PADDING_Y_XLARGE} ${LANDING_PADDING_X} bg-slate-900 text-white relative overflow-hidden`}
      id="apps"
    >
      <div className="absolute inset-0 bg-blue-900/10 z-0"></div>
      <div className={`${LANDING_CONTENT_MAX_WIDTH} mx-auto relative z-10`}>
        <div className="text-center mb-20">
          <div className="inline-block px-3 py-1 bg-primary/20 text-primary rounded-full text-xs font-bold uppercase tracking-wide mb-4">
            Ecosystem
          </div>
          <h2 className={`${SECTION_TITLE_SIZE} md:text-5xl font-bold mb-6`}>
            Field-Ready Apps
          </h2>
          <p className={`${SECTION_SUBTITLE_SIZE} text-text-secondary max-w-2xl mx-auto`}>
            A complete ecosystem that connects your HQ, drivers, and clients
            in real-time. Native iOS and Android apps included with every
            plan.
          </p>
        </div>
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
          {/* Driver App Frame */}
          <div className="flex flex-col items-center">
            <div className={`${MOBILE_FRAME_WIDTH} ${MOBILE_FRAME_HEIGHT} ${MOBILE_FRAME_SHADOW} ${MOBILE_FRAME_BG} transform hover:-translate-y-2 ${TRANSITION_TRANSFORM} duration-500`}>
              <div className="h-full w-full ${MOBILE_FRAME_BG} flex flex-col pt-10 px-4 pb-4">
                <div className="flex justify-between items-center mb-6">
                  <div>
                    <div className="text-xs text-slate-400">
                      Good Morning,
                    </div>
                    <div className="font-bold text-lg">
                      {userName || '---'}
                    </div>
                  </div>
                  <div className="w-8 h-8 rounded-full bg-slate-700"></div>
                </div>
                <div className="bg-red-500/10 border border-red-500/30 rounded-xl p-4 mb-6 flex items-center gap-4 cursor-pointer hover:bg-red-500/20 transition-colors">
                  <div className="bg-red-500 p-3 rounded-full text-white shadow-lg shadow-red-500/40 animate-pulse">
                    <span className="material-symbols-outlined text-2xl">
                      sos
                    </span>
                  </div>
                  <div>
                    <div className="font-bold text-red-500">
                      Emergency SOS
                    </div>
                    <div className="text-[10px] text-red-400">
                      Press and hold for 3s
                    </div>
                  </div>
                </div>
                <div className="mb-2 flex justify-between items-end">
                  <h4 className="font-bold text-sm">
                    {uiStrings?.landing?.tasks_title || "Today's Tasks"}
                  </h4>
                  <span className="text-[10px] text-primary">
                    {uiStrings?.landing?.tasks_view_all || 'View All'}
                  </span>
                </div>
                <div className="space-y-3 overflow-y-auto no-scrollbar flex-1">
                  {/* Task items would go here */}
                </div>
              </div>
            </div>
            <div className="mt-8 text-center">
              <h3 className="text-xl font-bold mb-2">
                {uiStrings?.landing?.driver_app_name ||
                  'Driver Companion App'}
              </h3>
              <p className="text-text-secondary text-sm max-w-xs mx-auto">
                {uiStrings?.landing?.driver_app_desc ||
                  'Task management, SOS alerts, and vehicle inspection forms.'}
              </p>
            </div>
          </div>

          {/* Manager App Frame */}
          <div className="flex flex-col items-center">
            <div className={`${MOBILE_FRAME_WIDTH} ${MOBILE_FRAME_HEIGHT} ${MOBILE_FRAME_SHADOW} ${MOBILE_FRAME_BG} transform hover:-translate-y-2 ${TRANSITION_TRANSFORM} duration-500 ${ANIMATION_DELAY_100} border-primary/30`}>
              {/* Manager app content */}
            </div>
            <div className="mt-8 text-center">
              <h3 className="text-xl font-bold mb-2">
                {uiStrings?.landing?.manager_app_title ||
                  'Manager Track App'}
              </h3>
              <p className="text-text-secondary text-sm max-w-xs mx-auto">
                {uiStrings?.landing?.manager_app_desc ||
                  'Real-time fleet visibility and instant alerts from your pocket.'}
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
