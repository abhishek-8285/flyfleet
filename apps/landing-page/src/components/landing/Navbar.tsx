'use client';

import Link from 'next/link';
import React from 'react';
import { Language, translations } from './translations';

interface NavbarProps {
  companyName: string;
  isDark: boolean;
  toggleTheme: () => void;
  language?: string;
}

const Navbar: React.FC<NavbarProps> = ({ companyName, isDark, toggleTheme, language = 'US (EN)' }) => {
  const t = translations[language as Language]?.nav || translations['US (EN)'].nav;

  return (
    <header className="fixed top-0 z-50 w-full border-b border-solid border-slate-200 dark:border-border-dark bg-white/80 dark:bg-background-dark/80 backdrop-blur-md transition-all">
      <div className="px-4 md:px-10 flex h-16 items-center justify-between max-w-[1440px] mx-auto">
        <div className="flex items-center gap-4 text-slate-900 dark:text-white">
          <Link href="/" className="flex items-center gap-4" aria-label={`${companyName} Home`}>
            <div className="size-8 flex items-center justify-center text-primary">
              <span className="material-symbols-outlined !text-3xl">local_shipping</span>
            </div>
            <h2 className="text-lg font-bold leading-tight tracking-[-0.015em] font-display">{companyName}</h2>
          </Link>
        </div>
        <nav className="hidden md:flex flex-1 justify-center gap-8 font-display" role="navigation">
          <Link className="text-slate-600 dark:text-text-secondary text-sm font-medium hover:text-primary transition-colors" href="/#success-stories">{t.success}</Link>
          <Link className="text-slate-600 dark:text-text-secondary text-sm font-medium hover:text-primary transition-colors" href="/#features">{t.platform}</Link>
          <Link className="text-slate-600 dark:text-text-secondary text-sm font-medium hover:text-primary transition-colors" href="/#apps">{t.apps}</Link>
          <Link className="text-slate-600 dark:text-text-secondary text-sm font-medium hover:text-primary transition-colors" href="/pricing">{t.pricing}</Link>
          <Link className="text-slate-600 dark:text-text-secondary text-sm font-medium hover:text-primary transition-colors" href="/security">{t.security}</Link>
        </nav>
        <div className="flex gap-4 items-center">
          <Link href="http://localhost:3005" className="hidden sm:flex text-slate-600 dark:text-text-secondary text-sm font-medium hover:text-primary transition-colors">
            {t.login}
          </Link>
          <button onClick={toggleTheme} className="p-2 rounded-full hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-600 dark:text-slate-300 transition-colors">
            <span className="material-symbols-outlined">{isDark ? 'light_mode' : 'dark_mode'}</span>
          </button>
          <button className="flex h-9 items-center justify-center rounded-lg px-4 bg-primary hover:bg-primary/90 text-white text-sm font-bold transition-colors shadow-lg shadow-primary/25">
            <span className="truncate">Start Free Trial</span>
          </button>
        </div>
      </div>
    </header>
  );
};

export default Navbar;
