'use client';

import { motion } from 'framer-motion';
import LiveMapPreview from './LiveMapPreview';

const fadeInUp = {
  initial: { opacity: 0, y: 30 },
  whileInView: { opacity: 1, y: 0 },
  viewport: { once: true },
  transition: { duration: 0.6 }
};

const InteractiveIntelligence = () => {
  return (
    <section className="py-24 px-4 md:px-10">
      <div className="max-w-[1280px] mx-auto">
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-16 items-center">
          <div className="lg:col-span-5">
            <motion.div {...fadeInUp}>
              <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-primary/10 border border-primary/20 text-primary text-xs font-black uppercase tracking-widest mb-6">
                Live Intelligence
              </div>
              <h2 className="text-4xl md:text-5xl font-black text-slate-900 dark:text-white mb-8 font-display leading-[1.1]">
                See Your Fleet <br/>
                <span className="text-primary italic">In Motion.</span>
              </h2>
              <p className="text-slate-600 dark:text-text-secondary text-lg mb-10 leading-relaxed">
                Stop guessing and start knowing. Our real-time engine processes billions of data points to give you a crystal-clear view of every vehicle, every route, and every cost-saving opportunity.
              </p>
              <div className="space-y-6">
                {[
                  { t: '5-Second Latency', d: 'The fastest data refresh rates in the industry.', i: 'bolt' },
                  { t: 'Predictive ETAs', d: 'AI-calculated arrivals based on live traffic and weather.', i: 'auto_graph' }
                ].map((item, i) => (
                  <div key={i} className="flex gap-4 p-4 rounded-2xl bg-slate-50 dark:bg-surface-dark border border-slate-100 dark:border-border-dark group hover:border-primary/30 transition-colors">
                    <div className="size-10 rounded-xl bg-white dark:bg-background-dark shadow-sm flex items-center justify-center text-primary group-hover:scale-110 transition-transform">
                      <span className="material-symbols-outlined !text-xl">{item.i}</span>
                    </div>
                    <div>
                      <h4 className="font-bold text-slate-900 dark:text-white">{item.t}</h4>
                      <p className="text-sm text-slate-500">{item.d}</p>
                    </div>
                  </div>
                ))}
              </div>
            </motion.div>
          </div>
          <div className="lg:col-span-7">
            <motion.div 
              initial={{ opacity: 0, scale: 0.95 }}
              whileInView={{ opacity: 1, scale: 1 }}
              viewport={{ once: true }}
              transition={{ duration: 0.8 }}
            >
              <LiveMapPreview />
            </motion.div>
          </div>
        </div>
      </div>
    </section>
  );
};

export default InteractiveIntelligence;
