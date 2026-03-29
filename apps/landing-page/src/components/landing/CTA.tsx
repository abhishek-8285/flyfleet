'use client';

import { useEffect, useRef } from 'react';
import { motion, useMotionValue, useSpring } from 'framer-motion';
import React from 'react';

// Magnetic button component
const MagneticButton: React.FC<{
  onClick: () => void;
  children: React.ReactNode;
  variant?: 'primary' | 'secondary';
  className?: string;
}> = ({ onClick, children, variant = 'primary', className = '' }) => {
  const x = useMotionValue(0);
  const y = useMotionValue(0);
  const springConfig = { damping: 15, stiffness: 150 };
  const xSpring = useSpring(x, springConfig);
  const ySpring = useSpring(y, springConfig);

  const handleMouseMove = (e: React.MouseEvent<HTMLButtonElement>) => {
    const rect = e.currentTarget.getBoundingClientRect();
    const centerX = rect.left + rect.width / 2;
    const centerY = rect.top + rect.height / 2;
    const mouseX = e.clientX - centerX;
    const mouseY = e.clientY - centerY;
    x.set(mouseX * 0.15);
    y.set(mouseY * 0.15);
  };

  const handleMouseLeave = () => {
    x.set(0);
    y.set(0);
  };

  const baseStyles = "h-12 px-8 rounded-xl font-bold transition-all flex items-center justify-center gap-2";
  const variantStyles = variant === 'primary'
    ? "bg-primary hover:bg-primary/90 text-white shadow-lg shadow-primary/30"
    : "border-2 border-slate-300 dark:border-border-dark bg-transparent hover:bg-slate-50 dark:hover:bg-border-dark text-slate-900 dark:text-white";

  return (
    <motion.button
      onClick={onClick}
      style={{ x: xSpring, y: ySpring }}
      onMouseMove={handleMouseMove}
      onMouseLeave={handleMouseLeave}
      whileTap={{ scale: 0.95 }}
      className={`${baseStyles} ${variantStyles} ${className}`}
    >
      {children}
    </motion.button>
  );
};

const CTA = () => {
  const sectionRef = useRef<HTMLElement>(null);

  useEffect(() => {
    const gsap = (window as any).gsap;
    if (!gsap) return;

    const ctx = gsap.context(() => {
      // Parallax for background glow blobs
      gsap.utils.toArray('.cta-glow').forEach((blob: Element, idx: number) => {
        gsap.to(blob, {
          y: idx === 0 ? -30 : 30,
          x: idx === 0 ? 20 : -20,
          duration: 4 + idx,
          repeat: -1,
          yoyo: true,
          ease: 'sine.inOut'
        });
      });

      // Section entrance
      gsap.fromTo(sectionRef.current,
        { opacity: 0, y: 100 },
        {
          opacity: 1,
          y: 0,
          duration: 1,
          scrollTrigger: {
            trigger: sectionRef.current,
            start: 'top 80%',
            toggleActions: 'play none none reverse'
          }
        }
      );
    }, sectionRef);

    return () => ctx.revert();
  }, []);

  return (
    <section
      ref={sectionRef}
      className="scene relative"
      style={{ '--scene-height': '150vh', paddingTop: '0', paddingBottom: '0' } as React.CSSProperties}
      data-scene="cta"
    >
      {/* Background depth layers */}
      <div className="depth-0 absolute inset-0 bg-white dark:bg-background-dark" />
      <div className="depth-1 absolute inset-0 pointer-events-none">
        <div className="cta-glow absolute top-0 right-0 w-96 h-96 bg-primary/10 rounded-full blur-3xl -translate-y-1/2 translate-x-1/2"></div>
        <div className="cta-glow absolute bottom-0 left-0 w-96 h-96 bg-blue-600/10 rounded-full blur-3xl translate-y-1/2 -translate-x-1/2"></div>
      </div>

      {/* Content */}
      <div className="relative z-10 max-w-5xl mx-auto px-4 md:px-10 py-24 flex items-center">
        <motion.div
          initial={{ opacity: 0, y: 50 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="depth-3 w-full rounded-3xl bg-gradient-to-br from-white to-slate-50 dark:from-surface-dark dark:to-background-dark border border-slate-200 dark:border-border-dark p-10 md:p-16 text-center relative overflow-hidden shadow-2xl dark:shadow-none"
        >
          {/* Inner decorative blobs */}
          <div className="absolute top-0 right-0 w-64 h-64 bg-primary/5 dark:bg-primary/10 rounded-full blur-3xl -translate-y-1/2 translate-x-1/2"></div>
          <div className="absolute bottom-0 left-0 w-64 h-64 bg-blue-600/5 dark:bg-blue-600/10 rounded-full blur-3xl translate-y-1/2 -translate-x-1/2"></div>

          <div className="relative z-10">
            <h2 className="text-3xl md:text-5xl font-bold text-slate-900 dark:text-white mb-6 font-display">
              Ready to modernize your fleet?
            </h2>
            <p className="text-slate-600 dark:text-text-secondary text-lg mb-10 max-w-2xl mx-auto">
              Join over 2,000 companies optimizing their operations with FlyFleet today. Start your 14-day free trial — no credit card required.
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <MagneticButton onClick={() => document.getElementById('contact')?.scrollIntoView({ behavior: 'smooth' })}>
                Start Free Trial
                <span className="material-symbols-outlined !text-lg">rocket_launch</span>
              </MagneticButton>
              <MagneticButton variant="secondary" onClick={() => document.getElementById('contact')?.scrollIntoView({ behavior: 'smooth' })}>
                Talk to Sales
                <span className="material-symbols-outlined !text-lg">headset_mic</span>
              </MagneticButton>
            </div>

            {/* Trust indicators */}
            <motion.div
              initial={{ opacity: 0 }}
              whileInView={{ opacity: 1 }}
              viewport={{ once: true }}
              transition={{ delay: 0.5, duration: 0.6 }}
              className="mt-12 pt-8 border-t border-slate-200 dark:border-border-dark"
            >
              <p className="text-sm text-slate-500 mb-4">Trusted by industry leaders</p>
              <div className="flex justify-center items-center gap-8 flex-wrap opacity-50">
                {/* Placeholder company logos - replace with actual SVGs */}
                {['TechCorp', 'LogiNet', 'FleetPro', 'TranspoX', 'DriveSmart'].map((company, idx) => (
                  <div key={idx} className="text-slate-400 font-bold text-sm tracking-wider">
                    {company}
                  </div>
                ))}
              </div>
            </motion.div>
          </div>
        </motion.div>
      </div>
    </section>
  );
};

export default CTA;
