import React from "react";

interface AuthButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  loading?: boolean;
  loadingText?: string;
  variant?: "primary" | "secondary";
}

const AuthButton: React.FC<AuthButtonProps> = ({
  children,
  loading,
  loadingText,
  variant = "primary",
  ...props
}) => {
  const variantClasses =
    variant === "primary"
      ? "bg-gradient-to-r from-primary to-blue-600 hover:from-primary-dark hover:to-blue-700 text-white shadow-primary/20"
      : "bg-slate-800/50 hover:bg-slate-800 text-white border border-white/5";

  return (
    <button
      {...props}
      disabled={loading || props.disabled}
      className={`w-full ${variantClasses} font-bold py-4 rounded-2xl transition-all shadow-xl disabled:opacity-50 flex items-center justify-center space-x-2 text-sm uppercase tracking-widest ${props.className || ""}`}
    >
      {loading ? (
        <>
          <svg
            className="animate-spin h-5 w-5 text-white"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              className="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              strokeWidth="4"
            ></circle>
            <path
              className="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            ></path>
          </svg>
          <span>{loadingText || "Processing..."}</span>
        </>
      ) : (
        children
      )}
    </button>
  );
};

export default AuthButton;
