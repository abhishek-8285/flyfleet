'use client';

import { useEffect } from 'react';
import { motion } from 'framer-motion';

interface LiveMapPreviewProps {
  className?: string;
}

const LiveMapPreview: React.FC<LiveMapPreviewProps> = ({ className = '' }) => {
  useEffect(() => {
    // Component mounted, GSAP already handled by parent
  }, []);

  return (
    <div
      className={`relative w-full h-[500px] bg-slate-50 dark:bg-slate-900/30 rounded-3xl border border-slate-200 dark:border-border-dark overflow-hidden group shadow-2xl ${className}`}
    >
      {/* DEPTH 0: Map background with grid - far background */}
      <div className="depth-0 absolute inset-0" style={{ filter: 'blur(8px)', transform: 'scale(0.70)' }}>
        <div className="absolute inset-0 opacity-[0.05] dark:opacity-[0.08]" style={{ backgroundImage: 'radial-gradient(#137fec 2px, transparent 2px)', backgroundSize: '40px 40px' }}></div>
      </div>

      {/* DEPTH 1: Glow atmosphere */}
      <div className="depth-1 absolute inset-0 pointer-events-none">
        <div className="absolute inset-0 bg-gradient-radial from-primary/5 via-transparent to-transparent"></div>
      </div>

      {/* DEPTH 2: Animated truck icons - floating at mid ground */}
      <div className="depth-2 relative w-full h-full pointer-events-none">
        {[1, 2, 3, 4].map((i) => (
          <motion.div
            key={i}
            className="absolute p-3 bg-white/90 dark:bg-surface-dark/90 rounded-xl shadow-2xl border border-primary/20 flex items-center gap-3 backdrop-blur-sm"
            initial={{ x: -200, y: 100 * i - 50 }}
            animate={{
              x: [null, 300 * i, 1400],
              y: [null, (100 * i) + (Math.sin(i) * 50), 100 * i]
            }}
            transition={{
              duration: 20 + i * 5,
              repeat: Infinity,
              ease: "linear"
            }}
            data-depth="2"
          >
            <div className="size-8 bg-primary/10 rounded-full flex items-center justify-center">
              <span className="material-symbols-outlined !text-lg text-primary">local_shipping</span>
            </div>
            <div>
              <div className="text-[10px] font-black text-slate-800 dark:text-white uppercase tracking-tighter">FLEET-TR-{400 + i}</div>
              <div className="flex items-center gap-1">
                <span className="size-1.5 rounded-full bg-green-500 animate-pulse"></span>
                <span className="text-[9px] text-slate-500 font-bold">58 MPH • I-95</span>
              </div>
            </div>
          </motion.div>
        ))}
      </div>

      {/* DEPTH 4: Map UI Overlays - UI layer */}
      <div className="depth-4 absolute top-8 left-8 z-20">
        <div className="p-6 bg-white/90 dark:bg-surface-dark/90 backdrop-blur-xl rounded-2xl border border-slate-200 dark:border-border-dark shadow-xl">
          <div className="text-[10px] font-black text-primary mb-4 uppercase tracking-[0.2em]">Real-time Analytics</div>
          <div className="space-y-6">
            <div>
              <div className="flex justify-between items-end mb-1">
                <span className="text-xs font-bold text-slate-500">Fleet Efficiency</span>
                <span className="text-xs font-black text-slate-900 dark:text-white">94%</span>
              </div>
              <div className="w-48 h-1.5 bg-slate-100 dark:bg-slate-800 rounded-full overflow-hidden">
                <motion.div
                  id="efficiency-bar"
                  className="h-full bg-primary"
                  initial={{ width: 0 }}
                />
              </div>
            </div>
            <div className="flex gap-8">
              <div>
                <div className="text-2xl font-black text-slate-900 dark:text-white">
                  <span id="vehicle-count">0</span>
                </div>
                <div className="text-[10px] font-bold text-slate-400 uppercase">Vehicles</div>
              </div>
              <div>
                <div className="text-2xl font-black text-green-500">
                  <span id="alerts-count">0</span>
                </div>
                <div className="text-[10px] font-bold text-slate-400 uppercase">Alerts</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* DEPTH 5: Foreground particles */}
      <div className="depth-5 absolute inset-0 pointer-events-none">
        {/* Static decorative element */}
        <div className="absolute top-0 right-0 w-32 h-32 bg-primary/5 rounded-full blur-2xl"></div>
      </div>

      {/* CTA Button remains on top */}
      <div className="absolute bottom-8 right-8 z-50">
        <button className="bg-primary hover:bg-primary/90 text-white px-6 py-3 rounded-xl font-black text-sm shadow-xl shadow-primary/30 flex items-center gap-3 transition-transform hover:scale-105 active:scale-95">
          Launch Command Center
          <span className="material-symbols-outlined !text-sm">rocket_launch</span>
        </button>
      </div>

      {/* Border overlay */}
      <div className="absolute inset-0 pointer-events-none border-[12px] border-white/5 dark:border-white/[0.02] rounded-3xl"></div>
    </div>
  );
};

export default LiveMapPreview;
