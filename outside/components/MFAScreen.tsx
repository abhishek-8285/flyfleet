import React, { useState, useRef, useEffect } from 'react';

interface MFAScreenProps {
  onSuccess: () => void;
  onCancel: () => void;
}

const MFAScreen: React.FC<MFAScreenProps> = ({ onSuccess, onCancel }) => {
  const [code, setCode] = useState<string[]>(Array(6).fill(""));
  const [isVerifying, setIsVerifying] = useState(false);
  const [hasError, setHasError] = useState(false);
  const inputRefs = useRef<(HTMLInputElement | null)[]>([]);

  useEffect(() => {
    // Focus first input on mount
    if (inputRefs.current[0]) {
      inputRefs.current[0].focus();
    }
  }, []);

  const handleChange = (index: number, value: string) => {
    // Allow only numbers
    if (!/^\d*$/.test(value)) return;

    const newCode = [...code];
    newCode[index] = value.slice(-1); // Take last char if multiple
    setCode(newCode);
    setHasError(false); // Clear error on edit

    // Move focus to next input
    if (value && index < 5) {
      inputRefs.current[index + 1]?.focus();
    }
  };

  const handleKeyDown = (index: number, e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Backspace' && !code[index] && index > 0) {
      // Move focus back if backspace pressed on empty input
      inputRefs.current[index - 1]?.focus();
    }
  };

  const handlePaste = (e: React.ClipboardEvent) => {
    e.preventDefault();
    const pastedData = e.clipboardData.getData('text').slice(0, 6);
    if (!/^\d+$/.test(pastedData)) return;

    const newCode = [...code];
    pastedData.split('').forEach((char, i) => {
      if (i < 6) newCode[i] = char;
    });
    setCode(newCode);
    inputRefs.current[Math.min(pastedData.length, 5)]?.focus();
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const fullCode = code.join('');
    if (fullCode.length !== 6) return;

    setIsVerifying(true);
    
    // Simulate verification
    setTimeout(() => {
      // Demo: fail if code is 000000
      if (fullCode === '000000') {
        setIsVerifying(false);
        setHasError(true);
        // Clear code on error for UX? Or keep it? keeping it is better usually
      } else {
        onSuccess();
      }
    }, 1500);
  };

  return (
    <div className="flex-1 flex flex-col items-center justify-center p-4 relative">
      {/* Minimal Header */}
      <header className="absolute top-0 left-0 w-full p-6 flex justify-between items-center">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 bg-primary rounded-lg flex items-center justify-center text-white font-bold text-lg">
            <span className="material-symbols-outlined text-[20px]">local_shipping</span>
          </div>
          <span className="text-slate-900 dark:text-white font-semibold text-lg tracking-tight">FleetCommand</span>
        </div>
        <button className="text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200 text-sm font-medium transition-colors">
          Help & Support
        </button>
      </header>

      <main className="w-full max-w-[480px]">
        <div className="bg-surface-light dark:bg-surface-dark rounded-2xl shadow-card border border-slate-200 dark:border-slate-700 overflow-hidden">
          {/* Header Graphic */}
          <div className="h-32 w-full bg-slate-50 dark:bg-slate-800 flex items-center justify-center relative overflow-hidden">
            <div className="absolute inset-0 bg-primary/5 dark:bg-primary/10"></div>
            <div className="relative z-10 w-16 h-16 bg-white dark:bg-slate-700 rounded-full shadow-sm flex items-center justify-center border border-slate-100 dark:border-slate-600">
              <span className="material-symbols-outlined text-primary text-4xl">lock_person</span>
            </div>
            <div className="absolute bottom-0 w-full h-px bg-gradient-to-r from-transparent via-slate-200 dark:via-slate-600 to-transparent"></div>
          </div>

          <div className="p-8 pt-6">
            <div className="text-center mb-8">
              <h1 className="text-slate-900 dark:text-white text-2xl font-bold tracking-tight mb-2">Two-Step Verification</h1>
              <p className="text-slate-500 dark:text-slate-400 text-sm leading-relaxed">
                We sent a verification code to your registered authenticator app. Please enter the code below to continue.
              </p>
            </div>

            <form onSubmit={handleSubmit} className="flex flex-col gap-6">
              <div>
                <label className="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider mb-3 text-center">
                  Authentication Code
                </label>
                <div className="flex justify-center gap-2 sm:gap-3">
                  {code.map((digit, index) => (
                    <React.Fragment key={index}>
                      <input
                        ref={el => inputRefs.current[index] = el}
                        type="text"
                        inputMode="numeric"
                        maxLength={1}
                        value={digit}
                        onChange={(e) => handleChange(index, e.target.value)}
                        onKeyDown={(e) => handleKeyDown(index, e)}
                        onPaste={handlePaste}
                        placeholder="-"
                        className={`w-10 h-12 sm:w-12 sm:h-14 text-center text-xl font-bold rounded-lg focus:outline-none transition-all placeholder:text-slate-300 dark:placeholder:text-slate-600
                          ${hasError
                            ? 'text-error bg-red-50 dark:bg-red-900/10 border border-error focus:ring-2 focus:ring-error'
                            : 'text-slate-900 dark:text-white bg-slate-50 dark:bg-slate-800 border border-slate-300 dark:border-slate-600 focus:ring-2 focus:ring-primary focus:border-primary'
                          }`}
                      />
                      {index === 2 && (
                        <div className="flex items-center justify-center w-2 text-slate-300 dark:text-slate-600">
                          <span className="w-full h-0.5 bg-current rounded-full"></span>
                        </div>
                      )}
                    </React.Fragment>
                  ))}
                </div>

                {hasError && (
                  <div className="mt-4 text-center animate-pulse">
                    <p className="text-error text-sm font-medium flex items-center justify-center gap-1.5 bg-red-50 dark:bg-red-900/20 py-2 px-3 rounded-md border border-red-100 dark:border-red-900/30">
                      <span className="material-symbols-outlined text-[18px]">error</span>
                      Invalid verification code. Please try again.
                    </p>
                  </div>
                )}
              </div>

              <div className="flex items-center justify-center gap-3 py-1">
                <div className="relative flex items-center">
                  <input
                    type="checkbox"
                    id="trust_device"
                    className="peer h-5 w-5 cursor-pointer appearance-none rounded border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 checked:bg-primary checked:border-primary focus:ring-2 focus:ring-primary focus:ring-offset-1 dark:focus:ring-offset-slate-900 transition-all"
                  />
                  <span className="pointer-events-none absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 text-white opacity-0 peer-checked:opacity-100">
                    <svg className="h-3.5 w-3.5" fill="currentColor" viewBox="0 0 20 20">
                      <path clipRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" fillRule="evenodd"></path>
                    </svg>
                  </span>
                </div>
                <label htmlFor="trust_device" className="text-sm font-medium text-slate-700 dark:text-slate-300 cursor-pointer select-none">
                  Trust this device for 30 days
                </label>
              </div>

              <div className="flex flex-col gap-3 mt-2">
                <button
                  type="submit"
                  disabled={isVerifying}
                  className="w-full bg-primary hover:bg-primary-dark text-white font-semibold py-3 px-4 rounded-lg shadow-sm transition-all active:scale-[0.98] focus:ring-4 focus:ring-primary/20 flex items-center justify-center gap-2 disabled:opacity-80 disabled:cursor-not-allowed"
                >
                  {isVerifying ? (
                    <>
                      <svg className="animate-spin h-5 w-5 text-white" fill="none" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                        <path className="opacity-75" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" fill="currentColor"></path>
                      </svg>
                      Verifying...
                    </>
                  ) : (
                    'Verify Identity'
                  )}
                </button>
                <button
                  type="button"
                  onClick={onCancel}
                  className="w-full bg-transparent hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-600 dark:text-slate-400 font-medium py-3 px-4 rounded-lg transition-all"
                >
                  Cancel
                </button>
              </div>
            </form>

            <div className="mt-8 pt-6 border-t border-slate-100 dark:border-slate-700 flex flex-col items-center gap-4">
              <div className="text-center">
                <p className="text-sm text-slate-500 dark:text-slate-400 mb-1">Didn't receive a code?</p>
                <div className="flex items-center justify-center gap-1 text-sm">
                  <span className="text-slate-400 dark:text-slate-500">Resend code in</span>
                  <span className="font-mono font-medium text-primary">00:28</span>
                </div>
              </div>
              <a href="#" className="inline-flex items-center gap-1 text-sm font-medium text-primary hover:text-primary-dark transition-colors group">
                Try another method
                <span className="material-symbols-outlined text-[16px] transition-transform group-hover:translate-x-0.5">arrow_forward</span>
              </a>
            </div>
          </div>
          
          {/* Progress Bar */}
          <div className="h-1 w-full bg-slate-100 dark:bg-slate-800">
            <div className="h-full bg-primary w-2/3 rounded-r-full"></div>
          </div>
        </div>
        
        <p className="text-center text-xs text-slate-400 dark:text-slate-500 mt-8">
          © 2026 FleetCommand Enterprise. Secure Login.
        </p>

        {/* Demo Hint */}
        <div className="fixed bottom-4 right-4 bg-slate-800 text-slate-400 text-xs p-2 rounded opacity-50 hover:opacity-100 transition-opacity">
          Hint: Use 000000 to see error.
        </div>
      </main>
    </div>
  );
};

export default MFAScreen;