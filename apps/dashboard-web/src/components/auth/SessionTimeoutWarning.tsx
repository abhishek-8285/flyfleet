import { AnimatePresence, motion } from 'framer-motion';
import React, { useEffect, useState } from 'react';
import api from '../../services/api';
import AuthButton from './AuthButton';

interface SessionTimeoutWarningProps {
  onLogout: () => void;
  expiryMinutes?: number;
}

const SessionTimeoutWarning: React.FC<SessionTimeoutWarningProps> = ({
  onLogout,
  expiryMinutes = 2,
}) => {
  const [timeLeft, setTimeLeft] = useState(expiryMinutes * 60);
  const [isVisible, setIsVisible] = useState(false);
  const [extending, setExtending] = useState(false);

  useEffect(() => {
    // In a real app, this would be triggered by a global store or token observer
    // Here we'll simulate the 2-minute warning appearing
    // const showTimer = setTimeout(() => setIsVisible(true), 1000);
    // return () => clearTimeout(showTimer);
  }, []);

  useEffect(() => {
    if (!isVisible || timeLeft <= 0) {
      if (timeLeft <= 0) onLogout();
      return;
    }

    const timer = setInterval(() => {
      setTimeLeft((prev) => prev - 1);
    }, 1000);

    return () => clearInterval(timer);
  }, [isVisible, timeLeft, onLogout]);

  const handleExtend = async () => {
    setExtending(true);
    try {
      await api.refreshToken();
      setIsVisible(false);
      setTimeLeft(expiryMinutes * 60);
    } catch (err) {
      onLogout();
    } finally {
      setExtending(false);
    }
  };

  const formatTime = (s: number) => {
    const mins = Math.floor(s / 60);
    const secs = s % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  return (
    <AnimatePresence>
      {isVisible && (
        <div className="fixed inset-0 z-[9999] flex items-center justify-center p-4 sm:p-6">
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="absolute inset-0 bg-slate-950/80 backdrop-blur-md"
          />

          <motion.div
            initial={{ opacity: 0, scale: 0.9, y: 20 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.95, y: 10 }}
            className="relative w-full max-w-md bg-slate-900 border border-white/10 rounded-[32px] overflow-hidden shadow-2xl"
          >
            <div className="absolute top-0 left-0 w-full h-1 bg-slate-800">
              <motion.div
                initial={{ width: '100%' }}
                animate={{ width: '0%' }}
                transition={{ duration: expiryMinutes * 60, ease: 'linear' }}
                className="h-full bg-primary shadow-[0_0_10px_rgba(59,130,246,0.5)]"
              />
            </div>

            <div className="p-8 pt-10 text-center space-y-6">
              <div className="w-16 h-16 bg-amber-500/10 rounded-2xl flex items-center justify-center mx-auto border border-amber-500/20">
                <span className="material-symbols-outlined text-amber-500 text-3xl">timer</span>
              </div>

              <div className="space-y-2">
                <h3 className="text-white text-xl font-bold tracking-tight">
                  Session Expiring Soon
                </h3>
                <p className="text-slate-400 text-sm leading-relaxed">
                  For your security, your platform terminal session will automatically terminate in:
                </p>
                <div className="text-3xl font-mono font-black text-white tracking-widest pt-2">
                  {formatTime(timeLeft)}
                </div>
              </div>

              <div className="flex flex-col gap-3 pt-2">
                <AuthButton onClick={handleExtend} loading={extending}>
                  Extend Session
                </AuthButton>
                <button
                  onClick={onLogout}
                  className="w-full py-4 text-xs font-bold text-slate-500 hover:text-white transition-colors uppercase tracking-widest"
                >
                  Logout Now
                </button>
              </div>
            </div>

            <div className="bg-slate-800/20 p-4 border-t border-white/5 flex items-center justify-center gap-2">
              <span className="material-symbols-outlined text-slate-500 text-sm">
                verified_user
              </span>
            </div>
          </motion.div>
        </div>
      )}
    </AnimatePresence>
  );
};

export default SessionTimeoutWarning;
