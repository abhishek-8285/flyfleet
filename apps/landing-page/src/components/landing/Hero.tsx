'use client';

import { useEffect, useRef, useState } from 'react';
import { motion, useMotionValue, useSpring } from 'framer-motion';
import React from 'react';
import { Language, translations } from './translations';

// Magnetic button component
const MagneticButton: React.FC<{
  onClick: () => void;
  children: React.ReactNode;
  className?: string;
}> = ({ onClick, children, className = '' }) => {
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
    x.set(mouseX * 0.2);
    y.set(mouseY * 0.2);
  };

  const handleMouseLeave = () => {
    x.set(0);
    y.set(0);
  };

  return (
    <motion.button
      onClick={onClick}
      style={{ x: xSpring, y: ySpring }}
      onMouseMove={handleMouseMove}
      onMouseLeave={handleMouseLeave}
      whileTap={{ scale: 0.98 }}
      className={`h-12 px-8 rounded-lg bg-primary hover:bg-primary/90 text-white text-base font-bold transition-all shadow-lg shadow-primary/30 flex items-center justify-center gap-2 ${className}`}
    >
      {children}
    </motion.button>
  );
};

interface HeroProps {
  onDemoRequest: () => void;
  onTourRequest: () => void;
  language?: string;
}

const Hero: React.FC<HeroProps> = ({ onDemoRequest, onTourRequest, language = 'US (EN)' }) => {
  const t = translations[language as Language]?.hero || translations['US (EN)'].hero;
  const heroRef = useRef<HTMLElement>(null);
  const truckRef = useRef<HTMLDivElement>(null);
  const headlineRef = useRef<HTMLHeadingElement>(null);

  // Split headline into words for split-text animation
  const headlineWords = t.title.split(' ');

  useEffect(() => {
    // Check if GSAP is available
    const gsap = (window as any).gsap;
    const ScrollTrigger = (window as any).ScrollTrigger;
    if (!gsap) return;

    const ctx = gsap.context(() => {
      // Parallax for depth layers
      gsap.utils.toArray('.depth-0').forEach((layer: Element) => {
        gsap.to(layer, {
          y: -150,
          ease: 'none',
          scrollTrigger: {
            trigger: heroRef.current,
            start: 'top bottom',
            end: 'bottom top',
            scrub: 1
          }
        });
      });

      gsap.utils.toArray('.depth-1').forEach((layer: Element) => {
        gsap.to(layer, {
          y: -250,
          ease: 'none',
          scrollTrigger: {
            trigger: heroRef.current,
            start: 'top bottom',
            end: 'bottom top',
            scrub: 1
          }
        });
      });

      gsap.utils.toArray('.depth-2').forEach((layer: Element) => {
        gsap.to(layer, {
          y: -200,
          ease: 'none',
          scrollTrigger: {
            trigger: heroRef.current,
            start: 'top bottom',
            end: 'bottom top',
            scrub: 1
          }
        });
      });

      gsap.utils.toArray('.depth-3').forEach((layer: Element) => {
        gsap.to(layer, {
          y: -250,
          scale: 1.15,
          ease: 'none',
          scrollTrigger: {
            trigger: heroRef.current,
            start: 'top bottom',
            end: 'bottom top',
            scrub: 1
          }
        });
      });

      gsap.utils.toArray('.depth-4').forEach((layer: Element) => {
        gsap.to(layer, {
          y: -300,
          opacity: 0,
          ease: 'none',
          scrollTrigger: {
            trigger: heroRef.current,
            start: 'top bottom',
            end: 'bottom top',
            scrub: 1
          }
        });
      });

      // Hero truck float animation (persistent)
      if (truckRef.current) {
        gsap.to(truckRef.current, {
          y: '-=30',
          duration: 4,
          repeat: -1,
          yoyo: true,
          ease: 'sine.inOut'
        });
      }

      // Split-text entrance animation for headline
      if (headlineRef.current) {
        const words = headlineRef.current.querySelectorAll('.split-word');
        gsap.from(Array.from(words), {
          opacity: 0,
          x: (index: number) => (index % 2 === 0 ? -100 : 100),
          y: 50,
          duration: 1,
          stagger: 0.1,
          ease: 'power3.out',
          delay: 0.3
        });
      }

      // Cleanup will-change after animations settle
      if (ScrollTrigger) {
        ScrollTrigger.addEventListener('refresh', () => {
          document.querySelectorAll('[will-change]').forEach(el => {
            el.setAttribute('data-has-willchange', 'true');
          });
        });
      }

    }, heroRef);

    return () => ctx.revert();
  }, []);

  // Early return if not client-side
  if (typeof window === 'undefined') {
    return (
      <section className="relative w-full py-16 md:py-24 overflow-hidden mt-10">
        <div className="glow-effect"></div>
        <div className="px-4 md:px-10 max-w-[1280px] mx-auto flex flex-col items-center text-center relative z-10">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-blue-500/10 border border-blue-500/20 text-blue-500 text-xs font-bold uppercase tracking-wide mb-6">
            {t.tag}
          </div>
          <h1 className="text-4xl md:text-6xl lg:text-7xl font-black leading-tight tracking-tight mb-6 max-w-5xl text-slate-900 dark:text-white font-display">
            {t.title}
          </h1>
          <p className="text-slate-600 dark:text-text-secondary text-lg md:text-xl max-w-2xl mb-10 leading-relaxed">
            {t.desc}
          </p>
          <div className="flex flex-col sm:flex-row gap-4 mb-16">
            <button className="h-12 px-8 rounded-lg bg-primary hover:bg-primary/90 text-white text-base font-bold transition-all shadow-lg shadow-primary/30 flex items-center justify-center gap-2">
              {t.demo}
              <span className="material-symbols-outlined !text-lg">arrow_forward</span>
            </button>
            <button className="h-12 px-8 rounded-lg border border-slate-300 dark:border-border-dark bg-white dark:bg-surface-dark hover:bg-slate-50 dark:hover:bg-border-dark text-slate-900 dark:text-white text-base font-bold transition-all flex items-center justify-center gap-2 group">
              <span className="material-symbols-outlined !text-lg text-primary group-hover:scale-110 transition-transform">play_circle</span>
              {t.tour}
            </button>
          </div>
        </div>
      </section>
    );
  }

  return (
    <section
      ref={heroRef}
      className="scene relative w-full overflow-visible"
      style={{ '--scene-height': '200vh' } as React.CSSProperties}
      data-scene="hero"
    >
      {/* DEPTH 0: Background gradient mesh (blurred) */}
      <div className="depth-0 layer">
        <div className="absolute inset-0 bg-gradient-to-b from-slate-50 via-blue-50/30 to-slate-100 dark:from-background-dark dark:via-primary/5 dark:to-background-dark" />
        {/* Subtle grid pattern */}
        <div className="absolute inset-0 opacity-[0.03] dark:opacity-[0.05]" style={{
          backgroundImage: 'radial-gradient(#137fec 1px, transparent 1px)',
          backgroundSize: '30px 30px'
        }}></div>
      </div>

      {/* DEPTH 1: Atmospheric glow blobs */}
      <div className="depth-1 layer pointer-events-none">
        <div className="absolute w-96 h-96 bg-primary/10 rounded-full blur-3xl -top-20 -left-20 mix-blend-multiply animate-pulse"></div>
        <div className="absolute w-96 h-96 bg-blue-400/10 rounded-full blur-3xl -bottom-20 -right-20 mix-blend-multiply animate-pulse" style={{ animationDelay: '2s' }}></div>
        <div className="absolute w-64 h-64 bg-cyan-300/10 rounded-full blur-3xl top-1/3 left-1/3 mix-blend-multiply animate-pulse" style={{ animationDelay: '4s' }}></div>
      </div>

      {/* DEPTH 2: Companion floating trucks (smaller) */}
      <div className="depth-2 layer pointer-events-none">
        {/* Companion truck - top left */}
        <motion.div
          className="absolute left-[10%] top-[20%] text-primary/20"
          initial={{ scale: 0 }}
          animate={{ scale: 1 }}
          transition={{ delay: 0.5, duration: 0.8, ease: [0.68, -0.55, 0.265, 1.55] as any }}
        >
          <span className="material-symbols-outlined text-8xl">local_shipping</span>
        </motion.div>

        {/* Companion truck - bottom right */}
        <motion.div
          className="absolute right-[10%] bottom-[30%] text-primary/20"
          initial={{ scale: 0 }}
          animate={{ scale: 1 }}
          transition={{ delay: 0.7, duration: 0.8, ease: [0.68, -0.55, 0.265, 1.55] as any }}
        >
          <span className="material-symbols-outlined text-9xl">local_shipping</span>
        </motion.div>
      </div>

      {/* DEPTH 3: HERO TRUCK (main product) */}
      <div className="depth-3 layer pointer-events-none">
        <motion.div
          ref={truckRef}
          className="relative"
          initial={{ opacity: 0, y: -200 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2, duration: 1.2 }}
        >
          {/* Large hero truck icon with glow */}
          <div className="relative">
            <div className="absolute inset-0 flex items-center justify-center">
              <span className="material-symbols-outlined text-[12rem] md:text-[16rem] text-primary/5">local_shipping</span>
            </div>
            <span className="material-symbols-outlined text-[10rem] md:text-[14rem] text-primary relative z-10 drop-shadow-2xl" style={{ filter: 'drop-shadow(0 25px 50px rgba(19,127,236,0.3))' }}>
              local_shipping
            </span>
          </div>

          {/* Floating label */}
          <motion.div
            className="absolute -bottom-8 left-1/2 -translate-x-1/2 px-4 py-2 bg-white/80 dark:bg-surface-dark/80 backdrop-blur-sm rounded-full border border-primary/20 shadow-lg"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.8 }}
          >
            <span className="text-xs font-bold text-primary uppercase tracking-wider">Introducing</span>
          </motion.div>
        </motion.div>
      </div>

      {/* DEPTH 4: Text content and CTA */}
      <div className="depth-4 layer px-4 md:px-10 max-w-[1280px] mx-auto flex flex-col items-center text-center">
        {/* Tag badge */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.4 }}
          className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-primary/10 border border-primary/30 text-primary text-xs font-bold uppercase tracking-widest mb-6"
        >
          <span className="relative flex h-2 w-2">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-primary opacity-75"></span>
            <span className="relative inline-flex rounded-full h-2 w-2 bg-primary"></span>
          </span>
          {t.tag}
        </motion.div>

        {/* Headline with split-word animation container */}
        <h1
          ref={headlineRef}
          className="text-4xl md:text-6xl lg:text-7xl font-black leading-tight tracking-tight mb-6 max-w-5xl text-slate-900 dark:text-white font-display"
        >
          {headlineWords.map((word, i) => (
            <span key={i} className="split-word inline-block mr-2 mb-2">{word}</span>
          ))}
        </h1>

        {/* Description */}
        <motion.p
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.6 }}
          className="text-slate-600 dark:text-text-secondary text-lg md:text-xl max-w-2xl mb-10 leading-relaxed"
        >
          {t.desc}
        </motion.p>

        {/* CTA buttons */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.8 }}
          className="flex flex-col sm:flex-row gap-4 mb-16"
        >
          <MagneticButton onClick={onDemoRequest}>
            {t.demo}
            <span className="material-symbols-outlined !text-lg">arrow_forward</span>
          </MagneticButton>
          <motion.button
            onClick={onTourRequest}
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.98 }}
            className="h-12 px-8 rounded-lg border-2 border-primary/30 dark:border-border-dark bg-white/80 dark:bg-surface-dark hover:bg-primary/10 dark:hover:bg-border-dark text-slate-900 dark:text-white text-base font-bold transition-all flex items-center justify-center gap-2 group"
          >
            <span className="material-symbols-outlined !text-xl text-primary group-hover:scale-125 transition-transform duration-300">play_circle</span>
            {t.tour}
          </motion.button>
        </motion.div>
      </div>

      {/* DEPTH 5: Foreground particles */}
      <div className="depth-5 layer pointer-events-none">
        {/* Sparkle particles */}
        {[...Array(8)].map((_, i) => (
          <motion.div
            key={i}
            className="absolute w-1 h-1 bg-white rounded-full shadow-[0_0_10px_rgba(255,255,255,0.8)]"
            style={{
              left: `${10 + i * 12}%`,
              top: `${20 + (i % 3) * 25}%`
            }}
            animate={{
              opacity: [0.3, 1, 0.3],
              scale: [0.8, 1.2, 0.8]
            }}
            transition={{
              duration: 3 + i * 0.5,
              repeat: Infinity,
              delay: i * 0.2
            }}
          />
        ))}
      </div>
    </section>
  );
};

export default Hero;
