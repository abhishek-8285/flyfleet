import React, { Component, ErrorInfo, ReactNode } from "react";
import { AlertTriangle, RotateCcw, Home } from "lucide-react";

interface Props {
  children?: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
    error: null
  };

  public static getDerivedStateFromError(error: Error): State {
    // Update state so the next render will show the fallback UI.
    return { hasError: true, error };
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error("Uncaught UI Exception:", error, errorInfo);
  }

  private handleReset = () => {
    this.setState({ hasError: false, error: null });
    window.location.reload();
  };

  private handleGoHome = () => {
    this.setState({ hasError: false, error: null });
    window.location.href = "/";
  };

  public render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen bg-[#0a0a0b] flex items-center justify-center p-6 font-sans">
          <div className="max-w-xl w-full bg-[#0d0d0f] border border-white/5 rounded-[3rem] p-12 text-center space-y-8 shadow-2xl relative overflow-hidden">
            {/* Background Glow */}
            <div className="absolute -top-24 -right-24 w-48 h-48 bg-danger/10 rounded-full blur-3xl" />
            <div className="absolute -bottom-24 -left-24 w-48 h-48 bg-primary/5 rounded-full blur-3xl" />

            <div className="relative space-y-8">
              <div className="w-20 h-20 bg-danger/10 rounded-3xl flex items-center justify-center mx-auto border border-danger/20 transform rotate-12">
                <AlertTriangle className="text-danger w-10 h-10 -rotate-12" />
              </div>

              <div className="space-y-3">
                <h2 className="text-2xl font-black text-white tracking-tight uppercase italic">
                  Runtime <span className="text-danger">Exception</span> Detected
                </h2>
                <p className="text-sm text-text-muted leading-relaxed max-w-sm mx-auto">
                  A critical synchronization error occurred in the UI renderer. The session remains secure, but a reload is required to restore the operational interface.
                </p>
              </div>

              {/* Error Detail (Miniature) */}
              <div className="bg-white/5 border border-white/5 rounded-2xl p-4 text-left">
                <p className="text-[10px] font-bold text-text-muted uppercase tracking-widest mb-2 flex items-center gap-2">
                  <div className="w-1.5 h-1.5 rounded-full bg-danger" /> Diagnostic Trace
                </p>
                <p className="text-[11px] font-mono text-danger/80 break-all line-clamp-2">
                  {this.state.error?.message || "Unknown rendering conflict"}
                </p>
              </div>

              <div className="grid grid-cols-2 gap-4 pt-4">
                <button 
                  onClick={this.handleReset}
                  className="py-4 bg-danger text-white rounded-2xl text-xs font-black transition-all flex items-center justify-center gap-3 active:scale-95 shadow-lg shadow-danger/20 hover:bg-danger/90"
                >
                  <RotateCcw className="w-4 h-4" /> REBOOT UI
                </button>
                <button 
                  onClick={this.handleGoHome}
                  className="py-4 bg-white/5 hover:bg-white/10 text-white rounded-2xl text-xs font-bold transition-all flex items-center justify-center gap-3 border border-white/5 active:scale-95"
                >
                  <Home className="w-4 h-4" /> HOMEBASE
                </button>
              </div>
              
              <p className="text-[10px] text-white/20 font-bold uppercase tracking-[0.3em]">
                System Recovery Protocol // 0xCC-FLT
              </p>
            </div>
          </div>
        </div>
      );
    }

    return this.children;
  }
}

export default ErrorBoundary;
