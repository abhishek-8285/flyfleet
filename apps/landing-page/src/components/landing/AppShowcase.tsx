'use client';

import { motion } from 'framer-motion';

const fadeInUp = {
  initial: { opacity: 0, y: 30 },
  whileInView: { opacity: 1, y: 0 },
  viewport: { once: true },
  transition: { duration: 0.6 }
};

const AppShowcase = () => {
  return (
    <section className="py-24 px-4 md:px-10 bg-slate-50 dark:bg-background-dark text-slate-900 dark:text-white relative overflow-hidden" id="apps">
      <div className="absolute inset-0 bg-primary/5 z-0"></div>
      <div className="max-w-[1280px] mx-auto relative z-10">
        <div className="text-center mb-20">
          <div className="inline-block px-3 py-1 bg-primary/20 text-primary rounded-full text-xs font-bold uppercase tracking-wide mb-4">Ecosystem</div>
          <h2 className="text-3xl md:text-5xl font-bold mb-6 font-display text-slate-900 dark:text-white">Field-Ready Apps</h2>
          <p className="text-slate-600 dark:text-text-secondary text-lg max-w-2xl mx-auto">
            A complete ecosystem that connects your HQ, drivers, and clients in real-time. Native iOS and Android apps included with every plan.
          </p>
        </div>
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
          {/* Driver App Frame */}
          <motion.div {...fadeInUp} className="flex flex-col items-center">
            <div className="mobile-frame w-[280px] h-[580px] shadow-2xl bg-surface-dark transform">
              <div className="h-full w-full bg-[#111a22] flex flex-col pt-10 px-4 pb-4">
                <div className="flex justify-between items-center mb-6">
                  <div>
                    <div className="text-xs text-slate-400">Good Morning,</div>
                    <div className="font-bold text-lg">Driver Mike</div>
                  </div>
                  <div className="w-8 h-8 rounded-full bg-slate-700"></div>
                </div>
                <div className="bg-red-500/10 border border-red-500/30 rounded-xl p-4 mb-6 flex items-center gap-4 cursor-pointer hover:bg-red-500/20 transition-colors">
                  <div className="bg-red-500 p-3 rounded-full text-white shadow-lg shadow-red-500/40 animate-pulse">
                    <span className="material-symbols-outlined text-2xl">sos</span>
                  </div>
                  <div>
                    <div className="font-bold text-red-500">Emergency SOS</div>
                    <div className="text-[10px] text-red-400">Press and hold for 3s</div>
                  </div>
                </div>
                <div className="mb-2 flex justify-between items-end">
                  <h4 className="font-bold text-sm">Today's Tasks</h4>
                  <span className="text-[10px] text-primary">View All</span>
                </div>
                <div className="space-y-3 overflow-y-auto no-scrollbar flex-1">
                  <div className="bg-surface-dark p-3 rounded-xl border border-border-dark flex gap-3">
                    <div className="mt-1"><span className="material-symbols-outlined text-green-500 text-sm">radio_button_checked</span></div>
                    <div>
                      <div className="text-xs font-bold">Delivery #4922</div>
                      <div className="text-[10px] text-slate-400 mb-1">123 Logistics Way, NY</div>
                      <span className="text-[9px] bg-green-500/20 text-green-500 px-1.5 py-0.5 rounded">Completed</span>
                    </div>
                  </div>
                  <div className="bg-surface-dark p-3 rounded-xl border border-primary/30 flex gap-3 relative overflow-hidden">
                    <div className="absolute left-0 top-0 bottom-0 w-1 bg-primary"></div>
                    <div className="mt-1"><span className="material-symbols-outlined text-primary text-sm">radio_button_unchecked</span></div>
                    <div>
                      <div className="text-xs font-bold text-white">Pickup #5011</div>
                      <div className="text-[10px] text-slate-400 mb-2">45 Harbor Side, NY</div>
                      <div className="flex gap-2">
                        <button className="bg-primary text-white text-[9px] font-bold px-3 py-1 rounded">Navigate</button>
                        <button className="bg-slate-700 text-white text-[9px] font-bold px-3 py-1 rounded">Details</button>
                      </div>
                    </div>
                  </div>
                </div>
                <div className="border-t border-border-dark pt-4 mt-2 flex justify-around text-slate-500">
                  <span className="material-symbols-outlined text-primary">home</span>
                  <span className="material-symbols-outlined">map</span>
                  <span className="material-symbols-outlined">person</span>
                </div>
              </div>
            </div>
            <div className="mt-8 text-center" aria-label="Driver Companion App Description">
              <h3 className="text-xl font-bold mb-2 font-display">Driver Companion App</h3>
              <p className="text-text-secondary text-sm max-w-xs mx-auto">Task management, SOS alerts, and vehicle inspection forms.</p>
            </div>
          </motion.div>
          {/* Manager App Frame */}
          <motion.div {...fadeInUp} transition={{ delay: 0.2 }} className="flex flex-col items-center">
            <div className="mobile-frame w-[280px] h-[580px] shadow-2xl bg-surface-dark transform border-primary/30">
              <div className="h-full w-full bg-[#111a22] flex flex-col pt-0 relative overflow-hidden">
                <div className="absolute inset-0 bg-slate-800 opacity-30 z-0 bg-cover grayscale" style={{ backgroundImage: "url('https://images.unsplash.com/photo-1524661135-423995f22d0b?q=80&w=2674&auto=format&fit=crop')" }}></div>
                <div className="absolute top-1/3 left-1/4 z-10">
                  <div className="w-12 h-12 bg-primary/20 rounded-full flex items-center justify-center animate-pulse">
                    <div className="w-3 h-3 bg-primary rounded-full ring-2 ring-white"></div>
                  </div>
                </div>
                <div className="absolute top-1/2 left-2/3 z-10">
                  <div className="w-3 h-3 bg-green-500 rounded-full ring-2 ring-white shadow-lg"></div>
                </div>
                <div className="absolute bottom-0 w-full bg-surface-dark rounded-t-2xl z-20 p-5 pb-8 shadow-[0_-10px_40px_rgba(0,0,0,0.5)]">
                  <div className="w-12 h-1 bg-slate-700 rounded-full mx-auto mb-4"></div>
                  <div className="flex justify-between items-start mb-4">
                    <div>
                      <div className="text-[10px] font-bold text-primary uppercase tracking-wider mb-1">Vehicle Status</div>
                      <div className="text-lg font-bold text-white">Truck TR-401</div>
                      <div className="text-xs text-text-secondary">55 mph • I-95 North</div>
                    </div>
                  </div>
                  <button className="w-full bg-primary py-3 rounded-lg text-xs font-bold text-white shadow-lg shadow-primary/20">Track Live</button>
                </div>
              </div>
            </div>
            <div className="mt-8 text-center" aria-label="Manager Track App Description">
              <h3 className="text-xl font-bold mb-2 font-display">Manager Track App</h3>
              <p className="text-text-secondary text-sm max-w-xs mx-auto">Real-time fleet visibility and instant alerts from your pocket.</p>
            </div>
          </motion.div>
        </div>
      </div>
    </section>
  );
};

export default AppShowcase;
