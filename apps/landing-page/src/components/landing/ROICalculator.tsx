'use client';

import { useEffect, useRef } from 'react';
import { motion } from 'framer-motion';
import React, { useState } from 'react';

interface ROICalculatorProps {
  companyName: string;
}

const ROICalculator: React.FC<ROICalculatorProps> = ({ companyName }) => {
  const [fleetSize, setFleetSize] = useState(50);
  const [fuelCost, setFuelCost] = useState(3000);
  const sectionRef = useRef<HTMLElement>(null);
  const resultRef = useRef<HTMLDivElement>(null);

  const monthlySavings = fleetSize * (fuelCost * 0.15 + 120);
  const yearlySavings = monthlySavings * 12;

  useEffect(() => {
    const gsap = (window as any).gsap;
    const ScrollTrigger = (window as any).ScrollTrigger;
    if (!gsap) return;

    const ctx = gsap.context(() => {
      if (ScrollTrigger) {
        // Pin the results card while user explores
        ScrollTrigger.create({
          trigger: sectionRef.current,
          start: 'top 20%',
          end: 'bottom 80%',
          pin: resultRef.current,
          pinSpacing: true,
          scrub: 1
        });

        // Animate yearly number based on scroll position within pinned section
        ScrollTrigger.create({
          trigger: sectionRef.current,
          start: 'top 10%',
          end: 'bottom 90%',
          onUpdate: (self: any) => {
            const progress = self.progress;
            const target = Math.round(yearlySavings * progress);
            const el = document.getElementById('yearly-savings');
            if (el) {
              gsap.to(el, {
                innerText: target,
                duration: 0.1,
                snap: { innerText: 1 }
              });
            }
          }
        });
      }
    }, sectionRef);

    return () => ctx.revert();
  }, [yearlySavings]);

  return (
    <section
      ref={sectionRef}
      className="scene relative"
      style={{ '--scene-height': '300vh', paddingTop: '0', paddingBottom: '0' } as React.CSSProperties}
      data-scene="roi"
    >
      {/* Background depth layers */}
      <div className="depth-0 absolute inset-0 bg-primary/5" />
      <div className="depth-1 absolute inset-0 pointer-events-none">
        <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-primary/10 rounded-full blur-3xl"></div>
        <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-blue-500/10 rounded-full blur-3xl"></div>
      </div>

      <div className="relative z-10 max-w-[1280px] mx-auto px-4 md:px-10 py-24">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
          {/* Left: Controls */}
          <div className="space-y-12">
            <motion.div
              initial={{ opacity: 0, x: -50 }}
              whileInView={{ opacity: 1, x: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.6 }}
            >
              <h2 className="text-3xl md:text-5xl font-bold text-slate-900 dark:text-white mb-6 font-display">
                Calculate Your <span className="text-primary italic">Estimated ROI</span>
              </h2>
              <p className="text-slate-600 dark:text-text-secondary text-lg mb-10">
                See how much you could save by switching to {companyName}'s AI-driven platform. Most enterprise fleets see a break-even within 4 months.
              </p>
            </motion.div>

            <motion.div
              initial={{ opacity: 0, x: -50 }}
              whileInView={{ opacity: 1, x: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.6, delay: 0.2 }}
              className="space-y-8"
            >
              <div className="group">
                <div className="flex justify-between items-center mb-4">
                  <label className="font-bold text-slate-700 dark:text-slate-300 text-sm uppercase tracking-wider">Fleet size</label>
                  <span className="text-primary font-black text-2xl">{fleetSize} <span className="text-sm text-slate-500 font-normal">Vehicles</span></span>
                </div>
                <input
                  type="range"
                  min="1"
                  max="1000"
                  step="5"
                  value={fleetSize}
                  onChange={(e) => setFleetSize(parseInt(e.target.value))}
                  className="w-full h-3 bg-slate-200 dark:bg-slate-700 rounded-full appearance-none cursor-pointer accent-primary hover:accent-primary/80 transition-all"
                  style={{
                    backgroundSize: `${(fleetSize / 1000) * 100}% 100%`,
                    backgroundImage: 'linear-gradient(#137fec, #137fec)'
                  }}
                />
              </div>

              <div className="group">
                <div className="flex justify-between items-center mb-4">
                  <label className="font-bold text-slate-700 dark:text-slate-300 text-sm uppercase tracking-wider">Monthly fuel cost per vehicle</label>
                  <span className="text-primary font-black text-2xl">${fuelCost.toLocaleString()}</span>
                </div>
                <input
                  type="range"
                  min="500"
                  max="10000"
                  step="100"
                  value={fuelCost}
                  onChange={(e) => setFuelCost(parseInt(e.target.value))}
                  className="w-full h-3 bg-slate-200 dark:bg-slate-700 rounded-full appearance-none cursor-pointer accent-primary hover:accent-primary/80 transition-all"
                  style={{
                    backgroundSize: `${((fuelCost - 500) / 9500) * 100}% 100%`,
                    backgroundImage: 'linear-gradient(#137fec, #137fec)'
                  }}
                />
              </div>
            </motion.div>
          </div>

          {/* Right: Pinned Results Card */}
          <div ref={resultRef} className="relative">
            <motion.div
              initial={{ opacity: 0, scale: 0.9, y: 50 }}
              whileInView={{ opacity: 1, scale: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.6, delay: 0.4 }}
              className="depth-3 bg-white dark:bg-surface-dark p-8 md:p-12 rounded-3xl shadow-2xl border border-slate-200 dark:border-border-dark relative overflow-hidden"
            >
              {/* Decorative glow */}
              <div className="absolute top-0 right-0 w-32 h-32 bg-primary/10 rounded-full -translate-y-1/2 translate-x-1/2 blur-2xl"></div>
              <div className="absolute bottom-0 left-0 w-32 h-32 bg-green-500/10 rounded-full translate-y-1/2 -translate-x-1/2 blur-2xl"></div>

              <div className="relative z-10">
                <h3 className="text-slate-500 dark:text-text-secondary font-bold uppercase tracking-widest text-xs mb-8">Estimated Annual Savings</h3>
                <div className="text-5xl md:text-7xl font-black text-slate-900 dark:text-white mb-4 font-display">
                  $<span id="yearly-savings">{yearlySavings.toLocaleString()}</span>
                </div>
                <motion.p
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  key={yearlySavings}
                  className="text-green-500 font-bold flex items-center gap-2 mb-10 text-lg"
                >
                  <motion.span
                    initial={{ scale: 0 }}
                    animate={{ scale: 1 }}
                    key="trend-icon"
                    className="material-symbols-outlined"
                  >
                    trending_up
                  </motion.span>
                  Approx. 22% reduction in total operational costs
                </motion.p>
                <div className="grid grid-cols-2 gap-4">
                  <motion.div
                    key={`monthly-${monthlySavings}`}
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    className="p-4 rounded-xl bg-slate-50 dark:bg-background-dark/50 border border-slate-100 dark:border-border-dark"
                  >
                    <div className="text-slate-500 text-xs mb-1 uppercase tracking-wide">Monthly Savings</div>
                    <div className="text-xl font-bold text-slate-900 dark:text-white">${monthlySavings.toLocaleString()}</div>
                  </motion.div>
                  <motion.div
                    key={`payback-${monthlySavings}`}
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: 0.1 }}
                    className="p-4 rounded-xl bg-slate-50 dark:bg-background-dark/50 border border-slate-100 dark:border-border-dark"
                  >
                    <div className="text-slate-500 text-xs mb-1 uppercase tracking-wide">Payback Period</div>
                    <div className="text-xl font-bold text-slate-900 dark:text-white">
                      {monthlySavings > 0 ? `~${(3.8 * (50 / fleetSize)).toFixed(1)} Months` : 'N/A'}
                    </div>
                  </motion.div>
                </div>
                <motion.button
                  whileHover={{ scale: 1.02, boxShadow: '0 20px 40px rgba(19,127,236,0.3)' }}
                  whileTap={{ scale: 0.98 }}
                  className="w-full mt-10 h-14 bg-primary hover:bg-primary/90 text-white font-bold rounded-xl shadow-lg shadow-primary/20 transition-all flex items-center justify-center gap-3"
                >
                  Get Detailed ROI Report
                  <span className="material-symbols-outlined !text-xl">description</span>
                </motion.button>
              </div>
            </motion.div>
          </div>
        </div>
      </div>
    </section>
  );
};

export default ROICalculator;
