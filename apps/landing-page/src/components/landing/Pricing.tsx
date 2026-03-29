'use client';

import { useEffect, useRef } from 'react';
import { motion } from 'framer-motion';
import React, { useState } from 'react';

const HARDWARE_ADDONS = [
  { icon: 'videocam', name: 'AI Dashcam', price: '$199' },
  { icon: 'thermostat', name: 'Temp Sensor', price: '$49' },
  { icon: 'door_sensor', name: 'Door Sensor', price: '$39' },
  { icon: 'lock', name: 'Remote Immobilizer', price: '$149' }
];

const Pricing = () => {
  const [isYearly, setIsYearly] = useState(false);
  const sectionRef = useRef<HTMLElement>(null);

  useEffect(() => {
    const gsap = (window as any).gsap;
    if (!gsap) return;

    const ctx = gsap.context(() => {
      // Floating animation for pricing cards
      gsap.utils.toArray('.pricing-card').forEach((card: Element, idx: number) => {
        gsap.to(card, {
          y: '+=15',
          duration: 4 + idx * 0.5,
          repeat: -1,
          yoyo: true,
          ease: 'sine.inOut',
          delay: idx * 0.3
        });
      });

      // Parallax on scroll
      gsap.utils.toArray('.pricing-card').forEach((card: Element, idx: number) => {
        gsap.to(card, {
          y: -20 * (idx + 1),
          ease: 'none',
          scrollTrigger: {
            trigger: sectionRef.current,
            start: 'top bottom',
            end: 'bottom top',
            scrub: 1
          }
        });
      });
    }, sectionRef);

    return () => ctx.revert();
  }, []);

  return (
    <section
      ref={sectionRef}
      className="scene relative"
      style={{ '--scene-height': '250vh', paddingTop: '0', paddingBottom: '0' } as React.CSSProperties}
      data-scene="pricing"
    >
      {/* Background depth layers */}
      <div className="depth-0 absolute inset-0 bg-white dark:bg-background-dark" />
      <div className="depth-1 absolute inset-0 pointer-events-none">
        <div className="absolute top-1/3 left-1/3 w-96 h-96 bg-primary/5 rounded-full blur-3xl"></div>
        <div className="absolute bottom-1/3 right-1/3 w-96 h-96 bg-blue-500/5 rounded-full blur-3xl"></div>
      </div>

      <div className="relative z-10 max-w-[1280px] mx-auto px-4 md:px-10 py-24">
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="text-center mb-10"
        >
          <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4 font-display">Transparent Platform Pricing</h2>
          <div className="flex items-center justify-center gap-4 mb-10">
            <span className="text-sm font-bold text-slate-600 dark:text-text-secondary">Monthly</span>
            <div
              className="relative w-14 h-7 bg-slate-300 dark:bg-slate-700 rounded-full cursor-pointer transition-colors p-1.5"
              onClick={() => setIsYearly(!isYearly)}
            >
              <motion.div
                className="w-4 h-4 bg-white rounded-full shadow-md"
                animate={{ x: isYearly ? 28 : 0 }}
                transition={{ type: "spring", stiffness: 500, damping: 30 }}
              />
              <span className="absolute inset-0 flex items-center text-[10px] font-bold text-slate-400">
                <span className="flex-1 text-center">MO</span>
                <span className="flex-1 text-center">YR</span>
              </span>
            </div>
            <span className="text-sm font-bold text-slate-900 dark:text-white">Annual <span className="text-green-500 font-extrabold text-xs ml-1 bg-green-500/10 px-2 py-0.5 rounded-full border border-green-500/20">SAVE 20%</span></span>
          </div>
        </motion.div>

        <div className="overflow-x-auto mb-20">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr>
                <th className="p-6 bg-transparent min-w-[200px]"></th>
                <th className="p-6 bg-white dark:bg-surface-dark border-x border-t border-slate-200 dark:border-border-dark rounded-t-xl min-w-[200px]">
                  <div className="text-lg font-bold text-slate-900 dark:text-white font-display">Starter</div>
                  <div className="text-3xl font-bold mt-2 text-slate-900 dark:text-white font-display">${isYearly ? '24' : '29'}</div>
                  <div className="text-xs text-slate-500 mb-4">/vehicle/mo</div>
                  <button className="w-full py-2 rounded border border-slate-300 dark:border-border-dark text-sm font-bold text-slate-900 dark:text-white hover:bg-slate-50 dark:hover:bg-background-dark transition-all">Get Started</button>
                </th>
                <th className="p-6 bg-primary/5 border-x border-t border-primary/20 rounded-t-xl min-w-[200px] relative depth-2">
                  <div className="absolute top-0 left-0 w-full h-1 bg-primary"></div>
                  <div className="absolute -inset-2 bg-primary/5 rounded-3xl -z-10 blur-xl"></div>
                  <div className="text-lg font-bold text-primary font-display">Pro</div>
                  <div className="text-3xl font-bold mt-2 text-slate-900 dark:text-white font-display">${isYearly ? '39' : '49'}</div>
                  <div className="text-xs text-slate-500 mb-4">/vehicle/mo</div>
                  <button className="w-full py-2 rounded bg-primary text-sm font-bold text-white hover:bg-primary/90 shadow-lg shadow-primary/20 transition-all">Free Trial</button>
                </th>
                <th className="p-6 bg-white dark:bg-surface-dark border-x border-t border-slate-200 dark:border-border-dark rounded-t-xl min-w-[200px]">
                  <div className="text-lg font-bold text-slate-900 dark:text-white font-display">Command</div>
                  <div className="text-3xl font-bold mt-2 text-slate-900 dark:text-white font-display">Custom</div>
                  <button className="w-full py-2 rounded border border-slate-300 dark:border-border-dark text-sm font-bold text-slate-900 dark:text-white hover:bg-slate-50 dark:hover:bg-background-dark transition-all">Contact Sales</button>
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-200 dark:divide-border-dark bg-white dark:bg-surface-dark border-b border-slate-200 dark:border-border-dark">
              <tr><td className="p-4 bg-slate-50 dark:bg-background-dark/30 text-xs font-bold text-slate-500 uppercase tracking-wide px-6" colSpan={4}>Core Features</td></tr>
              <tr>
                <td className="p-4 text-sm font-medium text-slate-900 dark:text-white pl-6">Real-time GPS Tracking</td>
                <td className="p-4 text-center"><span className="material-symbols-outlined text-green-500">check</span></td>
                <td className="p-4 text-center bg-primary/5"><span className="material-symbols-outlined text-green-500">check</span></td>
                <td className="p-4 text-center"><span className="material-symbols-outlined text-green-500">check</span></td>
              </tr>
              <tr>
                <td className="p-4 text-sm font-medium text-slate-900 dark:text-white pl-6">AI Route Optimization</td>
                <td className="p-4 text-center"><span className="material-symbols-outlined text-slate-300">remove</span></td>
                <td className="p-4 text-center bg-primary/5"><span className="material-symbols-outlined text-green-500">check</span></td>
                <td className="p-4 text-center"><span className="material-symbols-outlined text-green-500">check</span></td>
              </tr>
              <tr>
                <td className="p-4 text-sm font-medium text-slate-900 dark:text-white pl-6">API Access & SDKs</td>
                <td className="p-4 text-center"><span className="material-symbols-outlined text-slate-300">remove</span></td>
                <td className="p-4 text-center bg-primary/5"><span className="material-symbols-outlined text-green-500">check</span></td>
                <td className="p-4 text-center"><span className="material-symbols-outlined text-green-500">check</span></td>
              </tr>
            </tbody>
          </table>
        </div>

        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="text-center mb-12"
        >
          <h3 className="text-2xl font-bold text-slate-900 dark:text-white mb-4">Hardware Add-ons</h3>
          <p className="text-slate-500">One-time hardware costs to enhance your fleet capabilities.</p>
        </motion.div>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
          {HARDWARE_ADDONS.map((addon, i) => (
            <motion.div
              key={i}
              whileHover={{ scale: 1.05, y: -5 }}
              className="pricing-card depth-2 p-6 rounded-2xl bg-white dark:bg-surface-dark border border-slate-200 dark:border-border-dark flex flex-col items-center text-center shadow-sm hover:shadow-lg hover:shadow-primary/10 transition-all"
            >
              <div className="size-12 rounded-full bg-primary/10 flex items-center justify-center text-primary mb-4 group-hover:bg-primary group-hover:text-white transition-colors duration-300">
                <span className="material-symbols-outlined !text-2xl">{addon.icon}</span>
              </div>
              <h4 className="font-bold text-slate-800 dark:text-white text-sm mb-1">{addon.name}</h4>
              <div className="text-primary font-black">{addon.price}</div>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
};

export default Pricing;
