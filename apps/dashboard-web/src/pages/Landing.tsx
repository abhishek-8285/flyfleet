import branding from "@flyfleet/config";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useUIStrings } from "../hooks/useUIStrings";

const Landing = () => {
  const navigate = useNavigate();
  const [isYearly, setIsYearly] = useState(false);
  const [isDark, setIsDark] = useState(false);
  const [SUCCESS_STORIES, setSuccessStories] = useState<any[]>([]);
  const [FEATURES, setFeatures] = useState<any[]>([]);
  const [INTEGRATIONS, setIntegrations] = useState<any[]>([]);
  const [SECURITY_ITEMS, setSecurityItems] = useState<any[]>([]);
  const [HARDWARE_ADDONS, setHardwareAddons] = useState<any[]>([]);
  const [FAQS, setFaqs] = useState<any[]>([]);
  const [TRUSTED_BY, setTrustedBy] = useState<any[]>([]);
  const [TIMELINE, setTimeline] = useState<any[]>([]);
  const [PRICING, setPricing] = useState<any[]>([]);
  const [landingMetadata, setLandingMetadata] = useState<any>(null);
  const [systemMeta, setSystemMeta] = useState<any>(null);
  const [marketingMeta, setMarketingMeta] = useState<any>(null);
  const { uiStrings } = useUIStrings("landing");

  useEffect(() => {
    fetch("/api/v1/data")
      .then((res) => res.json())
      .then((data) => {
        if (data.successStories) setSuccessStories(data.successStories);
        if (data.features) setFeatures(data.features);
        if (data.integrations) setIntegrations(data.integrations);
        if (data.securityItems) setSecurityItems(data.securityItems);
        if (data.hardwareAddons) setHardwareAddons(data.hardwareAddons);
        if (data.faqs) setFaqs(data.faqs);
        if (data.trustedBy) setTrustedBy(data.trustedBy);
        if (data.timeline) setTimeline(data.timeline);
        if (data.pricing) setPricing(data.pricing);
      })
      .catch((err) => console.error(err));

    fetch("/api/v1/ui/global_config")
      .then((res) => res.json())
      .then((data) => {
        setLandingMetadata(data.landing_metadata || data);
        setSystemMeta(data.system_metadata);
        setMarketingMeta(data.marketing_metadata);
      })
      .catch((err) => console.error(err));
  }, []);

  useEffect(() => {
    if (isDark) {
      document.documentElement.classList.add("dark");
    } else {
      document.documentElement.classList.remove("dark");
    }
  }, [isDark]);

  const toggleTheme = () => setIsDark(!isDark);

  return (
    <div
      className={`min-h-screen font-display bg-background-light dark:bg-background-dark text-slate-900 dark:text-white overflow-x-hidden selection:bg-primary selection:text-white ${isDark ? "dark" : ""}`}
    >
      {/* Top Info Bar */}
      <div className="fixed top-0 w-full z-[60] bg-[#0f65bd] text-white text-sm font-medium h-10 flex items-center justify-center shadow-md">
        <div className="max-w-[1440px] w-full px-4 md:px-10 flex justify-between items-center">
          <div className="flex items-center gap-2 text-xs md:text-sm opacity-90">
            <span className="material-symbols-outlined !text-lg">
              verified_user
            </span>
            <span>
              {uiStrings?.landing?.security_text || "Operations Grade Security"}
            </span>
          </div>
          <div className="flex items-center gap-6">
            <a
              className="flex items-center gap-2 hover:text-white text-blue-100 transition-colors font-bold group"
              href={`tel:${landingMetadata?.sales_info?.phone || "+15551234567"}`}
            >
              <span className="material-symbols-outlined !text-lg group-hover:animate-pulse">
                call
              </span>
              {uiStrings?.landing?.sales_btn || "Talk to Sales"}:{" "}
              {landingMetadata?.sales_info?.phone || "+--"}
            </a>
            <div className="h-4 w-px bg-white/20 hidden md:block"></div>
            <div className="hidden md:flex items-center gap-1 cursor-pointer hover:text-white text-blue-100 group relative">
              <span className="material-symbols-outlined !text-lg">
                language
              </span>
              <span className="text-xs font-bold">
                {uiStrings?.landing?.locale_en || "US (EN)"}
              </span>
              <span className="material-symbols-outlined !text-sm">
                expand_more
              </span>
              <div className="absolute top-full right-0 mt-2 w-32 bg-white dark:bg-surface-dark rounded-lg shadow-xl border border-slate-200 dark:border-border-dark py-2 hidden group-hover:block text-slate-900 dark:text-white z-50">
                <a
                  className="block px-4 py-2 hover:bg-slate-50 dark:hover:bg-background-dark text-xs"
                  href="#"
                >
                  Global
                </a>
                <a
                  className="block px-4 py-2 hover:bg-slate-50 dark:hover:bg-background-dark text-xs"
                  href="#"
                >
                  Europe (EU)
                </a>
                <a
                  className="block px-4 py-2 hover:bg-slate-50 dark:hover:bg-background-dark text-xs font-bold text-primary"
                  href="#"
                >
                  United States
                </a>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Header */}
      <header className="fixed top-10 z-50 w-full border-b border-solid border-slate-200 dark:border-border-dark bg-white/80 dark:bg-[#111a22]/80 backdrop-blur-md transition-all">
        <div className="px-4 md:px-10 flex h-16 items-center justify-between max-w-[1440px] mx-auto">
          <div className="flex items-center gap-4 text-slate-900 dark:text-white">
            <div className="size-8 flex items-center justify-center text-primary">
              <span className="material-symbols-outlined !text-3xl">
                local_shipping
              </span>
            </div>
            <h2 className="text-lg font-bold leading-tight tracking-[-0.015em]">
              FlyFleet
            </h2>
          </div>
          <nav className="hidden md:flex flex-1 justify-center gap-8">
            <a
              className="text-slate-600 dark:text-text-secondary text-sm font-medium hover:text-primary transition-colors"
              href="#success-stories"
            >
              {uiStrings?.landing?.success_stories || "Success Stories"}
            </a>
            <a
              className="text-slate-600 dark:text-text-secondary text-sm font-medium hover:text-primary transition-colors"
              href="#features"
            >
              {uiStrings?.landing?.platform || "Platform"}
            </a>
            <a
              className="text-slate-600 dark:text-text-secondary text-sm font-medium hover:text-primary transition-colors"
              href="#apps"
            >
              {uiStrings?.landing?.apps || "Apps"}
            </a>
            <a
              className="text-slate-600 dark:text-text-secondary text-sm font-medium hover:text-primary transition-colors"
              href="#pricing"
            >
              {uiStrings?.landing?.pricing || "Pricing"}
            </a>
            <a
              className="text-slate-600 dark:text-text-secondary text-sm font-medium hover:text-primary transition-colors"
              href="#security"
            >
              {uiStrings?.landing?.security || "Security"}
            </a>
          </nav>
          <div className="flex gap-4 items-center">
            <button
              onClick={toggleTheme}
              className="p-2 rounded-full hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-600 dark:text-slate-300 transition-colors"
            >
              <span className="material-symbols-outlined">
                {isDark ? "light_mode" : "dark_mode"}
              </span>
            </button>
            <button
              onClick={() => navigate("/login")}
              className="hidden sm:flex h-9 items-center justify-center rounded-lg px-4 bg-transparent border border-slate-200 dark:border-border-dark text-slate-900 dark:text-white text-sm font-bold hover:bg-slate-100 dark:hover:bg-surface-dark transition-colors"
            >
              <span className="truncate">
                {uiStrings?.landing?.login_btn || "Login"}
              </span>
            </button>
            <button className="flex h-9 items-center justify-center rounded-lg px-4 bg-primary hover:bg-primary/90 text-white text-sm font-bold transition-colors shadow-lg shadow-primary/25">
              <span className="truncate">
                {uiStrings?.landing?.trial_btn || "Start Free Trial"}
              </span>
            </button>
          </div>
        </div>
      </header>

      <main className="flex min-h-screen w-full flex-col pt-26">
        {/* Hero */}
        <section className="relative w-full py-16 md:py-24 overflow-hidden mt-10">
          <div className="glow-effect"></div>
          <div className="px-4 md:px-10 max-w-[1280px] mx-auto flex flex-col items-center text-center relative z-10">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-blue-500/10 border border-blue-500/20 text-blue-500 text-xs font-bold uppercase tracking-wide mb-6">
              <span className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-500 opacity-75"></span>
                <span className="relative inline-flex rounded-full h-2 w-2 bg-blue-500"></span>
              </span>
              {uiStrings?.landing?.security_badge ||
                "Operations Grade Security & Reliability"}
            </div>
            <h1 className="text-4xl md:text-6xl lg:text-7xl font-black leading-tight tracking-tight mb-6 max-w-5xl text-slate-900 dark:text-white">
              {uiStrings?.landing?.hero_title ||
                "The Trust Standard for Commercial Fleets"}
            </h1>
            <p className="text-slate-600 dark:text-text-secondary text-lg md:text-xl max-w-2xl mb-10 leading-relaxed">
              {uiStrings?.landing?.hero_sub ||
                "Secure, scalable, and compliant fleet management."}
            </p>
            <div className="flex flex-col sm:flex-row gap-4 mb-16">
              <button className="h-12 px-8 rounded-lg bg-primary hover:bg-primary/90 text-white text-base font-bold transition-all shadow-lg shadow-primary/30 flex items-center justify-center gap-2">
                {uiStrings?.landing?.demo_btn || "Request Demo"}
                <span className="material-symbols-outlined !text-lg">
                  arrow_forward
                </span>
              </button>
              <button className="h-12 px-8 rounded-lg border border-slate-300 dark:border-border-dark bg-white dark:bg-surface-dark hover:bg-slate-50 dark:hover:bg-border-dark text-slate-900 dark:text-white text-base font-bold transition-all flex items-center justify-center gap-2 group">
                <span className="material-symbols-outlined !text-lg text-primary group-hover:scale-110 transition-transform">
                  play_circle
                </span>
                {uiStrings?.landing?.tour_btn || "Watch Product Tour"}
              </button>
            </div>
          </div>
        </section>

        {/* Trusted By */}
        <section className="border-y border-slate-200 dark:border-border-dark bg-slate-50 dark:bg-[#0d141c] py-10">
          <div className="max-w-[1280px] mx-auto px-4 md:px-10">
            <p className="text-center text-xs font-bold uppercase tracking-widest text-slate-500 dark:text-slate-400 mb-8">
              {uiStrings?.landing?.trusted_title ||
                "Trusted by Global Logistics Leaders"}
            </p>
            <div className="flex flex-wrap justify-center items-center gap-12 md:gap-20 opacity-60 grayscale hover:grayscale-0 transition-all duration-500">
              {TRUSTED_BY.map((brand, idx) => (
                <div
                  key={idx}
                  className="flex items-center gap-2 font-black text-xl text-slate-800 dark:text-white"
                >
                  <span className="material-symbols-outlined text-3xl">
                    {brand.icon}
                  </span>{" "}
                  {brand.name}
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* Success Stories */}
        <section className="py-20 px-4 md:px-10" id="success-stories">
          <div className="max-w-[1280px] mx-auto">
            <div className="mb-16 md:text-center max-w-3xl mx-auto">
              <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4">
                {uiStrings?.landing?.success_title ||
                  "Commercial Success Stories"}
              </h2>
              <p className="text-slate-600 dark:text-text-secondary text-lg">
                {uiStrings?.landing?.success_sub ||
                  "Real results from companies scaling their operations with FlyFleet."}
              </p>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
              {SUCCESS_STORIES.map((story, i) => (
                <div
                  key={i}
                  className="relative group rounded-2xl bg-white dark:bg-surface-dark border border-slate-200 dark:border-border-dark p-8 shadow-sm hover:shadow-xl transition-all"
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

        {/* Apps Showcase */}
        <section
          className="py-24 px-4 md:px-10 bg-slate-900 text-white relative overflow-hidden"
          id="apps"
        >
          <div className="absolute inset-0 bg-blue-900/10 z-0"></div>
          <div className="max-w-[1280px] mx-auto relative z-10">
            <div className="text-center mb-20">
              <div className="inline-block px-3 py-1 bg-primary/20 text-primary rounded-full text-xs font-bold uppercase tracking-wide mb-4">
                Ecosystem
              </div>
              <h2 className="text-3xl md:text-5xl font-bold mb-6">
                Field-Ready Apps
              </h2>
              <p className="text-text-secondary text-lg max-w-2xl mx-auto">
                A complete ecosystem that connects your HQ, drivers, and clients
                in real-time. Native iOS and Android apps included with every
                plan.
              </p>
            </div>
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
              {/* Driver App Frame */}
              <div className="flex flex-col items-center">
                <div className="mobile-frame w-[280px] h-[580px] shadow-2xl bg-surface-dark transform hover:-translate-y-2 transition-transform duration-500">
                  <div className="h-full w-full bg-[#111a22] flex flex-col pt-10 px-4 pb-4">
                    <div className="flex justify-between items-center mb-6">
                      <div>
                        <div className="text-xs text-slate-400">
                          Good Morning,
                        </div>
                        <div className="font-bold text-lg">
                          {landingMetadata?.user_metadata?.full_name || "---"}
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
                        {uiStrings?.landing?.tasks_view_all || "View All"}
                      </span>
                    </div>
                    <div className="space-y-3 overflow-y-auto no-scrollbar flex-1">
                      <div className="bg-surface-dark p-3 rounded-xl border border-border-dark flex gap-3">
                        <div className="mt-1">
                          <span className="material-symbols-outlined text-green-500 text-sm">
                            radio_button_checked
                          </span>
                        </div>
                        <div>
                          <div className="text-xs font-bold">
                            Delivery #4922
                          </div>
                          <div className="text-[10px] text-slate-400 mb-1">
                            {systemMeta?.location || "---"}
                          </div>
                          <span className="text-[9px] bg-green-500/20 text-green-500 px-1.5 py-0.5 rounded">
                            {uiStrings?.landing?.task_completed || "Completed"}
                          </span>
                        </div>
                      </div>
                      <div className="bg-surface-dark p-3 rounded-xl border border-primary/30 flex gap-3 relative overflow-hidden">
                        <div className="absolute left-0 top-0 bottom-0 w-1 bg-primary"></div>
                        <div className="mt-1">
                          <span className="material-symbols-outlined text-primary text-sm">
                            radio_button_unchecked
                          </span>
                        </div>
                        <div>
                          <div className="text-xs font-bold text-white">
                            Pickup #5011
                          </div>
                          <div className="text-[10px] text-slate-400 mb-2">
                            {systemMeta?.location || "---"}
                          </div>
                          <div className="flex gap-2">
                            <button className="bg-primary text-white text-[9px] font-bold px-3 py-1 rounded">
                              {uiStrings?.landing?.navigate_btn || "Navigate"}
                            </button>
                            <button className="bg-slate-700 text-white text-[9px] font-bold px-3 py-1 rounded">
                              {uiStrings?.landing?.details_btn || "Details"}
                            </button>
                          </div>
                        </div>
                      </div>
                    </div>
                    <div className="border-t border-border-dark pt-4 mt-2 flex justify-around text-slate-500">
                      <span className="material-symbols-outlined text-primary">
                        home
                      </span>
                      <span className="material-symbols-outlined">map</span>
                      <span className="material-symbols-outlined">person</span>
                    </div>
                  </div>
                </div>
                <div className="mt-8 text-center">
                  <h3 className="text-xl font-bold mb-2">
                    {uiStrings?.landing?.driver_app_name ||
                      "Driver Companion App"}
                  </h3>
                  <p className="text-text-secondary text-sm max-w-xs mx-auto">
                    {uiStrings?.landing?.driver_app_desc ||
                      "Task management, SOS alerts, and vehicle inspection forms."}
                  </p>
                </div>
              </div>
              {/* Manager App Frame */}
              <div className="flex flex-col items-center">
                <div className="mobile-frame w-[280px] h-[580px] shadow-2xl bg-surface-dark transform hover:-translate-y-2 transition-transform duration-500 delay-100 border-primary/30">
                  <div className="h-full w-full bg-[#111a22] flex flex-col pt-0 relative">
                    <div
                      className="absolute inset-0 bg-slate-800 opacity-60 z-0 bg-cover grayscale opacity-30"
                      style={{
                        backgroundImage: "url('/dashboard-preview.png')",
                      }}
                    ></div>
                    <div className="absolute top-1/3 left-1/4 z-10">
                      <div className="w-12 h-12 bg-primary/20 rounded-full flex items-center justify-center animate-pulse">
                        <div className="w-3 h-3 bg-primary rounded-full ring-2 ring-white"></div>
                      </div>
                    </div>
                    <div className="absolute top-1/2 left-2/3 z-10">
                      <div className="w-3 h-3 bg-green-500 rounded-full ring-2 ring-white shadow-lg"></div>
                    </div>
                    <div className="absolute top-12 left-4 right-4 bg-surface-dark/90 backdrop-blur p-3 rounded-xl border border-border-dark shadow-lg z-20">
                      <div className="flex gap-2 items-center text-xs text-slate-400 mb-1">
                        <span className="material-symbols-outlined text-sm">
                          search
                        </span>{" "}
                        {uiStrings?.common?.search_placeholder ||
                          "Search Fleet..."}
                      </div>
                    </div>
                    <div className="absolute bottom-0 w-full bg-surface-dark rounded-t-2xl z-20 p-5 pb-8 shadow-[0_-10px_40px_rgba(0,0,0,0.5)]">
                      <div className="w-12 h-1 bg-slate-700 rounded-full mx-auto mb-4"></div>
                      <div className="flex justify-between items-start mb-4">
                        <div>
                          <div className="text-[10px] font-bold text-primary uppercase tracking-wider mb-1">
                            Vehicle Status
                          </div>
                          <div className="text-lg font-bold text-white">
                            Truck TR-401
                          </div>
                          <div className="text-xs text-text-secondary">
                            55 mph • I-95 North
                          </div>
                        </div>
                        <div className="bg-green-500/20 text-green-500 p-2 rounded-lg">
                          <span className="material-symbols-outlined text-xl">
                            local_shipping
                          </span>
                        </div>
                      </div>
                      <div className="grid grid-cols-3 gap-2 mb-4">
                        <div className="bg-slate-800 p-2 rounded text-center">
                          <div className="text-[9px] text-slate-400">
                            {uiStrings?.landing?.fuel_stat || "Fuel"}
                          </div>
                          <div className="text-xs font-bold text-white">
                            {marketingMeta?.conversion_rate || "--%"}
                          </div>
                        </div>
                        <div className="bg-slate-800 p-2 rounded text-center">
                          <div className="text-[9px] text-slate-400">
                            {uiStrings?.landing?.temp_stat || "Temp"}
                          </div>
                          <div className="text-xs font-bold text-white">
                            {marketingMeta?.current_temp || "--°C"}
                          </div>
                        </div>
                        <div className="bg-slate-800 p-2 rounded text-center">
                          <div className="text-[9px] text-slate-400">
                            Driver
                          </div>
                          <div className="text-xs font-bold text-white">
                            {marketingMeta?.driver_name || "---"}
                          </div>
                        </div>
                      </div>
                      <button className="w-full bg-primary py-3 rounded-lg text-xs font-bold text-white shadow-lg shadow-primary/20">
                        {uiStrings?.landing?.track_live_btn || "Track Live"}
                      </button>
                    </div>
                  </div>
                </div>
                <div className="mt-8 text-center">
                  <h3 className="text-xl font-bold mb-2">
                    {uiStrings?.landing?.manager_app_title ||
                      "Manager Track App"}
                  </h3>
                  <p className="text-text-secondary text-sm max-w-xs mx-auto">
                    {uiStrings?.landing?.manager_app_desc ||
                      "Real-time fleet visibility and instant alerts from your pocket."}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* Features Grid */}
        <section className="py-20 px-4 md:px-10" id="features">
          <div className="max-w-[1280px] mx-auto">
            <div className="text-center mb-16 max-w-3xl mx-auto">
              <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4">
                Complete Fleet Control
              </h2>
              <p className="text-slate-600 dark:text-text-secondary text-lg">
                A comprehensive suite designed for data-driven decisions.
              </p>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {FEATURES.map((feature, idx) => (
                <div
                  key={idx}
                  className="group p-6 rounded-xl border border-slate-200 dark:border-border-dark bg-white dark:bg-surface-dark hover:border-primary/50 transition-all hover:shadow-lg hover:shadow-primary/5"
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

        {/* Deep Integrations */}
        <section className="py-20 px-4 md:px-10 bg-slate-50 dark:bg-[#0d141c]">
          <div className="max-w-[1280px] mx-auto">
            <div className="flex flex-col md:flex-row justify-between items-end mb-12 gap-6">
              <div>
                <h2 className="text-3xl font-bold text-slate-900 dark:text-white mb-4">
                  Deep Integrations
                </h2>
                <p className="text-slate-600 dark:text-text-secondary max-w-xl">
                  Seamlessly connect {branding.companyName} into your existing
                  tech stack. We support major ERPs, fuel card providers, and
                  messaging platforms.
                </p>
              </div>
              <a
                className="flex items-center gap-2 text-primary font-bold hover:underline"
                href="#"
              >
                View API Documentation{" "}
                <span className="material-symbols-outlined text-sm">
                  arrow_forward
                </span>
              </a>
            </div>
            <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
              {INTEGRATIONS.map((item, index) => (
                <div
                  key={index}
                  className="bg-white dark:bg-surface-dark p-6 rounded-xl border border-slate-200 dark:border-border-dark flex flex-col items-center justify-center gap-3 hover:border-primary transition-colors group"
                >
                  <span
                    className={`material-symbols-outlined text-4xl text-slate-400 group-hover:${item.color}`}
                  >
                    {item.icon}
                  </span>
                  <span className="font-bold text-sm text-slate-700 dark:text-slate-300">
                    {item.label}
                  </span>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* Security */}
        <section className="py-20 px-4 md:px-10" id="security">
          <div className="max-w-[1280px] mx-auto bg-slate-900 rounded-3xl p-8 md:p-16 border border-border-dark relative overflow-hidden">
            <div className="absolute top-0 right-0 w-96 h-96 bg-primary/10 rounded-full blur-3xl -translate-y-1/2 translate-x-1/2 pointer-events-none"></div>
            <div className="relative z-10 text-center mb-16">
              <h2 className="text-3xl md:text-4xl font-bold text-white mb-6">
                {uiStrings?.landing?.security_title ||
                  "Operations-Grade Security"}
              </h2>
              <p className="text-text-secondary max-w-2xl mx-auto text-lg">
                {uiStrings?.landing?.security_sub ||
                  "We maintain the highest standards of data protection and reliability."}
              </p>
            </div>
            <div className="relative z-10 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-8">
              {SECURITY_ITEMS.map((item, index) => (
                <div
                  key={index}
                  className="flex flex-col items-center text-center p-6 rounded-xl bg-surface-dark/50 border border-border-dark backdrop-blur-sm"
                >
                  <div
                    className={`w-16 h-16 rounded-full bg-${item.color}-500/10 flex items-center justify-center text-${item.color}-400 mb-4 border border-${item.color}-500/20`}
                  >
                    <span className="material-symbols-outlined !text-4xl">
                      {item.icon}
                    </span>
                  </div>
                  <h3 className="text-white font-bold text-lg mb-2">
                    {item.label}
                  </h3>
                  <p className="text-sm text-text-secondary">{item.desc}</p>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* Implementation Timeline */}
        <section
          className="py-20 px-4 md:px-10 bg-slate-50 dark:bg-[#0d141c]"
          id="how-it-works"
        >
          <div className="max-w-[1280px] mx-auto">
            <div className="text-center mb-16">
              <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4">
                {uiStrings?.landing?.implementation_title ||
                  "Implementation in 3 Weeks"}
              </h2>
              <p className="text-slate-600 dark:text-text-secondary text-lg">
                {uiStrings?.landing?.implementation_sub ||
                  "Rapid deployment with full data migration support."}
              </p>
            </div>
            {/* Timeline Line */}
            <div className="relative max-w-5xl mx-auto mb-16 hidden md:block">
              <div className="absolute top-1/2 left-0 w-full h-1 bg-slate-200 dark:bg-border-dark -translate-y-1/2 z-0 rounded-full"></div>
              <div className="relative z-10 flex justify-between w-full">
                {TIMELINE.map((step, idx) => (
                  <div
                    key={idx}
                    className={`bg-white dark:bg-surface-dark border-4 ${idx === TIMELINE.length - 1 ? "border-primary bg-primary text-white" : "border-slate-200 dark:border-border-dark text-slate-900 dark:text-white"} p-1 rounded-full w-12 h-12 flex items-center justify-center font-bold shadow-lg`}
                  >
                    {idx + 1}
                  </div>
                ))}
              </div>
              <div className="flex justify-between w-full mt-4 text-center">
                {TIMELINE.map((step, idx) => (
                  <div
                    key={idx}
                    className={`w-48 ${idx === 0 ? "-ml-16" : idx === TIMELINE.length - 1 ? "-mr-16" : idx === 1 ? "-ml-8" : "-mr-8"}`}
                  >
                    <h4
                      className={`font-bold ${idx === TIMELINE.length - 1 ? "text-primary" : "text-slate-900 dark:text-white"}`}
                    >
                      {step.title}
                    </h4>
                    <p className="text-sm text-slate-500">{step.desc}</p>
                  </div>
                ))}
              </div>
            </div>

            {/* Hardware Options Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-12 items-center">
              <div className="bg-white dark:bg-surface-dark p-8 rounded-2xl border border-slate-200 dark:border-border-dark shadow-lg">
                <h3 className="text-2xl font-bold text-slate-900 dark:text-white mb-4">
                  {uiStrings?.landing?.hardware_title || "Hardware Options"}
                </h3>
                <ul className="space-y-4">
                  <li className="flex items-start gap-4">
                    <span className="material-symbols-outlined text-primary mt-1">
                      settings_input_component
                    </span>
                    <div>
                      <h4 className="font-bold text-slate-900 dark:text-white">
                        Plug & Play OBD-II
                      </h4>
                      <p className="text-sm text-slate-600 dark:text-text-secondary">
                        Installs in 30 seconds. No mechanic required. Perfect
                        for light commercial vehicles.
                      </p>
                    </div>
                  </li>
                  <li className="flex items-start gap-4">
                    <span className="material-symbols-outlined text-primary mt-1">
                      cable
                    </span>
                    <div>
                      <h4 className="font-bold text-slate-900 dark:text-white">
                        Wired Installation
                      </h4>
                      <p className="text-sm text-slate-600 dark:text-text-secondary">
                        Hidden installation for theft prevention. Supports
                        remote immobilization and sensor integration.
                      </p>
                    </div>
                  </li>
                </ul>
              </div>
              <div className="relative bg-slate-900 rounded-2xl p-8 border border-border-dark aspect-video flex items-center justify-center overflow-hidden">
                <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,_var(--tw-gradient-stops))] from-blue-900/40 via-slate-900 to-slate-900"></div>
                <div className="relative z-10 text-center">
                  <span className="material-symbols-outlined !text-8xl text-slate-700">
                    router
                  </span>
                  {uiStrings?.landing?.obd_viz_placeholder ||
                    "OBD-II DEVICE VISUALIZATION"}
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* Demo Scheduler */}
        <section
          className="py-20 px-4 md:px-10 bg-white dark:bg-surface-dark border-t border-slate-200 dark:border-border-dark"
          id="demo-schedule"
        >
          <div className="max-w-[1280px] mx-auto grid lg:grid-cols-2 gap-16 items-center">
            <div>
              <div className="inline-block px-3 py-1 bg-blue-100 dark:bg-blue-900/30 text-primary rounded-full text-xs font-bold uppercase tracking-wide mb-4">
                Live Walkthrough
              </div>
              <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-6">
                {uiStrings?.landing?.demo_title || "See the platform in action"}
              </h2>
              <ul className="space-y-4 mb-8">
                {(
                  landingMetadata?.demo_features || [
                    "Custom ROI analysis for your fleet size",
                    "Deep dive into API & Integrations",
                    "Hardware installation preview",
                  ]
                ).map((item: string, i: number) => (
                  <li
                    key={i}
                    className="flex items-center gap-3 text-slate-700 dark:text-slate-300"
                  >
                    <span className="material-symbols-outlined text-green-500">
                      check_circle
                    </span>{" "}
                    {item}
                  </li>
                ))}
              </ul>
            </div>

            {/* Calendar Widget */}
            <div className="bg-slate-50 dark:bg-[#0d141c] rounded-2xl border border-slate-200 dark:border-border-dark p-6 shadow-xl">
              <div className="flex flex-col md:flex-row gap-6">
                <div className="flex-1">
                  <div className="text-sm font-bold text-slate-900 dark:text-white mb-4 text-center">
                    {uiStrings?.landing?.calendar_select_date ||
                      "Select a Date"}
                  </div>
                  <div className="grid grid-cols-7 gap-1 text-center text-xs mb-2 text-slate-400">
                    <span>S</span>
                    <span>M</span>
                    <span>T</span>
                    <span>W</span>
                    <span>T</span>
                    <span>F</span>
                    <span>S</span>
                  </div>
                  <div className="grid grid-cols-7 gap-1 text-center text-sm">
                    {(
                      landingMetadata?.demo_calendar?.available_days || [
                        29, 30, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
                      ]
                    ).map((day: number, i: number) => (
                      <span
                        key={i}
                        className={`p-2 rounded-lg ${day === (landingMetadata?.demo_calendar?.selected_day || 2) ? "bg-primary text-white shadow-lg shadow-primary/30" : "text-slate-400 hover:bg-white dark:hover:bg-surface-dark cursor-pointer"}`}
                      >
                        {day}
                      </span>
                    ))}
                  </div>
                </div>
                <div className="w-px bg-slate-200 dark:bg-border-dark hidden md:block"></div>
                <div className="flex-1">
                  <div className="text-sm font-bold text-slate-900 dark:text-white mb-4 text-center">
                    {uiStrings?.landing?.calendar_available_times ||
                      "Available Times"}
                  </div>
                  <div className="space-y-2">
                    {(
                      landingMetadata?.demo_calendar?.available_times || [
                        "9:00 AM",
                        "9:30 AM",
                      ]
                    ).map((time: string, idx: number) => (
                      <button
                        key={idx}
                        className={`w-full py-2 px-4 rounded-lg border ${idx === 0 ? "border-primary text-primary hover:bg-primary hover:text-white" : "border-slate-200 dark:border-border-dark text-slate-600 dark:text-text-secondary hover:border-primary hover:text-primary"} transition-colors text-sm font-bold`}
                      >
                        {time}
                      </button>
                    ))}
                  </div>
                </div>
              </div>
              <div className="mt-6 pt-6 border-t border-slate-200 dark:border-border-dark">
                <button className="w-full bg-slate-900 dark:bg-white text-white dark:text-slate-900 py-3 rounded-lg font-bold hover:opacity-90 transition-opacity">
                  {uiStrings?.landing?.schedule_demo_btn ||
                    "Schedule Live Demo"}
                </button>
              </div>
            </div>
          </div>
        </section>

        {/* Pricing Table */}
        <section className="py-20 px-4 md:px-10" id="pricing">
          <div className="max-w-[1280px] mx-auto">
            <div className="text-center mb-10">
              <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4">
                {uiStrings?.landing?.pricing_title ||
                  "Transparent Platform Pricing"}
              </h2>
              <div className="flex items-center justify-center gap-4 mb-10">
                <span className="text-sm font-bold text-slate-600 dark:text-text-secondary">
                  {uiStrings?.landing?.monthly_label || "Monthly"}
                </span>
                <div className="relative inline-block w-12 mr-2 align-middle select-none transition duration-200 ease-in">
                  <input
                    type="checkbox"
                    name="toggle"
                    id="billing-toggle"
                    className="toggle-checkbox absolute block w-6 h-6 rounded-full bg-white border-4 border-slate-300 appearance-none cursor-pointer"
                    checked={isYearly}
                    onChange={() => setIsYearly(!isYearly)}
                  />
                  <label
                    htmlFor="billing-toggle"
                    className="toggle-label block overflow-hidden h-6 rounded-full bg-slate-300 cursor-pointer"
                  ></label>
                </div>
                <span className="text-sm font-bold text-slate-900 dark:text-white">
                  {uiStrings?.landing?.yearly_label || "Annual"}{" "}
                  <span className="text-green-500 font-extrabold text-xs ml-1 bg-green-500/10 px-2 py-0.5 rounded-full border border-green-500/20">
                    {uiStrings?.landing?.save_label || "SAVE"}{" "}
                    {marketingMeta?.save_percent || "--%"}
                  </span>
                </span>
              </div>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse">
                <thead>
                  <tr>
                    <th className="p-6 bg-transparent min-w-[200px]"></th>
                    {PRICING.map((plan, idx) => (
                      <th
                        key={idx}
                        className={`p-6 ${plan.isPopular ? "bg-primary/5 border-x border-t border-primary/20 rounded-t-xl relative" : "bg-white dark:bg-surface-dark border-x border-t border-slate-200 dark:border-border-dark rounded-t-xl"} min-w-[200px]`}
                      >
                        {plan.isPopular && (
                          <div className="absolute top-0 left-0 w-full h-1 bg-primary"></div>
                        )}
                        <div
                          className={`text-lg font-bold ${plan.isPopular ? "text-primary" : "text-slate-900 dark:text-white"}`}
                        >
                          {plan.name}
                        </div>
                        <div className="text-3xl font-bold mt-2 text-slate-900 dark:text-white">
                          {typeof plan.monthly === "number"
                            ? `$${isYearly ? plan.yearly : plan.monthly}`
                            : plan.monthly}
                        </div>
                        <div className="text-xs text-slate-500 mb-4">
                          {typeof plan.monthly === "number"
                            ? "/vehicle/mo"
                            : ""}
                        </div>
                        <button
                          className={`w-full py-2 rounded ${plan.isPopular ? "bg-primary text-white hover:bg-primary/90 shadow-lg shadow-primary/20" : "border border-slate-300 dark:border-border-dark text-slate-900 dark:text-white hover:bg-slate-50 dark:hover:bg-background-dark"} text-sm font-bold`}
                        >
                          {plan.name === "Command"
                            ? landingMetadata?.cta?.sales_text ||
                              "Contact Sales"
                            : landingMetadata?.cta?.start_text || "Get Started"}
                        </button>
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-200 dark:divide-border-dark bg-white dark:bg-surface-dark border-b border-slate-200 dark:border-border-dark">
                  <tr>
                    <td
                      className="p-4 bg-slate-50 dark:bg-[#0d141c] text-xs font-bold text-slate-500 uppercase tracking-wide"
                      colSpan={PRICING.length + 1}
                    >
                      {uiStrings?.landing?.core_features_label ||
                        "Core Features"}
                    </td>
                  </tr>
                  <tr>
                    <td className="p-4 text-sm font-medium text-slate-900 dark:text-white pl-6">
                      Real-time GPS Tracking
                    </td>
                    {PRICING.map((_, idx) => (
                      <td key={idx} className="p-4 text-center">
                        <span className="material-symbols-outlined text-green-500">
                          check
                        </span>
                      </td>
                    ))}
                  </tr>
                </tbody>
              </table>
            </div>
            <div className="mt-12">
              <h4 className="text-lg font-bold text-slate-900 dark:text-white mb-6">
                {uiStrings?.landing?.hardware_addons_title ||
                  "Hardware Add-ons"}
              </h4>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                {HARDWARE_ADDONS.map((addon, i) => (
                  <div
                    key={i}
                    className="p-4 rounded-lg bg-white dark:bg-surface-dark border border-slate-200 dark:border-border-dark flex flex-col gap-2"
                  >
                    <div className="flex items-center gap-3">
                      <span className="material-symbols-outlined text-slate-400">
                        {addon.icon}
                      </span>
                      <div className="text-sm font-bold text-slate-700 dark:text-slate-200">
                        {addon.name}
                      </div>
                    </div>
                    <div className="text-xs font-mono text-primary ml-9">
                      {addon.price} / one-time
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </section>

        {/* FAQ */}
        <section className="py-20 px-4 md:px-10 bg-slate-50 dark:bg-[#0d141c]">
          <div className="max-w-4xl mx-auto">
            <h2 className="text-3xl font-bold text-center text-slate-900 dark:text-white mb-12">
              {uiStrings?.landing?.faq_title || "Frequently Asked Questions"}
            </h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {FAQS.map((faq, i) => (
                <details
                  key={i}
                  className="group bg-white dark:bg-surface-dark p-4 rounded-lg border border-slate-200 dark:border-border-dark"
                >
                  <summary className="flex justify-between items-center font-bold cursor-pointer list-none text-slate-900 dark:text-white">
                    {faq.q}
                    <span className="transition group-open:rotate-180 material-symbols-outlined">
                      expand_more
                    </span>
                  </summary>
                  <p className="text-slate-600 dark:text-text-secondary mt-3 text-sm leading-relaxed">
                    {faq.a}
                  </p>
                </details>
              ))}
            </div>
          </div>
        </section>

        {/* CTA */}
        <section className="py-20 px-4 md:px-10">
          <div className="max-w-5xl mx-auto rounded-2xl bg-gradient-to-br from-surface-dark to-background-dark border border-border-dark p-10 md:p-16 text-center relative overflow-hidden">
            <div className="absolute top-0 right-0 w-64 h-64 bg-primary/10 rounded-full blur-3xl -translate-y-1/2 translate-x-1/2"></div>
            <div className="absolute bottom-0 left-0 w-64 h-64 bg-blue-600/10 rounded-full blur-3xl translate-y-1/2 -translate-x-1/2"></div>
            <h2 className="relative z-10 text-3xl md:text-5xl font-bold text-white mb-6">
              {uiStrings?.landing?.cta_title ||
                "Ready to modernize your fleet?"}
            </h2>
            <p className="relative z-10 text-text-secondary text-lg mb-10 max-w-2xl mx-auto">
              {uiStrings?.landing?.cta_sub ||
                "Join over 2,000 companies optimizing their operations."}
            </p>
            <div className="relative z-10 flex flex-col sm:flex-row gap-4 justify-center">
              <button className="h-12 px-8 rounded-lg bg-primary hover:bg-primary/90 text-white text-base font-bold transition-all shadow-lg shadow-primary/25">
                {uiStrings?.landing?.trial_btn || "Start Free Trial"}
              </button>
              <button className="h-12 px-8 rounded-lg border border-border-dark bg-transparent hover:bg-border-dark text-white text-base font-bold transition-all">
                {uiStrings?.landing?.sales_btn || "Talk to Sales"}
              </button>
            </div>
          </div>
        </section>
      </main>

      {/* Footer */}
      <footer className="border-t border-slate-200 dark:border-border-dark bg-white dark:bg-background-dark pt-16 pb-8">
        <div className="px-4 md:px-10 max-w-[1280px] mx-auto">
          <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-8 mb-12">
            <div className="col-span-2 lg:col-span-2">
              <div className="flex items-center gap-3 text-slate-900 dark:text-white mb-4">
                <div className="size-6 flex items-center justify-center text-primary">
                  <span className="material-symbols-outlined">
                    local_shipping
                  </span>
                </div>
                <h2 className="text-lg font-bold">
                  {uiStrings?.landing?.company_label || "Fleet Ops"}
                </h2>
              </div>
              <p className="text-slate-500 dark:text-text-secondary text-sm max-w-xs mb-6">
                {uiStrings?.landing?.footer_desc ||
                  "The world's most advanced AI-driven fleet management platform."}
              </p>
            </div>
          </div>
          <div className="border-t border-slate-200 dark:border-border-dark pt-8 flex flex-col md:flex-row justify-between items-center gap-4">
            <p className="text-slate-400 text-sm">
              © {new Date().getFullYear()}{" "}
              {uiStrings?.landing?.footer_copy || "Flight Control Center"}
            </p>
          </div>
        </div>
      </footer>

      {/* Floating Chat Button */}
      <div className="fixed bottom-6 right-6 z-[60]">
        <button className="bg-primary hover:bg-primary-dark text-white px-6 py-4 rounded-full shadow-2xl shadow-primary/40 transition-transform hover:scale-105 flex items-center gap-3 font-bold text-base group">
          <span className="relative flex h-3 w-3">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
            <span className="relative inline-flex rounded-full h-3 w-3 bg-green-400"></span>
          </span>
          {uiStrings?.landing?.chat_btn || "Chat with an Expert"}
          <span className="material-symbols-outlined ml-1">chat</span>
        </button>
      </div>
    </div>
  );
};

export default Landing;
