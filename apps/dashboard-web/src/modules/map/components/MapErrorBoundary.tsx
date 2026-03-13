import { AlertTriangle, RefreshCw } from "lucide-react";
import { Component, ErrorInfo, ReactNode } from "react";

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export class MapErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
    error: null,
  };

  public static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error("[MapErrorBoundary] Uncaught map error:", error, errorInfo);
  }

  private handleReset = () => {
    this.setState({ hasError: false, error: null });
    window.location.reload(); // Hard reset for WebGL context recovery if necessary
  };

  public render() {
    if (this.state.hasError) {
      return (
        <div className="absolute inset-0 z-[200] bg-slate-950/90 backdrop-blur-xl flex items-center justify-center p-6 text-center">
          <div className="max-w-md space-y-6 animate-in fade-in zoom-in duration-300">
            <div className="mx-auto w-16 h-16 bg-rose-500/20 rounded-full flex items-center justify-center border border-rose-500/50 shadow-[0_0_30px_rgba(244,63,94,0.3)]">
              <AlertTriangle className="text-rose-500 w-8 h-8" />
            </div>

            <div className="space-y-2">
              <h2 className="text-xl font-black text-white uppercase tracking-tight">
                Rendering Engine Failure
              </h2>
              <p className="text-sm text-slate-400 leading-relaxed">
                The hyperscale MapLibre engine encountered a critical error.
                This often happens if the WebGL context is lost or a worker
                thread hangs.
              </p>
            </div>

            <div className="bg-slate-900 border border-white/5 rounded-xl p-3 font-mono text-[10px] text-rose-300/70 text-left overflow-hidden">
              <code>{this.state.error?.message}</code>
            </div>

            <button
              onClick={this.handleReset}
              className="flex items-center gap-2 mx-auto px-6 py-3 bg-primary text-white text-xs font-black uppercase tracking-widest rounded-xl hover:bg-primary-dark transition-all transform active:scale-95 shadow-[0_10px_30px_rgba(37,99,235,0.4)]"
            >
              <RefreshCw className="w-4 h-4" />
              Re-initialize Engine
            </button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
