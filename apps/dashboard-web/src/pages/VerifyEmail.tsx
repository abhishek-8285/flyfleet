import { AnimatePresence, motion } from 'framer-motion';
import React, { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import EnterpriseAuthLayout from '../components/auth/EnterpriseAuthLayout';

const VerifyEmail: React.FC = () => {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const token = searchParams.get('token');

  useEffect(() => {
    const verify = async () => {
      if (!token) {
        setError('Missing verification token. Please check your email link.');
        setLoading(false);
        return;
      }

      try {
        const response = await fetch(`/api/v1/auth/verify-email?token=${token}`, {
          method: 'GET',
        });

        if (!response.ok) {
          const data = await response.json();
          throw new Error(data.message || 'Verification failed');
        }

        setSuccess(true);
        setTimeout(() => navigate('/login'), 3000);
      } catch (err: any) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    verify();
  }, [token, navigate]);

  return (
    <EnterpriseAuthLayout 
      title="Identity Verification" 
      subtitle="Finalizing your enterprise registration"
    >
      <div className="flex flex-col items-center justify-center min-h-[200px]">
        <AnimatePresence mode="wait">
          {loading ? (
            <motion.div 
              key="loading"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="flex flex-col items-center gap-4"
            >
              <div className="w-12 h-12 border-4 border-primary/20 border-t-primary rounded-full animate-spin" />
              <p className="text-slate-400 text-sm font-medium animate-pulse">Verifying credentials...</p>
            </motion.div>
          ) : error ? (
            <motion.div 
              key="error"
              initial={{ opacity: 0, scale: 0.9 }}
              animate={{ opacity: 1, scale: 1 }}
              className="text-center"
            >
              <div className="w-16 h-16 bg-red-500/10 rounded-full flex items-center justify-center mx-auto mb-4 border border-red-500/20">
                <span className="material-symbols-outlined text-red-500 text-3xl">error</span>
              </div>
              <h3 className="text-white font-bold mb-2">Verification Error</h3>
              <p className="text-slate-400 text-sm mb-6 max-w-[280px] mx-auto">{error}</p>
              <button 
                onClick={() => navigate('/login')}
                className="bg-slate-800 hover:bg-slate-700 text-white px-6 py-2 rounded-xl text-xs font-bold uppercase tracking-widest transition-colors"
              >
                Back to Login
              </button>
            </motion.div>
          ) : (
            <motion.div 
              key="success"
              initial={{ opacity: 0, scale: 0.9 }}
              animate={{ opacity: 1, scale: 1 }}
              className="text-center"
            >
              <div className="w-16 h-16 bg-green-500/10 rounded-full flex items-center justify-center mx-auto mb-4 border border-green-500/20">
                <span className="material-symbols-outlined text-green-500 text-3xl">verified_user</span>
              </div>
              <h3 className="text-white font-bold mb-2">Account Verified</h3>
              <p className="text-slate-400 text-sm mb-4">Your corporate identity has been confirmed.</p>
              <p className="text-primary text-[10px] font-bold uppercase tracking-widest animate-pulse">Redirecting to Access Portal...</p>
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </EnterpriseAuthLayout>
  );
};

export default VerifyEmail;
