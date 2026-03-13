import React from "react";

interface AuthInputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label: string;
  icon: string;
  rightElement?: React.ReactNode;
}

const AuthInput: React.FC<AuthInputProps> = ({
  label,
  icon,
  rightElement,
  ...props
}) => {
  return (
    <div className="space-y-1.5">
      <label className="block text-xs font-bold text-slate-400 uppercase tracking-widest ml-1">
        {label}
      </label>
      <div className="relative group">
        <span className="absolute left-4 top-1/2 -translate-y-1/2 material-symbols-outlined text-slate-500 group-focus-within:text-primary transition-colors text-xl">
          {icon}
        </span>
        <input
          {...props}
          className={`w-full bg-slate-800/30 border border-white/5 rounded-2xl pl-12 ${rightElement ? "pr-12" : "pr-4"} py-3 text-white focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all placeholder:text-slate-600 ${props.className || ""}`}
        />
        {rightElement && (
          <div className="absolute right-4 top-1/2 -translate-y-1/2">
            {rightElement}
          </div>
        )}
      </div>
    </div>
  );
};

export default AuthInput;
