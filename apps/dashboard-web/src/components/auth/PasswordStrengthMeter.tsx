import { motion } from "framer-motion";
import React from "react";

interface PasswordStrengthMeterProps {
  strength: number;
  password?: string;
  label?: string;
  strengthLabels?: {
    weak: string;
    fair: string;
    good: string;
    strong: string;
  };
}

const PasswordStrengthMeter: React.FC<PasswordStrengthMeterProps> = ({
  strength,
  password,
  label = "Security Strength",
  strengthLabels = {
    weak: "Weak",
    fair: "Fair",
    good: "Good",
    strong: "Strong",
  },
}) => {
  const isEmpty = password !== undefined ? password.length === 0 : false;

  const getStrengthColor = (s: number) => {
    if (isEmpty) return "bg-transparent";
    if (s <= 25) return "bg-red-500";
    if (s <= 50) return "bg-yellow-500";
    if (s <= 75) return "bg-blue-500";
    return "bg-green-500";
  };

  const getStrengthText = (s: number) => {
    if (isEmpty) return "";
    if (s <= 25) return strengthLabels.weak;
    if (s <= 50) return strengthLabels.fair;
    if (s <= 75) return strengthLabels.good;
    return strengthLabels.strong;
  };

  const getTextColor = (s: number) => {
    if (isEmpty) return "text-slate-600";
    if (s <= 25) return "text-red-500";
    if (s <= 50) return "text-yellow-500";
    if (s <= 75) return "text-blue-500";
    return "text-green-500";
  };

  return (
    <div className="mt-2 px-1">
      <div className="flex justify-between items-center mb-1">
        <span className="text-[10px] font-bold text-slate-400 uppercase tracking-widest">
          {label}
        </span>
        <span
          className={`text-[10px] font-bold uppercase tracking-widest ${getTextColor(strength)}`}
        >
          {getStrengthText(strength) || "NONE"}
        </span>
      </div>
      <div className="h-1 w-full bg-slate-800 rounded-full overflow-hidden">
        <motion.div
          initial={{ width: 0 }}
          animate={{ width: isEmpty ? 0 : `${Math.max(strength, 5)}%` }}
          className={`h-full transition-colors duration-500 ${getStrengthColor(strength)}`}
        />
      </div>
    </div>
  );
};

export default PasswordStrengthMeter;
