'use client';

import { motion } from 'framer-motion';

const INTEGRATIONS = [
  { icon: 'database', label: 'SAP / ERP' },
  { icon: 'chat', label: 'WhatsApp API' },
  { icon: 'local_gas_station', label: 'Fuel Cards' },
  { icon: 'toll', label: 'FastTag' },
  { icon: 'map', label: 'Google Maps' },
  { icon: 'webhook', label: 'Webhooks' }
];

const fadeInUp = {
  initial: { opacity: 0, y: 30 },
  whileInView: { opacity: 1, y: 0 },
  viewport: { once: true },
  transition: { duration: 0.6 }
};

const staggerContainer = {
  initial: {},
  whileInView: { transition: { staggerChildren: 0.1 } }
};

const Integrations = () => {
  return (
    <section className="py-20 px-4 md:px-10 bg-slate-50 dark:bg-background-dark/50">
      <div className="max-w-[1280px] mx-auto">
        <div className="flex flex-col md:flex-row justify-between items-end mb-12 gap-6">
          <motion.div {...fadeInUp}>
            <h2 className="text-3xl font-bold text-slate-900 dark:text-white mb-4 font-display">Deep Integrations</h2>
            <p className="text-slate-600 dark:text-text-secondary max-w-xl">
              Seamlessly connect FlyFleet into your existing tech stack. We support major ERPs, fuel card providers, and messaging platforms.
            </p>
          </motion.div>
          <a className="flex items-center gap-2 text-primary font-bold hover:underline" href="#">
            View API Documentation <span className="material-symbols-outlined text-sm">arrow_forward</span>
          </a>
        </div>
        <motion.div 
          variants={staggerContainer}
          initial="initial"
          whileInView="whileInView"
          viewport={{ once: true }}
          className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4"
        >
          {INTEGRATIONS.map((item, index) => (
            <motion.div 
              key={index} 
              variants={fadeInUp}
              whileHover={{ y: -5 }}
              className="bg-white dark:bg-surface-dark p-6 rounded-xl border border-slate-200 dark:border-border-dark flex flex-col items-center justify-center gap-3 hover:border-primary transition-colors group"
            >
              <span className="material-symbols-outlined text-4xl text-slate-400 group-hover:text-primary">{item.icon}</span>
              <span className="font-bold text-sm text-slate-700 dark:text-slate-300 font-display">{item.label}</span>
            </motion.div>
          ))}
        </motion.div>
      </div>
    </section>
  );
};

export default Integrations;
