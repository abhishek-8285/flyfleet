'use client';

import { AnimatePresence, motion } from 'framer-motion';
import Link from 'next/link';
import { useEffect, useState } from 'react';

const CookieConsent = () => {
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    const consent = localStorage.getItem('cookie-consent');
    if (!consent) {
      const timer = setTimeout(() => setIsVisible(true), 2000);
      return () => clearTimeout(timer);
    }
  }, []);

  const handleAccept = () => {
    localStorage.setItem('cookie-consent', 'accepted');
    setIsVisible(false);
  };

  const handleDecline = () => {
    localStorage.setItem('cookie-consent', 'declined');
    setIsVisible(false);
  };

  return (
    <AnimatePresence>
      {isVisible && (
        <motion.div 
          initial={{ y: 100, opacity: 0 }}
          animate={{ y: 0, opacity: 1 }}
          exit={{ y: 100, opacity: 0 }}
          className="fixed bottom-6 left-6 right-6 md:left-auto md:right-8 z-[70] md:max-w-md"
        >
          <div className="bg-white dark:bg-surface-dark p-6 rounded-2xl border border-slate-200 dark:border-border-dark shadow-2xl backdrop-blur-xl bg-white/90 dark:bg-surface-dark/90">
            <div className="flex gap-4 items-start mb-4">
              <div className="size-10 rounded-full bg-primary/10 flex items-center justify-center text-primary flex-shrink-0">
                <span className="material-symbols-outlined">cookie</span>
              </div>
              <div>
                <h4 className="font-bold text-slate-900 dark:text-white">Cookie Preferences</h4>
                <p className="text-xs text-slate-500 dark:text-text-secondary leading-relaxed mt-1">
                  We use cookies to improve your experience and analyze our traffic. By clicking "Accept All", you consent to our use of cookies. Read our <Link href="/privacy" className="text-primary hover:underline">Privacy Policy</Link>.
                </p>
              </div>
            </div>
            <div className="flex gap-3">
              <button 
                onClick={handleAccept}
                className="flex-1 h-10 bg-primary text-white text-xs font-bold rounded-lg hover:bg-primary/90 transition-colors"
              >
                Accept All
              </button>
              <button 
                onClick={handleDecline}
                className="h-10 px-4 border border-slate-200 dark:border-border-dark text-slate-600 dark:text-slate-300 text-xs font-bold rounded-lg hover:bg-slate-50 dark:hover:bg-background-dark transition-colors"
              >
                Decline
              </button>
            </div>
          </div>
        </motion.div>
      )}
    </AnimatePresence>
  );
};

export default CookieConsent;
