import { AnimatePresence, motion } from 'framer-motion';
import React, { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import EnterpriseAuthLayout from '../components/auth/EnterpriseAuthLayout';

const ResetPassword: React.FC = () => {
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);
  const [passwordStrength, setPasswordStrength] = useState(0);
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const token = searchParams.get('token');

  useEffect(() => {
    if (!token) {
      setError('Invalid or expired reset token. Please request a new link.');
    }
  }, [token]);

  useEffect(() => {
    let strength = 0;
    if (password.length >= 8) strength += 25;
    if (/[A-Z]/.test(password)) strength += 25;
    if (/[0-9]/.test(password)) strength += 25;
    if (/[^A-Za-z0-9]/.test(password)) strength += 25;
    setPasswordStrength(strength);
  }, [password]);

  const handleReset = async (e: React.FormEvent) => {
    e.preventDefault();
    if (password !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }
    if (passwordStrength < 75) {
      setError('Password does not meet enterprise security requirements');
      return;
    }
    setLoading(true);
    setError('');

    try {
      const response = await fetch('/api/v1/auth/reset-password', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token, newPassword: password }),
      });

      if (!response.ok) {
        throw new Error('Failed to reset password. The link may have expired.');
      }

      setSuccess(true);
      setTimeout(() => navigate('/login'), 3000);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <EnterpriseAuthLayout 
      title="Reset Password" 
      subtitle="Update your corporate credentials"
    >
      <AnimatePresence mode="wait">
        {error && (
          <motion.div 
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            className="mb-6 p-3 bg-red-500/10 border border-red-500/50 rounded-xl text-red-500 text-xs font-medium flex items-center gap-2"
          >
            <span className="material-symbols-outlined text-sm">error</span>
            {error}
          </motion.div>
        )}

        {success && (
          <motion.div 
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            className="mb-6 p-4 bg-green-500/10 border border-green-500/50 rounded-xl text-green-500 text-xs font-medium flex items-center gap-2"
          >
            <span className="material-symbols-outlined text-sm">check_circle</span>
            Password reset successful. Redirecting to login...
          </motion.div>
        )}
      </AnimatePresence>

      {!success && !error.includes('expired reset token') && (
        <form onSubmit={handleReset} className="space-y-4">
          <div className="space-y-1.5">
            <label className="block text-xs font-bold text-slate-400 uppercase tracking-widest ml-1">
              New Password
            </label>
            <div className="relative group">
              <span className="absolute left-4 top-1/2 -translate-y-1/2 material-symbols-outlined text-slate-500 group-focus-within:text-primary transition-colors text-xl">
                key
              </span>
              <input
                type="password"
                required
                className="w-full bg-slate-800/30 border border-white/5 rounded-2xl pl-12 pr-4 py-3 text-white focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all placeholder:text-slate-600"
                placeholder="••••••••"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </div>
            {/* Password Strength Indicator */}
            <div className="mt-2 px-1">
              <div className="flex justify-between items-center mb-1">
                <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Security Strength</span>
                <span className={`text-[10px] font-bold uppercase tracking-widest ${
                  passwordStrength <= 25 ? 'text-red-500' :
                  passwordStrength <= 50 ? 'text-yellow-500' :
                  passwordStrength <= 75 ? 'text-blue-500' : 'text-green-500'
                }`}>
                  {passwordStrength <= 25 ? 'Weak' :
                   passwordStrength <= 50 ? 'Fair' :
                   passwordStrength <= 75 ? 'Good' : 'Strong'}
                </span>
              </div>
              <div className="h-1 w-full bg-slate-800 rounded-full overflow-hidden">
                <motion.div 
                  initial={{ width: 0 }}
                  animate={{ width: `${passwordStrength}%` }}
                  className={`h-full transition-colors duration-500 ${
                    passwordStrength <= 25 ? 'bg-red-500' :
                    passwordStrength <= 50 ? 'bg-yellow-500' :
                    passwordStrength <= 75 ? 'bg-blue-500' : 'bg-green-500'
                  }`}
                />
              </div>
            </div>
          </div>

          <div className="space-y-1.5">
            <label className="block text-xs font-bold text-slate-400 uppercase tracking-widest ml-1">
              Confirm Password
            </label>
            <div className="relative group">
              <span className="absolute left-4 top-1/2 -translate-y-1/2 material-symbols-outlined text-slate-500 group-focus-within:text-primary transition-colors text-xl">
                password
              </span>
              <input
                type="password"
                required
                className="w-full bg-slate-800/30 border border-white/5 rounded-2xl pl-12 pr-4 py-3 text-white focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all placeholder:text-slate-600"
                placeholder="••••••••"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
              />
            </div>
          </div>

          <button
            type="submit"
            disabled={loading || !token}
            className="w-full bg-gradient-to-r from-primary to-blue-600 hover:from-primary-dark hover:to-blue-700 text-white font-bold py-4 rounded-2xl transition-all shadow-xl shadow-primary/20 disabled:opacity-50 flex items-center justify-center space-x-2 text-sm uppercase tracking-widest mt-4"
          >
            {loading ? 'Updating Credentials...' : 'Update Password'}
          </button>
        </form>
      )}

      {(error.includes('expired reset token') || success) && (
        <div className="mt-8 pt-4">
          <button 
            onClick={() => navigate('/login')}
            className="w-full bg-slate-800/50 hover:bg-slate-800 text-white font-bold py-4 rounded-2xl transition-all border border-white/5 text-sm uppercase tracking-widest"
          >
            Return to Login
          </button>
        </div>
      )}
    </EnterpriseAuthLayout>
  );
};

export default ResetPassword;
