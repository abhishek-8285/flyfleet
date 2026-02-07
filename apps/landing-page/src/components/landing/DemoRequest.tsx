'use client';

import { motion } from 'framer-motion';

const DemoRequest = () => {
  return (
    <section className="py-24 px-4 md:px-10" id="contact">
      <div className="max-w-[1280px] mx-auto">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
          <motion.div 
            initial={{ opacity: 0, y: 30 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6 }}
          >
            <h2 className="text-4xl md:text-5xl font-black text-slate-900 dark:text-white mb-8 font-display leading-[1.1]">
              Schedule a <span className="text-primary">Custom Demo</span>
            </h2>
            <p className="text-slate-600 dark:text-text-secondary text-lg mb-10">
              See how FlyFleet can be tailored to your specific logistics challenges. Our experts will walk you through a setup that fits your fleet's unique requirements.
            </p>
            <div className="space-y-4">
              <div className="flex items-center gap-4 text-slate-700 dark:text-slate-300">
                <div className="size-6 rounded-full bg-green-500/10 flex items-center justify-center text-green-500">
                  <span className="material-symbols-outlined !text-sm">check</span>
                </div>
                <span className="font-bold">Personalized walkthrough</span>
              </div>
              <div className="flex items-center gap-4 text-slate-700 dark:text-slate-300">
                <div className="size-6 rounded-full bg-green-500/10 flex items-center justify-center text-green-500">
                  <span className="material-symbols-outlined !text-sm">check</span>
                </div>
                <span className="font-bold">Custom ROI analysis</span>
              </div>
              <div className="flex items-center gap-4 text-slate-700 dark:text-slate-300">
                <div className="size-6 rounded-full bg-green-500/10 flex items-center justify-center text-green-500">
                  <span className="material-symbols-outlined !text-sm">check</span>
                </div>
                <span className="font-bold">Integration consultation</span>
              </div>
            </div>
          </motion.div>
          <motion.div 
            initial={{ opacity: 0, scale: 0.95 }}
            whileInView={{ opacity: 1, scale: 1 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6, delay: 0.2 }}
            className="bg-white dark:bg-surface-dark p-8 md:p-10 rounded-[32px] border border-slate-200 dark:border-border-dark shadow-2xl relative overflow-hidden"
          >
            <div className="absolute top-0 right-0 w-32 h-32 bg-primary/5 rounded-full -translate-y-1/2 translate-x-1/2 blur-2xl"></div>
            <form className="relative z-10 space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="space-y-2">
                  <label className="text-sm font-bold text-slate-700 dark:text-slate-300">First Name</label>
                  <input type="text" className="w-full h-12 px-4 rounded-xl border border-slate-200 dark:border-border-dark bg-slate-50 dark:bg-background-dark text-slate-900 dark:text-white outline-none focus:border-primary transition-colors" placeholder="John" />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-bold text-slate-700 dark:text-slate-300">Last Name</label>
                  <input type="text" className="w-full h-12 px-4 rounded-xl border border-slate-200 dark:border-border-dark bg-slate-50 dark:bg-background-dark text-slate-900 dark:text-white outline-none focus:border-primary transition-colors" placeholder="Doe" />
                </div>
              </div>
              <div className="space-y-2">
                <label className="text-sm font-bold text-slate-700 dark:text-slate-300">Work Email</label>
                <input type="email" className="w-full h-12 px-4 rounded-xl border border-slate-200 dark:border-border-dark bg-slate-50 dark:bg-background-dark text-slate-900 dark:text-white outline-none focus:border-primary transition-colors" placeholder="john@company.com" />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-bold text-slate-700 dark:text-slate-300">Fleet Size</label>
                <select className="w-full h-12 px-4 rounded-xl border border-slate-200 dark:border-border-dark bg-slate-50 dark:bg-background-dark text-slate-900 dark:text-white outline-none focus:border-primary transition-colors">
                  <option>1-50 Vehicles</option>
                  <option>51-200 Vehicles</option>
                  <option>200+ Vehicles</option>
                </select>
              </div>
              <button className="w-full h-14 bg-primary hover:bg-primary/90 text-white font-black rounded-xl shadow-lg shadow-primary/30 transition-all flex items-center justify-center gap-3">
                Request Demo Call
                <span className="material-symbols-outlined !text-xl">calendar_month</span>
              </button>
              <p className="text-[10px] text-center text-slate-500">By clicking, you agree to our Privacy Policy and Terms of Service.</p>
            </form>
          </motion.div>
        </div>
      </div>
    </section>
  );
};

export default DemoRequest;
