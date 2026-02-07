'use client';

import { motion } from 'framer-motion';

const FEATURES = [
  { icon: 'location_on', title: 'Live Tracking', desc: 'Real-time GPS coordinates with 5-second refresh rates and traffic overlays.' },
  { icon: 'route', title: 'Trip Management', desc: 'Historical route playback, trip classification (business/personal), and stop duration logs.' },
  { icon: 'share_location', title: 'Smart Geofencing', desc: 'Create unlimited custom zones and get instant alerts when vehicles enter or exit specific areas.' },
  { icon: 'water_drop', title: 'Fuel Monitoring', desc: 'Track consumption patterns, detect theft events, and optimize refueling stops.' },
  { icon: 'car_repair', title: 'Maintenance Alerts', desc: 'Automated reminders for oil changes, tire rotations, and insurance renewals.' },
  { icon: 'speed', title: 'Driver Scorecards', desc: 'Gamify safety with driver scores based on speeding, harsh braking, and cornering.' }
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

const Features = () => {
  return (
    <section className="py-20 px-4 md:px-10" id="features">
      <div className="max-w-[1280px] mx-auto">
        <motion.div {...fadeInUp} className="text-center mb-16 max-w-3xl mx-auto">
          <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4 font-display">Complete Fleet Control</h2>
          <p className="text-slate-600 dark:text-text-secondary text-lg">A comprehensive suite designed for data-driven decisions.</p>
        </motion.div>
        <motion.div 
          variants={staggerContainer}
          initial="initial"
          whileInView="whileInView"
          viewport={{ once: true }}
          className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6"
        >
          {FEATURES.map((feature, idx) => (
            <motion.div 
              key={idx} 
              variants={fadeInUp}
              whileHover={{ scale: 1.02 }}
              className="group p-6 rounded-xl border border-slate-200 dark:border-border-dark bg-white dark:bg-surface-dark hover:border-primary/50 transition-all hover:shadow-lg hover:shadow-primary/5"
            >
              <div className="w-12 h-12 rounded-lg bg-primary/10 flex items-center justify-center text-primary mb-4">
                <span className="material-symbols-outlined !text-3xl">{feature.icon}</span>
              </div>
              <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-2 font-display">{feature.title}</h3>
              <p className="text-slate-600 dark:text-text-secondary text-sm">{feature.desc}</p>
            </motion.div>
          ))}
        </motion.div>
      </div>
    </section>
  );
};

export default Features;
