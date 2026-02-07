'use client';

import { AnimatePresence, motion } from 'framer-motion';
import { useState } from 'react';

const FAQS = [
  { q: "How difficult is installation?", a: "Extremely simple. Our plug-and-play OBD-II devices install in seconds." },
  { q: "Is my data secure?", a: "We are ISO 27001 certified and GDPR compliant. All data is encrypted." },
  { q: "What if a driver unplugs the device?", a: "You will receive an instant 'Device Unplugged' alert via SMS and email." },
  { q: "What support options are available?", a: "All plans include email support. Pro gets chat, Enterprise gets 24/7 phone." }
];

const FAQ = () => {
  const [activeFaq, setActiveFaq] = useState<number | null>(null);

  return (
    <section className="py-20 px-4 md:px-10 bg-slate-50 dark:bg-background-dark/50">
      <div className="max-w-4xl mx-auto">
        <h2 className="text-3xl font-bold text-center text-slate-900 dark:text-white mb-12 font-display">Frequently Asked Questions</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {FAQS.map((faq, i) => (
            <div key={i} className="group bg-white dark:bg-surface-dark rounded-lg border border-slate-200 dark:border-border-dark overflow-hidden">
              <button 
                className="w-full flex justify-between items-center p-4 font-bold text-left text-slate-900 dark:text-white"
                onClick={() => setActiveFaq(activeFaq === i ? null : i)}
              >
                {faq.q}
                <motion.span 
                  animate={{ rotate: activeFaq === i ? 180 : 0 }}
                  className="material-symbols-outlined"
                >
                  expand_more
                </motion.span>
              </button>
              <AnimatePresence>
                {activeFaq === i && (
                  <motion.div 
                    initial={{ height: 0, opacity: 0 }}
                    animate={{ height: "auto", opacity: 1 }}
                    exit={{ height: 0, opacity: 0 }}
                    className="px-4 pb-4 overflow-hidden"
                  >
                    <p className="text-slate-600 dark:text-text-secondary text-sm leading-relaxed">{faq.a}</p>
                  </motion.div>
                )}
              </AnimatePresence>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
};

export default FAQ;
