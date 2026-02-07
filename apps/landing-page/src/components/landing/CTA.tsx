'use client';

import { motion } from 'framer-motion';

const fadeInUp = {
  initial: { opacity: 0, y: 30 },
  whileInView: { opacity: 1, y: 0 },
  viewport: { once: true },
  transition: { duration: 0.6 }
};

const CTA = () => {
  return (
    <motion.section 
      {...fadeInUp}
      className="py-20 px-4 md:px-10"
    >
      <div className="max-w-5xl mx-auto rounded-3xl bg-gradient-to-br from-white to-slate-50 dark:from-surface-dark dark:to-background-dark border border-slate-200 dark:border-border-dark p-10 md:p-16 text-center relative overflow-hidden shadow-2xl dark:shadow-none">
        <div className="absolute top-0 right-0 w-64 h-64 bg-primary/5 dark:bg-primary/10 rounded-full blur-3xl -translate-y-1/2 translate-x-1/2"></div>
        <div className="absolute bottom-0 left-0 w-64 h-64 bg-blue-600/5 dark:bg-blue-600/10 rounded-full blur-3xl translate-y-1/2 -translate-x-1/2"></div>
        <h2 className="relative z-10 text-3xl md:text-5xl font-bold text-slate-900 dark:text-white mb-6 font-display">Ready to modernize your fleet?</h2>
        <p className="relative z-10 text-slate-600 dark:text-text-secondary text-lg mb-10 max-w-2xl mx-auto">Join over 2,000 companies optimizing their operations with FlyFleet today.</p>
        <div className="relative z-10 flex flex-col sm:flex-row gap-4 justify-center">
          <button className="h-12 px-8 rounded-lg bg-primary hover:bg-primary/90 text-white text-base font-bold transition-all shadow-lg shadow-primary/25">Start Free Trial</button>
          <button className="h-12 px-8 rounded-lg border border-slate-300 dark:border-border-dark bg-transparent hover:bg-slate-100 dark:hover:bg-border-dark text-slate-900 dark:text-white text-base font-bold transition-all">Talk to Sales</button>
        </div>
      </div>
    </motion.section>
  );
};

export default CTA;
