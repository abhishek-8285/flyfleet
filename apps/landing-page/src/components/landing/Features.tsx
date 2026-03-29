'use client';

import { useEffect, useRef } from 'react';
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
  const sectionRef = useRef<HTMLElement>(null);

  useEffect(() => {
    const gsap = (window as any).gsap;
    const ScrollTrigger = (window as any).ScrollTrigger;
    if (!gsap) return;

    const ctx = gsap.context(() => {
      // Parallax depth for feature cards - different speeds based on index
      gsap.utils.toArray('.feature-card').forEach((card: Element, idx: number) => {
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

      // Entrance animation: cards stack and then cascade
      if (ScrollTrigger) {
        gsap.utils.toArray('.feature-card').forEach((card: Element, idx: number) => {
          gsap.from(card, {
            opacity: 0,
            y: 100,
            rotation: (idx % 2 === 0 ? -5 : 5),
            duration: 0.8,
            delay: idx * 0.1,
            ease: 'power3.out',
            scrollTrigger: {
              trigger: sectionRef.current,
              start: 'top 60%',
              toggleActions: 'play none none reverse'
            }
          });
        });
      }
    }, sectionRef);

    return () => ctx.revert();
  }, []);

  return (
    <section
      ref={sectionRef}
      className="scene relative"
      style={{ '--scene-height': '200vh', paddingTop: '0', paddingBottom: '0' } as React.CSSProperties}
      data-scene="features"
    >
      {/* DEPTH 0: Background */}
      <div className="depth-0 absolute inset-0 bg-white dark:bg-background-dark" />

      {/* DEPTH 1: Subtle atmospheric glow */}
      <div className="depth-1 absolute inset-0 pointer-events-none">
        <div className="absolute top-0 left-1/4 w-96 h-96 bg-primary/5 rounded-full blur-3xl"></div>
        <div className="absolute bottom-0 right-1/4 w-96 h-96 bg-blue-400/5 rounded-full blur-3xl"></div>
      </div>

      {/* Content */}
      <div className="relative z-10 max-w-[1280px] mx-auto px-4 md:px-10 py-24">
        <motion.div {...fadeInUp} className="text-center mb-16 max-w-3xl mx-auto">
          <h2 className="text-3xl md:text-4xl font-bold text-slate-900 dark:text-white mb-4 font-display">Complete Fleet Control</h2>
          <p className="text-slate-600 dark:text-text-secondary text-lg">A comprehensive suite designed for data-driven decisions.</p>
        </motion.div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {FEATURES.map((feature, idx) => (
            <motion.div
              key={idx}
              variants={fadeInUp}
              whileHover={{ scale: 1.03, y: -5 }}
              className="feature-card depth-2 group p-6 rounded-xl border border-slate-200 dark:border-border-dark bg-white dark:bg-surface-dark hover:border-primary/50 transition-all hover:shadow-xl hover:shadow-primary/10"
            >
              <div className="w-12 h-12 rounded-lg bg-primary/10 flex items-center justify-center text-primary mb-4 group-hover:bg-primary group-hover:text-white transition-colors duration-300">
                <span className="material-symbols-outlined !text-3xl">{feature.icon}</span>
              </div>
              <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-2 font-display">{feature.title}</h3>
              <p className="text-slate-600 dark:text-text-secondary text-sm">{feature.desc}</p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
};

export default Features;
