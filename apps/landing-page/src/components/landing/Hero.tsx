'use client';

import { motion } from 'framer-motion';
import React from 'react';
import { Language, translations } from './translations';

interface HeroProps {
  onDemoRequest: () => void;
  onTourRequest: () => void;
  language?: string;
}

const Hero: React.FC<HeroProps> = ({ onDemoRequest, onTourRequest, language = 'US (EN)' }) => {
  const t = translations[language as Language]?.hero || translations['US (EN)'].hero;

  return (
    <section className="relative w-full py-16 md:py-24 overflow-hidden mt-10">
      <div className="glow-effect"></div>
      <div className="px-4 md:px-10 max-w-[1280px] mx-auto flex flex-col items-center text-center relative z-10">
        <motion.div 
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-blue-500/10 border border-blue-500/20 text-blue-500 text-xs font-bold uppercase tracking-wide mb-6"
        >
          <span className="relative flex h-2 w-2">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-500 opacity-75"></span>
            <span className="relative inline-flex rounded-full h-2 w-2 bg-blue-500"></span>
          </span>
          {t.tag}
        </motion.div>
        <motion.h1 
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
          className="text-4xl md:text-6xl lg:text-7xl font-black leading-tight tracking-tight mb-6 max-w-5xl text-slate-900 dark:text-white font-display"
        >
          {t.title}
        </motion.h1>
        <motion.p 
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
          className="text-slate-600 dark:text-text-secondary text-lg md:text-xl max-w-2xl mb-10 leading-relaxed"
        >
          {t.desc}
        </motion.p>
        <motion.div 
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3 }}
          className="flex flex-col sm:flex-row gap-4 mb-16"
        >
          <button 
            onClick={onDemoRequest}
            className="h-12 px-8 rounded-lg bg-primary hover:bg-primary/90 text-white text-base font-bold transition-all shadow-lg shadow-primary/30 flex items-center justify-center gap-2"
          >
            {t.demo}
            <span className="material-symbols-outlined !text-lg">arrow_forward</span>
          </button>
          <button 
            onClick={onTourRequest}
            className="h-12 px-8 rounded-lg border border-slate-300 dark:border-border-dark bg-white dark:bg-surface-dark hover:bg-slate-50 dark:hover:bg-border-dark text-slate-900 dark:text-white text-base font-bold transition-all flex items-center justify-center gap-2 group"
          >
            <span className="material-symbols-outlined !text-lg text-primary group-hover:scale-110 transition-transform">play_circle</span>
            {t.tour}
          </button>
        </motion.div>
      </div>
    </section>
  );
};

export default Hero;
