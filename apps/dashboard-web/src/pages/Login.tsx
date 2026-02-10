import { AnimatePresence, motion } from 'framer-motion';
import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import EnterpriseAuthLayout from '../components/auth/EnterpriseAuthLayout';
import { useAuth } from '../context/AuthProvider';

const Login: React.FC = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [rememberMe, setRememberMe] = useState(false);
  const [captchaRequired, setCaptchaRequired] = useState(false);
  const [captchaSolved, setCaptchaSolved] = useState(false);
  const navigate = useNavigate();
  const { login } = useAuth();

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      const response = await fetch('/api/v1/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      });

      if (!response.ok) {
        const data = await response.json();
        if (data.message === 'CAPTCHA_REQUIRED') {
          setCaptchaRequired(true);
          throw new Error('Please solve the security verification to continue.');
        }
        if (data.message === 'account is temporarily locked due to multiple failed attempts') {
          // You could navigate to a specific /account-locked page or just show error
          throw new Error('ACCOUNT_LOCKED');
        }
        throw new Error(data.error || data.message || 'Invalid credentials');
      }

      const data = await response.json();

      if (data.mfa_wait) {
        navigate('/mfa-verify', { state: { userId: data.user.id } });
        return;
      }

      login(data.token || data.access_token, data.refresh_token || '', data.user);

      if (rememberMe) {
        localStorage.setItem('remembered_email', email);
      }
      navigate('/admin'); // Redirect to admin dashboard instead of just /dashboard
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <EnterpriseAuthLayout 
      title="Welcome Back" 
      subtitle="Sign in to your enterprise account"
    >
      {error && (
        <motion.div 
          initial={{ opacity: 0, height: 0 }}
          animate={{ opacity: 1, height: 'auto' }}
          className={`mb-6 p-3 rounded-xl text-xs font-medium flex items-center gap-2 ${
            error === 'ACCOUNT_LOCKED' 
              ? 'bg-orange-500/10 border border-orange-500/50 text-orange-500' 
              : 'bg-red-500/10 border border-red-500/50 text-red-500'
          }`}
        >
          <span className="material-symbols-outlined text-sm">
            {error === 'ACCOUNT_LOCKED' ? 'lock_person' : 'error'}
          </span>
          {error === 'ACCOUNT_LOCKED' 
            ? 'Account locked for 15 minutes. Use "Forgot Password" to recover.' 
            : error}
        </motion.div>
      )}

      <form onSubmit={handleLogin} className="space-y-5">
        <div className="space-y-1.5">
          <label className="block text-xs font-bold text-slate-400 uppercase tracking-widest ml-1">
            Corporate Email
          </label>
          <div className="relative group">
            <span className="absolute left-4 top-1/2 -translate-y-1/2 material-symbols-outlined text-slate-500 group-focus-within:text-primary transition-colors text-xl">
              mail
            </span>
            <input
              type="email"
              required
              className="w-full bg-slate-800/30 border border-white/5 rounded-2xl pl-12 pr-4 py-3.5 text-white focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary/50 transition-all placeholder:text-slate-600"
              placeholder="name@company.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
          </div>
        </div>

        <div className="space-y-1.5">
          <div className="flex justify-between items-center ml-1">
            <label className="block text-xs font-bold text-slate-400 uppercase tracking-widest">
              Security Key
            </label>
            <Link to="/forgot-password" title="sm" className="text-primary hover:text-primary-dark transition-colors text-[10px] font-bold uppercase tracking-wider">
              Forgot?
            </Link>
          </div>
          <div className="relative group">
            <span className="absolute left-4 top-1/2 -translate-y-1/2 material-symbols-outlined text-slate-500 group-focus-within:text-primary transition-colors text-xl">
              lock
            </span>
            <input
              type={showPassword ? "text" : "password"}
              required
              className="w-full bg-slate-800/30 border border-white/5 rounded-2xl pl-12 pr-12 py-3.5 text-white focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary/50 transition-all placeholder:text-slate-600"
              placeholder="••••••••"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
            <button
              type="button"
              onClick={() => setShowPassword(!showPassword)}
              className="absolute right-4 top-1/2 -translate-y-1/2 text-slate-500 hover:text-slate-300 transition-colors"
            >
              <span className="material-symbols-outlined text-xl">
                {showPassword ? 'visibility_off' : 'visibility'}
              </span>
            </button>
          </div>
        </div>

        <div className="flex items-center justify-between px-1">
          <label className="flex items-center gap-2 cursor-pointer group">
            <input 
              type="checkbox" 
              className="w-4 h-4 rounded border-white/10 bg-slate-800 text-primary focus:ring-primary/50 transition-all cursor-pointer"
              checked={rememberMe}
              onChange={(e) => setRememberMe(e.target.checked)}
            />
            <span className="text-xs text-slate-400 font-medium group-hover:text-slate-300 transition-colors">Remember this device</span>
          </label>
        </div>

        {/* Improved CAPTCHA Placeholder */}
        <AnimatePresence>
          {captchaRequired && (
            <motion.div 
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              onClick={() => setCaptchaSolved(true)}
              className={`bg-slate-800/20 border rounded-2xl p-4 flex items-center justify-between group hover:bg-slate-800/30 transition-all cursor-pointer ${
                captchaSolved ? 'border-green-500/50' : 'border-primary/30'
              }`}
            >
              <div className="flex items-center gap-3">
                <div className={`w-5 h-5 border-2 rounded flex items-center justify-center transition-colors ${
                  captchaSolved ? 'bg-green-500 border-green-500' : 'border-slate-600 group-hover:border-primary'
                }`}>
                  {captchaSolved && <span className="material-symbols-outlined text-white text-sm">check</span>}
                </div>
                <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">
                  {captchaSolved ? 'Verification Success' : 'Security verification required'}
                </span>
              </div>
              <span className={`material-symbols-outlined text-xl ${captchaSolved ? 'text-green-500' : 'text-slate-600'}`}>
                {captchaSolved ? 'verified' : 'verified_user'}
              </span>
            </motion.div>
          )}
        </AnimatePresence>

        <button
          type="submit"
          disabled={loading || (captchaRequired && !captchaSolved)}
          className="w-full bg-gradient-to-r from-primary to-blue-600 hover:from-primary-dark hover:to-blue-700 text-white font-bold py-4 rounded-2xl transition-all shadow-xl shadow-primary/20 disabled:opacity-50 flex items-center justify-center space-x-2 text-sm uppercase tracking-widest mt-2"
        >
          {loading ? (
            <>
              <svg className="animate-spin h-5 w-5 text-white" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <span>Authenticating...</span>
            </>
          ) : (
            <span>Sign In</span>
          )}
        </button>
      </form>

      <div className="mt-8">
        <div className="relative mb-6">
          <div className="absolute inset-0 flex items-center">
            <div className="w-full border-t border-slate-800"></div>
          </div>
          <div className="relative flex justify-center text-[10px] uppercase font-bold tracking-[0.2em]">
            <span className="bg-[#111827] px-4 text-slate-600">Enterprise SSO</span>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <button className="flex items-center justify-center gap-2 bg-slate-800/30 border border-white/5 hover:bg-slate-800/50 py-3 rounded-2xl transition-all group">
            <svg className="w-4 h-4" viewBox="0 0 24 24">
              <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"></path>
              <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"></path>
              <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05"></path>
              <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"></path>
            </svg>
            <span className="text-xs font-bold text-slate-300 group-hover:text-white transition-colors uppercase tracking-tight">Google</span>
          </button>
          <button className="flex items-center justify-center gap-2 bg-slate-800/30 border border-white/5 hover:bg-slate-800/50 py-3 rounded-2xl transition-all group">
            <svg className="w-4 h-4" viewBox="0 0 21 21">
              <rect fill="#F25022" height="9" width="9" x="1" y="1"></rect>
              <rect fill="#7FBA00" height="9" width="9" x="11" y="1"></rect>
              <rect fill="#00A4EF" height="9" width="9" x="1" y="11"></rect>
              <rect fill="#FFB900" height="9" width="9" x="11" y="11"></rect>
            </svg>
            <span className="text-xs font-bold text-slate-300 group-hover:text-white transition-colors uppercase tracking-tight">Azure AD</span>
          </button>
        </div>
      </div>

      <div className="mt-8 text-center">
        <p className="text-slate-500 text-xs font-medium">
          New fleet officer?{' '}
          <Link to="/signup" className="text-primary hover:text-primary-dark font-bold uppercase tracking-wider ml-1">Create Account</Link>
        </p>
      </div>
    </EnterpriseAuthLayout>
  );
};

export default Login;
