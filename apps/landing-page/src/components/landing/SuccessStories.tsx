'use client';

import { motion } from 'framer-motion';

const SUCCESS_STORIES = [
  {
    icon: 'trending_down', value: '24%', title: 'Reduced Fuel Costs',
    quote: "Global Logistics implemented our fuel monitoring solution across 450 trucks, identifying theft and idling inefficiencies within week one.",
    company: 'Global Logistics Inc.', fleet: 'Fleet of 450+ Vehicles', badge: 'GL'
  },
  {
    icon: 'security', value: '40%', title: 'Fewer Accidents',
    quote: "By utilizing the Driver Scorecard module, CityTrans gamified safety, resulting in a massive drop in harsh braking events and insurance claims.",
    company: 'CityTrans Delivery', fleet: 'Fleet of 120+ Vans', badge: 'CT'
  },
  {
    icon: 'schedule', value: '15hr', title: 'Weekly Admin Saved',
    quote: "Automating trip classification and expense reports saved the finance team 15 hours per week on manual data entry.",
    company: 'BuildCorp Construction', fleet: 'Heavy Equipment Fleet', badge: 'BC'
  }
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

const SuccessStories = () => {
  return (
    <section className="py-20 px-4 md:px-10" id="success-stories">
      <div className="max-w-[1280px] mx-auto">
        <motion.div {...fadeInUp} className="mb-16 md:text-center max-w-3xl mx-auto">
          <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4 font-display">Enterprise Success Stories</h2>
          <p className="text-slate-600 dark:text-text-secondary text-lg">Real results from companies scaling their operations with FlyFleet.</p>
        </motion.div>
        <motion.div 
          variants={staggerContainer}
          initial="initial"
          whileInView="whileInView"
          viewport={{ once: true }}
          className="grid grid-cols-1 md:grid-cols-3 gap-8"
        >
          {SUCCESS_STORIES.map((story, i) => (
            <motion.div 
              key={i} 
              variants={fadeInUp}
              whileHover={{ y: -10 }}
              className="relative group rounded-2xl bg-white dark:bg-surface-dark border border-slate-200 dark:border-border-dark p-8 shadow-sm hover:shadow-xl transition-all"
            >
              <div className="text-primary mb-6">
                <span className="material-symbols-outlined !text-5xl">{story.icon}</span>
              </div>
              <div className="text-4xl md:text-5xl font-black text-slate-900 dark:text-white mb-2 font-display">{story.value}</div>
              <h3 className="text-lg font-bold text-slate-700 dark:text-slate-300 mb-6 font-display">{story.title}</h3>
              <p className="text-slate-600 dark:text-text-secondary mb-4 text-sm leading-relaxed">
                "{story.quote}"
              </p>
              <a className="inline-flex items-center gap-2 text-primary text-xs font-bold hover:underline mb-6" href="#">
                <span className="material-symbols-outlined text-sm">picture_as_pdf</span>
                Download PDF Case Study
              </a>
              <div className="flex items-center gap-3 border-t border-slate-100 dark:border-border-dark pt-6">
                <div className="w-10 h-10 rounded-full bg-primary/10 text-primary flex items-center justify-center font-bold text-xs">{story.badge}</div>
                <div>
                  <p className="text-sm font-bold text-slate-900 dark:text-white">{story.company}</p>
                  <p className="text-xs text-slate-500">{story.fleet}</p>
                </div>
              </div>
            </motion.div>
          ))}
        </motion.div>
      </div>
    </section>
  );
};

export default SuccessStories;
