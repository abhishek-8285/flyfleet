'use client';

import { motion } from 'framer-motion';

const LiveMapPreview = () => {
  return (
    <div className="relative w-full h-[500px] bg-slate-50 dark:bg-slate-900/30 rounded-3xl border border-slate-200 dark:border-border-dark overflow-hidden group shadow-2xl">
      <div className="absolute inset-0 opacity-[0.03] dark:opacity-[0.07]" style={{ backgroundImage: 'radial-gradient(#137fec 2px, transparent 2px)', backgroundSize: '40px 40px' }}></div>
      
      {/* Animated Trails */}
      <svg className="absolute inset-0 w-full h-full opacity-10 pointer-events-none">
        <motion.path 
          d="M 0 250 Q 300 100 600 250 T 1280 250" 
          stroke="url(#grad1)" strokeWidth="4" fill="none"
          initial={{ pathLength: 0 }}
          animate={{ pathLength: 1 }}
          transition={{ duration: 10, repeat: Infinity }}
        />
        <defs>
          <linearGradient id="grad1" x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stopColor="#137fec" stopOpacity="0" />
            <stop offset="50%" stopColor="#137fec" stopOpacity="1" />
            <stop offset="100%" stopColor="#137fec" stopOpacity="0" />
          </linearGradient>
        </defs>
      </svg>

      {/* Animated Trucks */}
      {[1, 2, 3, 4].map((i) => (
        <motion.div
          key={i}
          className="absolute p-3 bg-white dark:bg-surface-dark rounded-xl shadow-2xl border border-primary/20 flex items-center gap-3 z-10 backdrop-blur-sm"
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
      
      {/* Map UI Overlays */}
      <div className="absolute top-8 left-8 p-6 bg-white/90 dark:bg-surface-dark/90 backdrop-blur-xl rounded-2xl border border-slate-200 dark:border-border-dark z-20 shadow-xl">
        <div className="text-[10px] font-black text-primary mb-4 uppercase tracking-[0.2em]">Real-time Analytics</div>
        <div className="space-y-6">
          <div>
            <div className="flex justify-between items-end mb-1">
              <span className="text-xs font-bold text-slate-500">Fleet Efficiency</span>
              <span className="text-xs font-black text-slate-900 dark:text-white">94%</span>
            </div>
            <div className="w-48 h-1.5 bg-slate-100 dark:bg-slate-800 rounded-full overflow-hidden">
              <motion.div 
                className="h-full bg-primary"
                initial={{ width: 0 }}
                whileInView={{ width: "94%" }}
                transition={{ duration: 1, delay: 0.5 }}
              />
            </div>
          </div>
          <div className="flex gap-8">
            <div>
              <div className="text-2xl font-black text-slate-900 dark:text-white">128</div>
              <div className="text-[10px] font-bold text-slate-400 uppercase">Vehicles</div>
            </div>
            <div>
              <div className="text-2xl font-black text-green-500">02</div>
              <div className="text-[10px] font-bold text-slate-400 uppercase">Alerts</div>
            </div>
          </div>
        </div>
      </div>
      
      <div className="absolute bottom-8 right-8 z-20">
        <button className="bg-primary hover:bg-primary/90 text-white px-6 py-3 rounded-xl font-black text-sm shadow-xl shadow-primary/30 flex items-center gap-3 transition-transform hover:scale-105 active:scale-95">
          Launch Command Center
          <span className="material-symbols-outlined !text-sm">rocket_launch</span>
        </button>
      </div>

      <div className="absolute inset-0 pointer-events-none border-[12px] border-white/5 dark:border-white/[0.02] rounded-3xl"></div>
    </div>
  );
};

export default LiveMapPreview;
