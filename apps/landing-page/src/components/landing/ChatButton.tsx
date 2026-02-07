'use client';

import { AnimatePresence, motion } from 'framer-motion';
import { useState } from 'react';

const ChatButton = () => {
  const [isExpanded, setIsExpanded] = useState(false);
  return (
    <div className="fixed bottom-20 right-6 z-[60] flex flex-col items-end gap-4">
      <AnimatePresence>
        {isExpanded && (
          <motion.div
            initial={{ opacity: 0, y: 20, scale: 0.95 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: 20, scale: 0.95 }}
            className="bg-white dark:bg-surface-dark border border-slate-200 dark:border-border-dark rounded-2xl shadow-2xl p-4 w-72 mb-2 origin-bottom-right"
          >
            <div className="flex items-center gap-3 mb-4 pb-4 border-b border-slate-100 dark:border-border-dark">
              <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center text-primary">
                <span className="material-symbols-outlined">support_agent</span>
              </div>
              <div>
                <h4 className="font-bold text-slate-900 dark:text-white text-sm">Fleet Expert</h4>
                <p className="text-xs text-green-500 flex items-center gap-1">
                  <span className="w-1.5 h-1.5 rounded-full bg-green-500"></span>
                  Online Now
                </p>
              </div>
            </div>
            
            <div className="space-y-2">
              <a href="#" className="flex items-center gap-3 p-3 rounded-xl hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors group">
                <span className="material-symbols-outlined text-slate-400 group-hover:text-primary transition-colors">chat</span>
                <span className="text-sm font-medium text-slate-700 dark:text-slate-300">Live Chat</span>
              </a>
              <a href="#" className="flex items-center gap-3 p-3 rounded-xl hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors group">
                <span className="material-symbols-outlined text-slate-400 group-hover:text-green-500 transition-colors">call</span>
                <span className="text-sm font-medium text-slate-700 dark:text-slate-300">WhatsApp</span>
              </a>
              <a href="#" className="flex items-center gap-3 p-3 rounded-xl hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors group">
                <span className="material-symbols-outlined text-slate-400 group-hover:text-blue-500 transition-colors">mail</span>
                <span className="text-sm font-medium text-slate-700 dark:text-slate-300">Email Support</span>
              </a>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      <motion.button 
        layout
        initial={{ scale: 0 }}
        animate={{ scale: 1 }}
        transition={{ delay: 1 }}
        onClick={() => setIsExpanded(!isExpanded)}
        className="bg-primary hover:bg-primary-dark text-white p-4 rounded-full shadow-2xl shadow-primary/40 transition-transform hover:scale-105 flex items-center justify-center gap-2 font-bold text-base group"
        aria-label={isExpanded ? "Close Chat" : "Chat with an Expert"}
      >
        <AnimatePresence mode='wait'>
          {isExpanded ? (
            <motion.div
              key="close"
              initial={{ opacity: 0, scale: 0.5 }}
              animate={{ opacity: 1, scale: 1 }}
              exit={{ opacity: 0, scale: 0.5 }}
              className="flex items-center gap-2"
            >
              <span className="material-symbols-outlined text-xl">close</span>
              <span>Close</span>
            </motion.div>
          ) : (
            <motion.div
              key="chat"
              initial={{ opacity: 0, scale: 0.5 }}
              animate={{ opacity: 1, scale: 1 }}
              exit={{ opacity: 0, scale: 0.5 }}
              className="flex items-center gap-2"
            >
              <span className="relative flex h-3 w-3">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                <span className="relative inline-flex rounded-full h-3 w-3 bg-green-400"></span>
              </span>
              <span className="material-symbols-outlined ml-1">chat</span>
            </motion.div>
          )}
        </AnimatePresence>
      </motion.button>
    </div>
  );
};

export default ChatButton;
