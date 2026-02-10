import { AnimatePresence, motion } from 'framer-motion';
import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import EnterpriseAuthLayout from '../components/auth/EnterpriseAuthLayout';
import Modal from '../components/common/Modal';
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
  const [activeModal, setActiveModal] = useState<'privacy' | 'terms' | 'support' | null>(null);
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
          throw new Error('ACCOUNT_LOCKED');
        }
        throw new Error(data.error || data.message || 'Invalid credentials');
      }

      const data = await response.json();

      if (data.mfa_wait) {
        navigate('/mfa-verify', { state: { userId: data.user.id } });
        return;
      }

      const userData = {
        ...data.user,
        permissions: data.permissions?.map((p: any) => 
            typeof p === 'string' ? p : `${p.module}:${p.action}`
        ) || []
      };

      login(data.token || data.access_token, data.refresh_token || '', userData);

      if (rememberMe) {
        localStorage.setItem('remembered_email', email);
      }
      navigate('/admin');
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
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

        <form onSubmit={handleLogin} className="space-y-4">
          <div className="space-y-1">
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
                className="w-full bg-slate-800/30 border border-white/5 rounded-2xl pl-12 pr-4 py-3 text-white focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary/50 transition-all placeholder:text-slate-600"
                placeholder="name@company.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
            </div>
          </div>

          <div className="space-y-1">
            <div className="ml-1">
              <label className="block text-xs font-bold text-slate-400 uppercase tracking-widest">
                Security Key
              </label>
            </div>
            <div className="relative group">
              <span className="absolute left-4 top-1/2 -translate-y-1/2 material-symbols-outlined text-slate-500 group-focus-within:text-primary transition-colors text-xl">
                lock
              </span>
              <input
                type={showPassword ? "text" : "password"}
                required
                className="w-full bg-slate-800/30 border border-white/5 rounded-2xl pl-12 pr-12 py-3 text-white focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary/50 transition-all placeholder:text-slate-600"
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
          
          <div className="flex justify-end -mt-0.5">
            <Link to="/forgot-password" title="sm" className="text-primary hover:text-primary-dark transition-colors text-[10px] font-bold uppercase tracking-wider">
              Forgot your password?
            </Link>
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
          
          <button
            type="submit"
            disabled={loading || (captchaRequired && !captchaSolved)}
            className="w-full bg-gradient-to-r from-primary to-blue-600 hover:from-primary-dark hover:to-blue-700 text-white font-bold py-3 rounded-2xl transition-all shadow-xl shadow-primary/20 disabled:opacity-50 flex items-center justify-center space-x-2 text-sm uppercase tracking-widest mt-2"
          >
            {loading ? (
              <>
                <svg className="animate-spin h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                <span>Signing In...</span>
              </>
            ) : (
              <span>Sign In</span>
            )}
          </button>

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
        </form>

        <div className="mt-6">
          <div className="relative mb-5">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-slate-800"></div>
            </div>
            <div className="relative flex justify-center text-[10px] uppercase font-bold tracking-[0.2em]">
              <span className="bg-[#111827] px-4 text-slate-600">Enterprise SSO</span>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <button className="flex items-center justify-center gap-2 bg-slate-800/30 border border-white/5 hover:bg-slate-800/50 py-2.5 rounded-2xl transition-all group">
              <svg className="w-4 h-4" viewBox="0 0 24 24">
                <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"></path>
                <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"></path>
                <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05"></path>
                <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"></path>
              </svg>
              <span className="text-xs font-bold text-slate-300 group-hover:text-white transition-colors uppercase tracking-tight">Google</span>
            </button>
            <button className="flex items-center justify-center gap-2 bg-slate-800/30 border border-white/5 hover:bg-slate-800/50 py-2.5 rounded-2xl transition-all group">
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

        <div className="mt-6 pt-4 border-t border-slate-800/50 text-center">
          <div className="flex justify-center gap-4 text-xs text-slate-500">
            <button 
              onClick={() => setActiveModal('privacy')}
              className="hover:text-primary transition-colors"
            >
              Privacy
            </button>
            <span className="text-slate-800">•</span>
            <button 
              onClick={() => setActiveModal('terms')}
              className="hover:text-primary transition-colors"
            >
              Terms
            </button>
            <span className="text-slate-800">•</span>
            <button 
              onClick={() => setActiveModal('support')}
              className="hover:text-primary transition-colors"
            >
              Support
            </button>
          </div>
          <p className="text-slate-500 text-xs font-medium mt-4">
            New fleet officer?{' '}
            <Link to="/signup" className="text-primary hover:text-primary-dark font-bold uppercase tracking-wider ml-1">Create Account</Link>
          </p>
        </div>
      </EnterpriseAuthLayout>

      {/* Privacy Modal */}
      <Modal
        isOpen={activeModal === 'privacy'}
        onClose={() => setActiveModal(null)}
        title="Privacy Policy"
        size="lg"
      >
        <div className="prose prose-invert max-w-none space-y-6">
          <div className="bg-slate-800/20 border border-white/5 rounded-2xl p-5">
            <h4 className="text-white mb-3 flex items-center gap-2">
              <span className="material-symbols-outlined text-primary">policy</span>
              Last Updated: January 15, 2026
            </h4>
            <p className="text-slate-400 text-sm">
              This Privacy Policy describes how FlyFleet Global Operations collects, uses, and protects your personal information.
            </p>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">Information We Collect</h4>
            
            <div className="space-y-4">
              <div>
                <h5 className="text-slate-200 font-semibold mb-2">Account Information</h5>
                <ul className="text-slate-400 space-y-1 text-sm ml-4">
                  <li className="list-disc">Full name and job title</li>
                  <li className="list-disc">Corporate email address</li>
                  <li className="list-disc">Phone number and contact details</li>
                  <li className="list-disc">Company name and organization details</li>
                  <li className="list-disc">Role and department information</li>
                </ul>
              </div>
              
              <div>
                <h5 className="text-slate-200 font-semibold mb-2">Usage Data</h5>
                <ul className="text-slate-400 space-y-1 text-sm ml-4">
                  <li className="list-disc">Login times and session duration</li>
                  <li className="list-disc">Pages visited and features used</li>
                  <li className="list-disc">API calls and data requests</li>
                  <li className="list-disc">Device information and IP addresses</li>
                  <li className="list-disc">Performance metrics and error logs</li>
                </ul>
              </div>
              
              <div>
                <h5 className="text-slate-200 font-semibold mb-2">Fleet Data</h5>
                <ul className="text-slate-400 space-y-1 text-sm ml-4">
                  <li className="list-disc">Vehicle identification numbers (VINs)</li>
                  <li className="list-disc">GPS location and tracking data</li>
                  <li className="list-disc">Maintenance records and service history</li>
                  <li className="list-disc">Fuel consumption and efficiency metrics</li>
                  <li className="list-disc">Driver behavior and safety data</li>
                </ul>
              </div>
            </div>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">How We Use Your Information</h4>
            <div className="space-y-3 text-slate-400">
              <p><span className="font-semibold text-slate-300">Service Provision:</span> To operate and maintain our fleet management platform, including vehicle tracking, maintenance scheduling, and reporting.</p>
              <p><span className="font-semibold text-slate-300">Communication:</span> To send you important updates, security alerts, product announcements, and customer support responses.</p>
              <p><span className="font-semibold text-slate-300">Analytics:</span> To understand usage patterns, improve our services, and develop new features based on customer needs.</p>
              <p><span className="font-semibold text-slate-300">Compliance:</span> To meet legal requirements, industry regulations, and audit obligations.</p>
              <p><span className="font-semibold text-slate-300">Security:</span> To protect against unauthorized access, fraud, and other security threats.</p>
            </div>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">Data Protection & Security</h4>
            <div className="space-y-3 text-slate-400">
              <p>We implement robust security measures to protect your data:</p>
              <ul className="space-y-2 ml-4">
                <li className="flex items-start gap-2">
                  <span className="material-symbols-outlined text-green-500 mt-0.5 text-sm">lock</span>
                  <span><span className="font-semibold">Encryption:</span> All data is encrypted using AES-256 encryption both in transit and at rest.</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="material-symbols-outlined text-green-500 mt-0.5 text-sm">security</span>
                  <span><span className="font-semibold">Access Controls:</span> Role-based access controls and multi-factor authentication for all accounts.</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="material-symbols-outlined text-green-500 mt-0.5 text-sm">cloud_done</span>
                  <span><span className="font-semibold">Infrastructure:</span> Secure cloud infrastructure with regular security audits and penetration testing.</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="material-symbols-outlined text-green-500 mt-0.5 text-sm">backup</span>
                  <span><span className="font-semibold">Backups:</span> Automated daily backups with 90-day retention and disaster recovery capabilities.</span>
                </li>
              </ul>
            </div>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">Your Rights & Choices</h4>
            <div className="space-y-3 text-slate-400">
              <p>You have the following rights regarding your personal information:</p>
              <ul className="space-y-2 ml-4">
                <li className="flex items-start gap-2">
                  <span className="material-symbols-outlined text-blue-500 mt-0.5 text-sm">visibility</span>
                  <span><span className="font-semibold">Access:</span> Request a copy of the personal information we hold about you.</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="material-symbols-outlined text-blue-500 mt-0.5 text-sm">edit</span>
                  <span><span className="font-semibold">Correction:</span> Request correction of inaccurate or incomplete personal information.</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="material-symbols-outlined text-blue-500 mt-0.5 text-sm">delete</span>
                  <span><span className="font-semibold">Deletion:</span> Request deletion of your personal information (subject to legal requirements).</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="material-symbols-outlined text-blue-500 mt-0.5 text-sm">block</span>
                  <span><span className="font-semibold">Objection:</span> Object to processing of your personal information for marketing purposes.</span>
                </li>
              </ul>
              <p className="mt-3">To exercise these rights, contact our Data Protection Officer at <a href="mailto:dpo@flyfleet.com" className="text-primary hover:underline">dpo@flyfleet.com</a></p>
            </div>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">Data Retention</h4>
            <div className="space-y-3 text-slate-400">
              <p>We retain your information for as long as necessary to provide our services and comply with legal obligations:</p>
              <ul className="space-y-1 ml-4 text-sm">
                <li className="list-disc">Account data: While your account is active, plus 7 years for legal compliance</li>
                <li className="list-disc">Usage logs: 2 years for analytics and troubleshooting</li>
                <li className="list-disc">Financial records: 7 years as required by tax laws</li>
                <li className="list-disc">Audit logs: 5 years for security and compliance purposes</li>
              </ul>
            </div>
          </div>

          <div className="bg-amber-500/10 border border-amber-500/30 rounded-2xl p-4">
            <h4 className="text-amber-500 mb-2 flex items-center gap-2">
              <span className="material-symbols-outlined">info</span>
              International Data Transfers
            </h4>
            <p className="text-amber-200 text-sm">
              Your data may be transferred to and processed in countries outside your residence. We ensure adequate protection through standard contractual clauses and compliance with GDPR requirements.
            </p>
          </div>
        </div>
      </Modal>

      {/* Terms Modal */}
      <Modal
        isOpen={activeModal === 'terms'}
        onClose={() => setActiveModal(null)}
        title="Terms of Service"
        size="lg"
      >
        <div className="prose prose-invert max-w-none space-y-6">
          <div className="bg-slate-800/20 border border-white/5 rounded-2xl p-5">
            <h4 className="text-white mb-3 flex items-center gap-2">
              <span className="material-symbols-outlined text-primary">description</span>
              Effective Date: January 1, 2026
            </h4>
            <p className="text-slate-400 text-sm">
              These Terms of Service govern your access to and use of FlyFleet's fleet management platform and services.
            </p>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">1. Acceptance of Terms</h4>
            <div className="space-y-3 text-slate-400">
              <p>By accessing, browsing, or using the FlyFleet platform ("Service"), you acknowledge that you have read, understood, and agree to be bound by these Terms of Service ("Terms"). If you are using the Service on behalf of an organization, you represent and warrant that you have the authority to bind that organization to these Terms.</p>
              <p className="bg-red-500/10 border border-red-500/30 rounded-lg p-3">
                <span className="font-semibold text-red-300">Important:</span> If you do not agree to these Terms, you must not access or use the Service.
              </p>
            </div>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">2. Service Description</h4>
            <div className="space-y-3 text-slate-400">
              <p>FlyFleet provides a comprehensive fleet management platform that includes:</p>
              <ul className="space-y-2 ml-4">
                <li className="flex items-start gap-2">
                  <span className="material-symbols-outlined text-blue-500 mt-0.5 text-sm">directions_car</span>
                  <span>Real-time vehicle tracking and GPS monitoring</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="material-symbols-outlined text-blue-500 mt-0.5 text-sm">build</span>
                  <span>Maintenance scheduling and service management</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="material-symbols-outlined text-blue-500 mt-0.5 text-sm">analytics</span>
                  <span>Fuel management and cost analytics</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="material-symbols-outlined text-blue-500 mt-0.5 text-sm">people</span>
                  <span>Driver management and performance tracking</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="material-symbols-outlined text-blue-500 mt-0.5 text-sm">report</span>
                  <span>Comprehensive reporting and compliance tools</span>
                </li>
              </ul>
            </div>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">3. Account Registration & Security</h4>
            <div className="space-y-4 text-slate-400">
              <div>
                <h5 className="text-slate-200 font-semibold mb-2">Registration Requirements</h5>
                <ul className="space-y-1 ml-4 text-sm">
                  <li className="list-disc">Provide accurate and complete registration information</li>
                  <li className="list-disc">Maintain and update your information regularly</li>
                  <li className="list-disc">Use a valid corporate email address</li>
                  <li className="list-disc">Be at least 18 years old or have parental consent</li>
                </ul>
              </div>
              
              <div>
                <h5 className="text-slate-200 font-semibold mb-2">Security Responsibilities</h5>
                <ul className="space-y-1 ml-4 text-sm">
                  <li className="list-disc">Maintain the confidentiality of your account credentials</li>
                  <li className="list-disc">Enable two-factor authentication when available</li>
                  <li className="list-disc">Notify us immediately of any unauthorized access</li>
                  <li className="list-disc">Log out of your account after each session</li>
                </ul>
              </div>
              
              <div className="bg-amber-500/10 border border-amber-500/30 rounded-xl p-3">
                <p className="text-amber-200 text-sm">
                  <span className="font-semibold">Note:</span> You are solely responsible for all activities that occur under your account, whether authorized or not.
                </p>
              </div>
            </div>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">4. Acceptable Use Policy</h4>
            <div className="space-y-3 text-slate-400">
              <p>You agree to use the Service only for lawful purposes and in accordance with these Terms. You shall not:</p>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-3">
                <div className="bg-red-500/10 border border-red-500/30 rounded-xl p-4">
                  <h5 className="text-red-300 font-semibold mb-2">Prohibited Activities</h5>
                  <ul className="space-y-1 text-sm">
                    <li className="flex items-start gap-1">
                      <span>•</span>
                      <span>Violate any applicable laws or regulations</span>
                    </li>
                    <li className="flex items-start gap-1">
                      <span>•</span>
                      <span>Infringe intellectual property rights</span>
                    </li>
                    <li className="flex items-start gap-1">
                      <span>•</span>
                      <span>Reverse engineer or hack the Service</span>
                    </li>
                    <li className="flex items-start gap-1">
                      <span>•</span>
                      <span>Use the Service for illegal activities</span>
                    </li>
                  </ul>
                </div>
                
                <div className="bg-green-500/10 border border-green-500/30 rounded-xl p-4">
                  <h5 className="text-green-300 font-semibold mb-2">Permitted Uses</h5>
                  <ul className="space-y-1 text-sm">
                    <li className="flex items-start gap-1">
                      <span>•</span>
                      <span>Manage your organization's fleet</span>
                    </li>
                    <li className="flex items-start gap-1">
                      <span>•</span>
                      <span>Generate legitimate business reports</span>
                    </li>
                    <li className="flex items-start gap-1">
                      <span>•</span>
                      <span>Track vehicles for operational purposes</span>
                    </li>
                    <li className="flex items-start gap-1">
                      <span>•</span>
                      <span>Monitor driver performance metrics</span>
                    </li>
                  </ul>
                </div>
              </div>
            </div>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">5. Data Ownership & Rights</h4>
            <div className="space-y-3 text-slate-400">
              <p><span className="font-semibold text-slate-300">Your Data:</span> You retain all rights, title, and interest in your data uploaded to the Service. We obtain no rights to your data except as necessary to provide the Service.</p>
              <p><span className="font-semibold text-slate-300">Our Rights:</span> FlyFleet retains all rights, title, and interest in and to the Service, including all software, technology, and intellectual property.</p>
              <p><span className="font-semibold text-slate-300">Feedback:</span> If you provide feedback or suggestions about the Service, we may use that feedback without obligation to you.</p>
            </div>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">6. Subscription Plans & Payment</h4>
            <div className="space-y-3 text-slate-400">
              <div className="bg-slate-800/30 border border-white/10 rounded-xl p-4">
                <h5 className="text-slate-200 font-semibold mb-3">Plan Details</h5>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-sm">
                  <div>
                    <h6 className="font-medium text-white mb-1">Starter Plan</h6>
                    <p className="text-slate-400">Up to 10 vehicles • $49/month</p>
                  </div>
                  <div>
                    <h6 className="font-medium text-white mb-1">Professional Plan</h6>
                    <p className="text-slate-400">Up to 100 vehicles • $199/month</p>
                  </div>
                  <div>
                    <h6 className="font-medium text-white mb-1">Enterprise Plan</h6>
                    <p className="text-slate-400">Unlimited vehicles • Custom pricing</p>
                  </div>
                </div>
              </div>
              
              <div className="space-y-2">
                <p><span className="font-semibold text-slate-300">Payment Terms:</span> Fees are billed in advance on a monthly or annual basis. All fees are non-refundable.</p>
                <p><span className="font-semibold text-slate-300">Auto-Renewal:</span> Subscriptions automatically renew unless canceled 30 days prior to renewal.</p>
                <p><span className="font-semibold text-slate-300">Taxes:</span> You are responsible for all applicable taxes, duties, and government charges.</p>
              </div>
            </div>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">7. Termination</h4>
            <div className="space-y-3 text-slate-400">
              <div>
                <h5 className="text-slate-200 font-semibold mb-2">By You</h5>
                <p>You may terminate your account at any time by contacting support. Upon termination, you will lose access to the Service and your data may be deleted according to our retention policy.</p>
              </div>
              
              <div>
                <h5 className="text-slate-200 font-semibold mb-2">By Us</h5>
                <p>We may suspend or terminate your account immediately without notice if you:</p>
                <ul className="space-y-1 ml-4 mt-2 text-sm">
                  <li className="list-disc">Violate these Terms</li>
                  <li className="list-disc">Fail to pay fees</li>
                  <li className="list-disc">Engage in fraudulent activity</li>
                  <li className="list-disc">Pose a security risk</li>
                </ul>
              </div>
            </div>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">8. Disclaimer of Warranties</h4>
            <div className="space-y-3 text-slate-400">
              <p className="bg-red-500/10 border border-red-500/30 rounded-xl p-4">
                THE SERVICE IS PROVIDED "AS IS" AND "AS AVAILABLE" WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED. FLYFLEET DISCLAIMS ALL WARRANTIES, INCLUDING MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, AND NON-INFRINGEMENT.
              </p>
              <p>We do not guarantee that the Service will be uninterrupted, secure, or error-free. Your use of the Service is at your own risk.</p>
            </div>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">9. Limitation of Liability</h4>
            <div className="space-y-3 text-slate-400">
              <p className="bg-red-500/10 border border-red-500/30 rounded-xl p-4">
                TO THE MAXIMUM EXTENT PERMITTED BY LAW, FLYFLEET SHALL NOT BE LIABLE FOR ANY INDIRECT, INCIDENTAL, SPECIAL, CONSEQUENTIAL, OR PUNITIVE DAMAGES ARISING OUT OF OR RELATED TO THESE TERMS OR THE SERVICE.
              </p>
              <p>Our total liability for any claims arising from these Terms shall not exceed the amount you paid us in the twelve months preceding the claim.</p>
            </div>
          </div>

          <div>
            <h4 className="text-white mb-3 text-lg font-bold border-b border-slate-700 pb-2">10. Governing Law & Dispute Resolution</h4>
            <div className="space-y-3 text-slate-400">
              <p>These Terms shall be governed by and construed in accordance with the laws of Delaware, United States, without regard to its conflict of law provisions.</p>
              <p>Any disputes arising from these Terms shall be resolved through binding arbitration in accordance with the rules of the American Arbitration Association.</p>
            </div>
          </div>

          <div className="bg-blue-500/10 border border-blue-500/30 rounded-2xl p-4">
            <h4 className="text-blue-300 mb-2 flex items-center gap-2">
              <span className="material-symbols-outlined">contact_support</span>
              Questions About These Terms?
            </h4>
            <p className="text-blue-200 text-sm">
              Contact our Legal Team at <a href="mailto:legal@flyfleet.com" className="text-blue-400 hover:underline font-medium">legal@flyfleet.com</a> or call +1 (555) 123-4567.
            </p>
          </div>
        </div>
      </Modal>

      {/* Support Modal */}
      <Modal
        isOpen={activeModal === 'support'}
        onClose={() => setActiveModal(null)}
        title="Support Center"
        size="lg"
      >
        <div className="space-y-6">
          <div className="bg-gradient-to-r from-primary/10 to-blue-600/10 border border-primary/20 rounded-2xl p-5">
            <div className="text-center mb-4">
              <span className="material-symbols-outlined text-primary text-3xl mb-2">support_agent</span>
              <h3 className="text-white text-lg font-bold">How Can We Help You?</h3>
              <p className="text-slate-400 text-sm mt-1">Our support team is ready to assist you 24/7</p>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Email Support */}
            <div className="bg-slate-800/30 border border-white/5 rounded-2xl p-5 hover:border-primary/30 transition-all duration-300">
              <div className="flex items-start gap-4">
                <div className="bg-primary/10 p-3 rounded-xl flex-shrink-0">
                  <span className="material-symbols-outlined text-primary text-xl">mail</span>
                </div>
                <div className="flex-1">
                  <h4 className="text-white font-bold mb-2">Email Support</h4>
                  <p className="text-slate-400 text-sm mb-4">Submit a detailed support ticket and get a response within 2 hours</p>
                  <div className="space-y-2">
                    <div className="flex items-center gap-2 text-xs text-slate-500">
                      <span className="material-symbols-outlined text-green-500 text-sm">check_circle</span>
                      <span>Response within 2 hours (business hours)</span>
                    </div>
                    <div className="flex items-center gap-2 text-xs text-slate-500">
                      <span className="material-symbols-outlined text-green-500 text-sm">check_circle</span>
                      <span>24/7 emergency support available</span>
                    </div>
                    <div className="flex items-center gap-2 text-xs text-slate-500">
                      <span className="material-symbols-outlined text-green-500 text-sm">check_circle</span>
                      <span>Priority escalation for Enterprise customers</span>
                    </div>
                  </div>
                  <a 
                    href="mailto:support@flyfleet.com" 
                    className="mt-4 inline-flex items-center gap-2 bg-primary hover:bg-primary-dark text-white px-4 py-2 rounded-lg text-sm font-medium transition-all shadow-lg shadow-primary/20"
                  >
                    <span className="material-symbols-outlined text-base">send</span>
                    Send Email
                  </a>
                </div>
              </div>
            </div>
            
            {/* Live Chat */}
            <div className="bg-slate-800/30 border border-white/5 rounded-2xl p-5 hover:border-green-500/30 transition-all duration-300">
              <div className="flex items-start gap-4">
                <div className="bg-green-500/10 p-3 rounded-xl flex-shrink-0">
                  <span className="material-symbols-outlined text-green-500 text-xl">chat</span>
                </div>
                <div className="flex-1">
                  <h4 className="text-white font-bold mb-2">Live Chat</h4>
                  <p className="text-slate-400 text-sm mb-4">Get instant help from our support specialists</p>
                  <div className="space-y-2">
                    <div className="flex items-center gap-2 text-xs text-slate-500">
                      <span className="material-symbols-outlined text-green-500 text-sm">schedule</span>
                      <span>24/7 Availability</span>
                    </div>
                    <div className="flex items-center gap-2 text-xs text-slate-500">
                      <span className="material-symbols-outlined text-green-500 text-sm">bolt</span>
                      <span>Average response time: &lt; 30 seconds</span>
                    </div>
                    <div className="flex items-center gap-2 text-xs text-slate-500">
                      <span className="material-symbols-outlined text-green-500 text-sm">language</span>
                      <span>English, Spanish, French support</span>
                    </div>
                  </div>
                  <button className="mt-4 inline-flex items-center gap-2 bg-green-500 hover:bg-green-600 text-white px-4 py-2 rounded-lg text-sm font-medium transition-all shadow-lg shadow-green-500/20">
                    <span className="material-symbols-outlined text-base">chat_bubble</span>
                    Start Chat
                  </button>
                </div>
              </div>
            </div>
            
            {/* Phone Support */}
            <div className="bg-slate-800/30 border border-white/5 rounded-2xl p-5 hover:border-purple-500/30 transition-all duration-300">
              <div className="flex items-start gap-4">
                <div className="bg-purple-500/10 p-3 rounded-xl flex-shrink-0">
                  <span className="material-symbols-outlined text-purple-500 text-xl">phone</span>
                </div>
                <div className="flex-1">
                  <h4 className="text-white font-bold mb-2">Phone Support</h4>
                  <p className="text-slate-400 text-sm mb-4">Speak directly with our support engineers</p>
                  <div className="space-y-2">
                    <div className="text-slate-400 text-sm">
                      <div className="font-medium text-white">North America:</div>
                      <div>+1 (800) 555-0123</div>
                    </div>
                    <div className="text-slate-400 text-sm">
                      <div className="font-medium text-white">Europe:</div>
                      <div>+44 20 7946 0123</div>
                    </div>
                    <div className="text-slate-400 text-sm">
                      <div className="font-medium text-white">Asia-Pacific:</div>
                      <div>+61 2 9876 5432</div>
                    </div>
                  </div>
                  <div className="mt-3 text-xs text-slate-500">
                    Business Hours: Mon-Fri 8AM-8PM EST
                  </div>
                </div>
              </div>
            </div>
            
            {/* Documentation */}
            <div className="bg-slate-800/30 border border-white/5 rounded-2xl p-5 hover:border-blue-500/30 transition-all duration-300">
              <div className="flex items-start gap-4">
                <div className="bg-blue-500/10 p-3 rounded-xl flex-shrink-0">
                  <span className="material-symbols-outlined text-blue-500 text-xl">book</span>
                </div>
                <div className="flex-1">
                  <h4 className="text-white font-bold mb-2">Documentation</h4>
                  <p className="text-slate-400 text-sm mb-4">Comprehensive guides and tutorials</p>
                  <div className="space-y-2">
                    <div className="flex items-center gap-2 text-xs text-slate-500">
                      <span className="material-symbols-outlined text-blue-500 text-sm">article</span>
                      <span>User Guides & Tutorials</span>
                    </div>
                    <div className="flex items-center gap-2 text-xs text-slate-500">
                      <span className="material-symbols-outlined text-blue-500 text-sm">api</span>
                      <span>API Documentation</span>
                    </div>
                    <div className="flex items-center gap-2 text-xs text-slate-500">
                      <span className="material-symbols-outlined text-blue-500 text-sm">videocam</span>
                      <span>Video Training Library</span>
                    </div>
                    <div className="flex items-center gap-2 text-xs text-slate-500">
                      <span className="material-symbols-outlined text-blue-500 text-sm">quiz</span>
                      <span>FAQ & Troubleshooting</span>
                    </div>
                  </div>
                  <button className="mtCorporate-4 inline-flex items-center gap-2 bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-lg text-sm font-medium transition-all shadow-lg shadow-blue-500/20">
                    <span className="material-symbols-outlined text-base">menu_book</span>
                    Browse Docs
                  </button>
                </div>
              </div>
            </div>
          </div>

          <div className="bg-slate-800/20 border border-white/5 rounded-2xl p-5">
            <h4 className="text-white font-bold mb-3 flex items-center gap-2">
              <span className="material-symbols-outlined text-yellow-500">star</span>
              Premium Support Options
            </h4>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="text-center p-3 bg-slate-800/30 rounded-xl">
                <div className="text-2xl font-bold text-white mb-1">24/7</div>
                <div className="text-slate-400 text-sm">Round-the-clock support</div>
              </div>
              <div className="text-center p-3 bg-slate-800/30 rounded-xl">
                <div className="text-2xl font-bold text-white mb-1">&lt; 15 min</div>
                <div className="text-slate-400 text-sm">Average response time</div>
              </div>
              <div className="text-center p-3 bg-slate-800/30 rounded-xl">
                <div className="text-2xl font-bold text-white mb-1">99.9%</div>
                <div className="text-slate-400 text-sm">Customer satisfaction</div>
              </div>
            </div>
          </div>

          <div className="bg-amber-500/10 border border-amber-500/30 rounded-2xl p-4">
            <h4 className="text-amber-300 mb-2 flex items-center gap-2">
              <span className="material-symbols-outlined">emergency</span>
              Emergency Support
            </h4>
            <p className="text-amber-200 text-sm">
              For critical system outages or security incidents, contact our emergency line:
              <br />
              <span className="font-bold text-white">+1 (800) EMERGENCY</span> or email <a href="mailto:emergency@flyfleet.com" className="text-amber-400 hover:underline">emergency@flyfleet.com</a>
            </p>
          </div>
        </div>
      </Modal>
    </>
  );
};

export default Login;