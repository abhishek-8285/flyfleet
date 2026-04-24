import React from "react";

interface InputProps {
  name: string;
  value: string;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onBlur?: () => void;
  label: string;
  placeholder?: string;
  type?: string;
  icon?: string;
  error?: string;
  required?: boolean;
  hint?: string;
  className?: string;
  maxLength?: number;
  disabled?: boolean;
}

export const TextInput: React.FC<InputProps> = ({
  name, value, onChange, onBlur, label, placeholder, type = "text",
  icon, error, required, hint, className, maxLength, disabled,
}) => (
  <div className={`space-y-1.5 ${className || ""}`}>
    <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
      {label} {required && <span className="text-danger">*</span>}
    </label>
    <div className="relative">
      {icon && (
        <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
          <span className="material-symbols-outlined text-[18px]">{icon}</span>
        </span>
      )}
      <input
        type={type}
        name={name}
        value={value}
        onChange={onChange}
        onBlur={onBlur}
        placeholder={placeholder}
        maxLength={maxLength}
        disabled={disabled}
        className={`w-full bg-surface-darker border rounded-lg py-2.5 ${
          icon ? "pl-10" : "px-4"
        } pr-4 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none placeholder-text-muted/50 ${
          error
            ? "border-danger/50 focus:ring-danger focus:border-danger"
            : "border-border-color"
        } ${disabled ? "opacity-50 cursor-not-allowed" : ""}`}
      />
    </div>
    {error ? (
      <p className="text-[10px] text-danger flex items-center gap-1 mt-1">
        <span className="material-symbols-outlined text-[12px]">error</span>
        {error}
      </p>
    ) : hint ? (
      <p className="text-[10px] text-text-muted mt-1">{hint}</p>
    ) : null}
  </div>
);

interface SelectOption {
  value: string;
  label: string;
}

interface SelectProps {
  name: string;
  value: string;
  onChange: (e: React.ChangeEvent<HTMLSelectElement>) => void;
  onBlur?: () => void;
  label: string;
  options: SelectOption[];
  icon?: string;
  error?: string;
  required?: boolean;
  placeholder?: string;
  className?: string;
}

export const SelectInput: React.FC<SelectProps> = ({
  name, value, onChange, onBlur, label, options, icon, error, required, placeholder, className,
}) => (
  <div className={`space-y-1.5 ${className || ""}`}>
    <label className="text-xs font-medium text-text-muted uppercase tracking-wide">
      {label} {required && <span className="text-danger">*</span>}
    </label>
    <div className="relative">
      {icon && (
        <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-text-muted">
          <span className="material-symbols-outlined text-[18px]">{icon}</span>
        </span>
      )}
      <select
        name={name}
        value={value}
        onChange={onChange}
        onBlur={onBlur}
        className={`w-full bg-surface-darker border rounded-lg py-2.5 ${
          icon ? "pl-10" : "pl-4"
        } pr-10 text-sm text-white focus:ring-1 focus:ring-primary focus:border-primary outline-none appearance-none ${
          error
            ? "border-danger/50 focus:ring-danger focus:border-danger"
            : "border-border-color"
        } ${!value ? "text-text-muted/50" : ""}`}
      >
        {placeholder && <option value="">{placeholder}</option>}
        {options.map((o) => (
          <option key={o.value} value={o.value}>{o.label}</option>
        ))}
      </select>
      <span className="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none text-text-muted">
        <span className="material-symbols-outlined text-[18px]">expand_more</span>
      </span>
    </div>
    {error && (
      <p className="text-[10px] text-danger flex items-center gap-1 mt-1">
        <span className="material-symbols-outlined text-[12px]">error</span>
        {error}
      </p>
    )}
  </div>
);

interface ToggleProps {
  label: string;
  description?: string;
  checked: boolean;
  onChange: (checked: boolean) => void;
  icon?: string;
}

export const ToggleInput: React.FC<ToggleProps> = ({
  label, description, checked, onChange, icon,
}) => (
  <div className="flex items-center justify-between p-4 bg-surface-darker rounded-lg border border-border-color">
    <div className="flex items-center gap-3">
      {icon && (
        <span className="material-symbols-outlined text-text-muted">{icon}</span>
      )}
      <div>
        <div className="text-sm font-medium text-white">{label}</div>
        {description && (
          <div className="text-[10px] text-text-muted mt-0.5">{description}</div>
        )}
      </div>
    </div>
    <label className="relative inline-flex items-center cursor-pointer">
      <input
        type="checkbox"
        checked={checked}
        onChange={(e) => onChange(e.target.checked)}
        className="sr-only peer"
      />
      <div className="w-9 h-5 bg-border-color peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-primary/30 rounded-full peer peer-checked:after:translate-x-full after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-primary"></div>
    </label>
  </div>
);
