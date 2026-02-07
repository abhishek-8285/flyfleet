'use client';

import { motion } from 'framer-motion';

const ImplementationRoadmap = () => {
  return (
    <section className="py-24 px-4 md:px-10 bg-slate-50 dark:bg-background-dark/30" id="how-it-works">
      <div className="max-w-[1280px] mx-auto">
        <motion.div 
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="text-center mb-20"
        >
          <div className="inline-block px-3 py-1 bg-green-500/10 text-green-500 rounded-full text-xs font-black uppercase tracking-widest mb-4">The Blueprint</div>
          <h2 className="text-3xl md:text-5xl font-black text-slate-900 dark:text-white mb-6 font-display">Go-Live in 21 Days</h2>
          <p className="text-slate-600 dark:text-text-secondary text-lg max-w-2xl mx-auto">Our white-glove implementation process ensures zero downtime for your operations.</p>
        </motion.div>

        <div className="max-w-5xl mx-auto mb-24 relative">
          <div className="absolute top-1/2 left-0 w-full h-[2px] bg-slate-200 dark:bg-border-dark -translate-y-1/2 hidden md:block z-0"></div>
          
          <div className="grid grid-cols-1 md:grid-cols-4 gap-12 relative z-10">
            {[
              { w: 'Week 1', t: 'Analysis', d: 'Hardware audit & legacy data migration planning.', i: 'analytics', c: 'blue' },
              { w: 'Week 2', t: 'Integration', d: 'ERP setup & geofence zone configuration.', i: 'api', c: 'green' },
              { w: 'Week 3', t: 'Rollout', d: 'Rapid hardware install & driver on-boarding.', i: 'moped', c: 'blue' },
              { w: 'Go-Live', t: 'Scale', d: 'Full platform launch & 24/7 success monitoring.', i: 'rocket_launch', c: 'green' }
            ].map((step, idx) => (
              <motion.div 
                key={idx}
                whileHover={{ y: -5 }}
                className="text-center group"
              >
                <div className={`size-16 mx-auto rounded-2xl bg-${step.c === 'blue' ? 'primary' : step.c === 'green' ? 'green-500' : step.c === 'purple' ? 'purple-500' : 'orange-500'} flex items-center justify-center text-white shadow-xl shadow-${step.c === 'blue' ? 'primary' : step.c}-500/30 mb-6 transition-transform group-hover:scale-110`}>
                  <span className="material-symbols-outlined !text-3xl">{step.i}</span>
                </div>
                <h4 className={`text-xs font-black uppercase tracking-widest mb-2 text-${step.c === 'blue' ? 'primary' : step.c === 'green' ? 'green-500' : step.c === 'purple' ? 'purple-500' : 'orange-500'}`}>{step.w}</h4>
                <h3 className="text-lg font-bold text-slate-900 dark:text-white mb-3 font-display">{step.t}</h3>
                <p className="text-xs text-slate-500 dark:text-text-secondary leading-relaxed px-4">{step.d}</p>
              </motion.div>
            ))}
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          <motion.div 
            initial={{ opacity: 0, y: 30 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6 }}
            className="group p-10 rounded-3xl bg-white dark:bg-surface-dark border border-slate-200 dark:border-border-dark hover:border-primary/40 transition-all shadow-xl"
          >
            <h3 className="text-2xl font-black text-slate-900 dark:text-white mb-8 font-display">Hardware Ecosystem</h3>
            <div className="space-y-8">
              {[
                { t: 'OBD-II Smart Connect', d: '30-second plug-and-play installation for all post-2008 vehicles.', i: 'settings_input_component' },
                { t: 'Hardwired Stealth', d: 'Hidden installation with remote immobilization & anti-tamper alerts.', i: 'cable' }
              ].map((item, i) => (
                <div key={i} className="flex gap-6 items-start">
                  <div className="size-12 rounded-xl bg-slate-50 dark:bg-background-dark border border-slate-100 dark:border-border-dark flex items-center justify-center text-primary flex-shrink-0">
                    <span className="material-symbols-outlined">{item.i}</span>
                  </div>
                  <div>
                    <h4 className="text-lg font-bold text-slate-800 dark:text-white mb-1">{item.t}</h4>
                    <p className="text-sm text-slate-500 dark:text-text-secondary leading-relaxed">{item.d}</p>
                  </div>
                </div>
              ))}
            </div>
          </motion.div>

          <motion.div 
            initial={{ opacity: 0, scale: 0.95 }}
            whileInView={{ opacity: 1, scale: 1 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6, delay: 0.2 }}
            className="relative bg-white dark:bg-[#0d141b] rounded-3xl overflow-hidden min-h-[400px] flex items-center justify-center border border-slate-200 dark:border-white/5 shadow-xl dark:shadow-none"
          >
            <div className="absolute inset-0 bg-[radial-gradient(circle_at_50%_50%,_rgba(19,127,236,0.05),transparent_70%)] dark:bg-[radial-gradient(circle_at_50%_50%,_rgba(19,127,236,0.15),transparent_70%)]"></div>
            <div className="relative z-10 flex flex-col items-center">
              <motion.div 
                animate={{ y: [0, -20, 0] }}
                transition={{ duration: 4, repeat: Infinity, ease: "easeInOut" }}
                className="size-48 bg-gradient-to-br from-slate-100 to-slate-200 dark:from-slate-700 dark:to-slate-900 rounded-2xl shadow-2xl flex items-center justify-center relative border border-white/20 dark:border-white/10"
              >
                <span className="material-symbols-outlined !text-8xl text-slate-400 dark:text-slate-500 opacity-50">router</span>
                <div className="absolute bottom-4 left-1/2 -translate-x-1/2 flex gap-2">
                  <div className="size-1.5 rounded-full bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.8)]"></div>
                  <div className="size-1.5 rounded-full bg-blue-500 shadow-[0_0_8px_rgba(59,130,246,0.8)]"></div>
                  <div className="size-1.5 rounded-full bg-slate-600"></div>
                </div>
              </motion.div>
              <div className="mt-8 text-center">
                <p className="text-blue-500 dark:text-blue-400 font-black text-xs uppercase tracking-widest mb-2">Model F-200X</p>
                <p className="text-slate-500 text-[10px] font-mono">5G LTE • DUAL CAN-BUS • INTERNAL BATTERY</p>
              </div>
            </div>
          </motion.div>
        </div>
      </div>
    </section>
  );
};

export default ImplementationRoadmap;
