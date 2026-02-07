'use client';

import { motion } from 'framer-motion';

const SECURITY_ITEMS = [
  { icon: 'verified_user', label: 'ISO 27001', desc: 'Certified information security management systems.' },
  { icon: 'shield', label: 'GDPR Compliant', desc: 'Full compliance with European data privacy regulations.' },
  { icon: 'lock', label: 'SOC 2 Type II', desc: 'Audited controls for security, availability, and confidentiality.' },
  { icon: 'timer', label: '99.9% SLA', desc: 'Guaranteed uptime with financial backing for enterprise plans.' }
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

const Security = () => {
  return (
    <section className="py-20 px-4 md:px-10" id="security">
      <div className="max-w-[1280px] mx-auto bg-white dark:bg-background-dark rounded-3xl p-8 md:p-16 border border-slate-200 dark:border-border-dark relative overflow-hidden shadow-xl dark:shadow-none">
        <div className="absolute top-0 right-0 w-96 h-96 bg-primary/10 rounded-full blur-3xl -translate-y-1/2 translate-x-1/2 pointer-events-none"></div>
        <motion.div {...fadeInUp} className="relative z-10 text-center mb-16">
          <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-6 font-display">Enterprise-Grade Security</h2>
          <p className="text-slate-600 dark:text-text-secondary max-w-2xl mx-auto text-lg">
            We maintain the highest standards of data protection and reliability, ensuring your fleet data is safe, compliant, and always accessible.
          </p>
        </motion.div>
        <motion.div 
          variants={staggerContainer}
          initial="initial"
          whileInView="whileInView"
          viewport={{ once: true }}
          className="relative z-10 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-8"
        >
          {SECURITY_ITEMS.map((item, index) => (
            <motion.div 
              key={index} 
              variants={fadeInUp}
              className="flex flex-col items-center text-center p-6 rounded-xl bg-slate-50 dark:bg-surface-dark/50 border border-slate-100 dark:border-border-dark backdrop-blur-sm"
            >
              <div className="w-16 h-16 rounded-full bg-primary/10 flex items-center justify-center text-primary mb-4 border border-primary/20">
                <span className="material-symbols-outlined !text-4xl">{item.icon}</span>
              </div>
              <h3 className="text-slate-900 dark:text-white font-bold text-lg mb-2 font-display">{item.label}</h3>
              <p className="text-sm text-slate-600 dark:text-text-secondary">{item.desc}</p>
            </motion.div>
          ))}
        </motion.div>
      </div>
    </section>
  );
};

export default Security;
