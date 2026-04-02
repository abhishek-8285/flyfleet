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
      <label className="block text-xs font-semibold text-slate-600 dark:text-slate-300 uppercase tracking-wide ml-1">
        {label}
      </label>
      <div className="relative group">
        <span className="absolute left-3 top-1/2 -translate-y-1/2 material-symbols-outlined text-slate-400 dark:text-slate-500 group-focus-within:text-blue-600 dark:group-focus-within:text-white transition-colors text-lg font-light">
          {icon}
        </span>
        <input
          {...props}
          className={`w-full bg-white dark:bg-transparent border border-slate-200 dark:border-slate-700 rounded-md pl-10 ${rightElement ? "pr-10" : "pr-3"} py-3 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-blue-600 dark:focus:ring-white dark:focus:border-white transition-all placeholder:text-slate-400 dark:placeholder:text-slate-500 ${props.className || ""}`}
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
