import React, { useState } from 'react';

interface LoginScreenProps {
  onLoginSuccess: (email: string) => void;
}

const LoginScreen: React.FC<LoginScreenProps> = ({ onLoginSuccess }) => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [hasError, setHasError] = useState(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setHasError(false);

    // Simulate network request
    setTimeout(() => {
      // Demo logic: If email contains "error", show error state.
      // Otherwise proceed.
      if (email.toLowerCase().includes('error') || password === 'wrong') {
        setHasError(true);
        setIsLoading(false);
      } else {
        onLoginSuccess(email);
      }
    }, 1500);
  };

  return (
    <div className="flex-1 flex flex-col items-center justify-center p-4">
      <div className="w-full max-w-[480px] flex flex-col items-center">
        {/* Logo Section */}
        <div className="mb-8 flex justify-center">
          <div className="h-16 w-16 bg-primary/10 rounded-xl flex items-center justify-center">
            <span className="material-symbols-outlined text-primary text-4xl">local_shipping</span>
          </div>
        </div>

        {/* Card */}
        <div className="w-full bg-white dark:bg-[#151f2b] rounded-xl shadow-[0_8px_30px_rgb(0,0,0,0.04)] border border-slate-200 dark:border-slate-800 overflow-hidden">
          <div className="p-8 sm:p-10">
            <div className="text-center mb-8">
              <h2 className="text-slate-900 dark:text-white tracking-tight text-2xl font-bold leading-tight mb-2">
                Sign in to FleetCommand
              </h2>
              <p className="text-slate-500 dark:text-slate-400 text-sm font-normal leading-normal">
                Enter your credentials to access the fleet dashboard.
              </p>
            </div>

            <form onSubmit={handleSubmit} className="space-y-6">
              {/* Email Input */}
              <div className="flex flex-col gap-1">
                <label className="text-slate-900 dark:text-slate-200 text-sm font-medium leading-normal" htmlFor="email">
                  Email Address
                </label>
                <div className="relative">
                  <input
                    id="email"
                    type="email"
                    required
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="name@company.com"
                    className={`form-input flex w-full min-w-0 flex-1 resize-none overflow-hidden rounded-lg focus:outline-0 h-12 px-4 text-base font-normal leading-normal transition-all duration-200 ease-in-out
                      ${hasError 
                        ? 'text-slate-900 dark:text-white border-error dark:border-error bg-slate-50 dark:bg-gray-800 focus:ring-2 focus:ring-error/20 focus:border-error' 
                        : 'text-slate-900 dark:text-white border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-gray-800 focus:ring-2 focus:ring-primary/20 focus:border-primary placeholder:text-slate-400'
                      }`}
                  />
                  {hasError && (
                    <div className="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none">
                      <span className="material-symbols-outlined text-error" style={{ fontSize: '20px' }}>error</span>
                    </div>
                  )}
                </div>
                {hasError && (
                  <p className="text-error text-sm font-medium mt-0.5 animate-pulse">Invalid email format</p>
                )}
              </div>

              {/* Password Input */}
              <div className="flex flex-col gap-1">
                <div className="flex items-center justify-between">
                  <label className="text-slate-900 dark:text-slate-200 text-sm font-medium leading-normal" htmlFor="password">
                    Password
                  </label>
                  <button type="button" className="text-sm font-medium text-primary hover:text-blue-500 transition-colors">
                    Forgot password?
                  </button>
                </div>
                <div className="relative flex w-full items-stretch rounded-lg">
                  <input
                    id="password"
                    type={showPassword ? "text" : "password"}
                    required
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="Enter your password"
                    className={`form-input flex w-full min-w-0 flex-1 resize-none overflow-hidden rounded-l-lg focus:outline-0 h-12 px-4 border-r-0 text-base font-normal leading-normal transition-all duration-200 ease-in-out
                       ${hasError 
                        ? 'text-slate-900 dark:text-white border-error dark:border-error bg-slate-50 dark:bg-gray-800 focus:ring-2 focus:ring-error/20 focus:border-error' 
                        : 'text-slate-900 dark:text-white border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-gray-800 focus:ring-2 focus:ring-primary/20 focus:border-primary placeholder:text-slate-400'
                      }`}
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className={`flex items-center justify-center px-4 rounded-r-lg border border-l-0 transition-colors group
                      ${hasError
                        ? 'border-error dark:border-error bg-slate-50 dark:bg-gray-800'
                        : 'border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-gray-800 hover:bg-slate-100 dark:hover:bg-slate-700'
                      }`}
                  >
                    <span className="material-symbols-outlined text-slate-400 group-hover:text-slate-600 dark:group-hover:text-slate-300" style={{ fontSize: '20px' }}>
                      {showPassword ? 'visibility_off' : 'visibility'}
                    </span>
                  </button>
                </div>
                {hasError && (
                  <p className="text-error text-sm font-medium mt-0.5">Incorrect password</p>
                )}
              </div>

              {/* Submit Button */}
              <div className="pt-2">
                <button
                  type="submit"
                  disabled={isLoading}
                  className={`flex w-full items-center justify-center overflow-hidden rounded-lg h-12 px-5 text-white text-base font-bold leading-normal tracking-[0.015em] transition-all duration-200 shadow-sm
                    ${isLoading ? 'bg-primary/80 cursor-not-allowed' : 'bg-primary hover:bg-blue-600 active:scale-[0.98] hover:shadow'}`}
                >
                  {isLoading ? (
                    <svg className="animate-spin -ml-1 mr-2 h-5 w-5 text-white" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" fill="currentColor"></path>
                    </svg>
                  ) : (
                    <span className="truncate">Sign In</span>
                  )}
                </button>
              </div>

              {/* SSO Options */}
              <div className="space-y-4">
                <div className="relative flex items-center justify-center">
                  <div className="flex-grow border-t border-slate-200 dark:border-slate-700"></div>
                  <span className="flex-shrink-0 mx-4 text-slate-400 text-xs font-medium uppercase tracking-wider">OR</span>
                  <div className="flex-grow border-t border-slate-200 dark:border-slate-700"></div>
                </div>
                
                <div className="flex flex-col gap-3">
                  <button type="button" className="flex w-full items-center justify-center gap-3 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-gray-800 px-4 h-12 text-sm font-medium text-slate-700 dark:text-slate-200 hover:bg-slate-50 dark:hover:bg-[#1e293b] focus:outline-none focus:ring-2 focus:ring-slate-200 transition-all shadow-sm">
                    <svg className="h-5 w-5" viewBox="0 0 24 24">
                      <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"></path>
                      <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"></path>
                      <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05"></path>
                      <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"></path>
                    </svg>
                    <span>Sign in with Google</span>
                  </button>
                  <button type="button" className="flex w-full items-center justify-center gap-3 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-gray-800 px-4 h-12 text-sm font-medium text-slate-700 dark:text-slate-200 hover:bg-slate-50 dark:hover:bg-[#1e293b] focus:outline-none focus:ring-2 focus:ring-slate-200 transition-all shadow-sm">
                    <svg className="h-5 w-5" viewBox="0 0 21 21">
                      <rect fill="#F25022" height="9" width="9" x="1" y="1"></rect>
                      <rect fill="#7FBA00" height="9" width="9" x="11" y="1"></rect>
                      <rect fill="#00A4EF" height="9" width="9" x="1" y="11"></rect>
                      <rect fill="#FFB900" height="9" width="9" x="11" y="11"></rect>
                    </svg>
                    <span>Sign in with Microsoft</span>
                  </button>
                  <button type="button" className="flex w-full items-center justify-center gap-3 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-gray-800 px-4 h-12 text-sm font-medium text-slate-700 dark:text-slate-200 hover:bg-slate-50 dark:hover:bg-[#1e293b] focus:outline-none focus:ring-2 focus:ring-slate-200 transition-all shadow-sm">
                    <span className="material-symbols-outlined text-slate-500" style={{ fontSize: '20px' }}>vpn_key</span>
                    <span>Sign in with SSO</span>
                  </button>
                </div>
              </div>
            </form>
          </div>
          
          {/* Card Footer */}
          <div className="bg-slate-50 dark:bg-slate-800/50 px-8 py-4 border-t border-slate-100 dark:border-slate-800 text-center">
            <p className="text-xs text-slate-500 dark:text-slate-400">
              Need help? <a href="#" className="text-primary hover:underline">Contact Support</a>
            </p>
          </div>
        </div>

        {/* Footer */}
        <div className="mt-8 text-center">
          <p className="text-slate-400 text-xs">
            © 2026 FleetCommand Enterprise. All rights reserved.
          </p>
          <div className="mt-2 flex gap-4 justify-center text-xs text-slate-400">
            <a href="#" className="hover:text-slate-600 dark:hover:text-slate-300 transition-colors">Privacy Policy</a>
            <span>·</span>
            <a href="#" className="hover:text-slate-600 dark:hover:text-slate-300 transition-colors">Terms of Service</a>
          </div>
        </div>
      </div>
      
      {/* Hint for demo purposes */}
      <div className="fixed bottom-4 right-4 bg-slate-800 text-slate-400 text-xs p-2 rounded opacity-50 hover:opacity-100 transition-opacity">
        Hint: Use email "error" to see validation. Any other to proceed.
      </div>
    </div>
  );
};

export default LoginScreen;