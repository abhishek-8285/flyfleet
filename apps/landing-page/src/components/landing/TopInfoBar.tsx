'use client';

import { AnimatePresence, motion } from 'framer-motion';
import React from 'react';
import { Language } from './translations';

interface TopInfoBarProps {
  phone: string;
  language: string;
  onLanguageChange: (lang: string) => void;
  showLangDropdown: boolean;
  setShowLangDropdown: (show: boolean) => void;
}

const TopInfoBar: React.FC<TopInfoBarProps> = ({ 
  phone, 
  language, 
  onLanguageChange, 
  showLangDropdown, 
  setShowLangDropdown 
}) => {
  const languages: Language[] = ['US (EN)', 'DE (DE)', 'IN (HI)'];

  return (
    <div className="fixed bottom-0 w-full z-[60] bg-primary text-white text-sm font-medium h-6 flex items-center justify-center shadow-md">
      <div className="max-w-[1440px] w-full px-4 md:px-10 flex justify-between items-center h-full">
        <div className="flex items-center gap-6 overflow-hidden">
          <div className="flex items-center gap-2 flex-shrink-0">
             <span className="material-symbols-outlined !text-sm">verified_user</span>
             <span className="hidden sm:inline">ISO 27001 Certified System</span>
             <span className="sm:hidden">ISO 27001</span>
          </div>
          <div className="h-4 w-px bg-white/20 hidden md:block"></div>
          <div className="hidden md:flex items-center gap-2">
             <span className="material-symbols-outlined !text-sm">support_agent</span>
             <span>Enterprise Support: <a href={`tel:${phone}`} className="font-bold underline">{phone}</a></span>
          </div>
        </div>
        
        <div className="relative">
          <button 
            onClick={() => setShowLangDropdown(!showLangDropdown)}
            className="flex items-center gap-2 bg-white/10 hover:bg-white/20 px-3 py-1 rounded transition-colors"
          >
            <span className="material-symbols-outlined !text-sm">public</span>
            {language}
            <span className="material-symbols-outlined !text-xs">expand_less</span>
          </button>

          <AnimatePresence>
            {showLangDropdown && (
              <motion.div 
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: 10 }}
                className="absolute right-0 bottom-full mb-2 w-40 bg-white dark:bg-surface-dark rounded-lg shadow-2xl border border-slate-200 dark:border-border-dark overflow-hidden z-50 py-1"
              >
                {languages.map((lang) => (
                  <button
                    key={lang}
                    onClick={() => {
                      onLanguageChange(lang);
                      setShowLangDropdown(false);
                    }}
                    className={`w-full text-left px-4 py-2 text-xs font-bold hover:bg-slate-50 dark:hover:bg-background-dark transition-colors ${language === lang ? 'text-primary' : 'text-slate-700 dark:text-slate-300'}`}
                  >
                    {lang}
                  </button>
                ))}
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      </div>
    </div>
  );
};

export default TopInfoBar;
