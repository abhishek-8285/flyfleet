import React, { useState } from "react";
import apiClient from "../services/apiClient";

const scenarios = [
  {
    id: "NORMAL",
    label: "Nominal Operations",
    icon: "check_circle",
    color: "text-emerald-500",
  },
  {
    id: "CRISIS",
    label: "Fleet Crisis Mode",
    icon: "warning",
    color: "text-rose-500",
  },
  {
    id: "MAINTENANCE",
    label: "Maintenance Peak",
    icon: "build",
    color: "text-amber-500",
  },
  {
    id: "NETWORK_FAILURE",
    label: "API Degradation",
    icon: "signal_wifi_off",
    color: "text-white/40",
  },
];

const ScenarioSwitcher: React.FC = () => {
  const [activeScenario, setActiveScenario] = useState("NORMAL");
  const [isOpen, setIsOpen] = useState(false);
  const [loading, setLoading] = useState(false);

  const switchScenario = async (type: string) => {
    setLoading(true);
    try {
      if (type === "NORMAL") {
        await apiClient.post("/mock/scenario/reset");
      } else {
        await apiClient.post("/mock/scenario/start", { type });
      }
      setActiveScenario(type);
    } catch (err) {
      console.error("Failed to switch scenario", err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed bottom-8 left-8 z-[9999]">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className={`w-14 h-14 rounded-full flex items-center justify-center shadow-2xl transition-all duration-500 ${isOpen ? "bg-white rotate-180" : "bg-primary hover:scale-110"}`}
      >
        <span
          className={`material-symbols-outlined ${isOpen ? "text-[#001219]" : "text-white"} text-[28px]`}
        >
          {isOpen ? "close" : "science"}
        </span>
      </button>

      {isOpen && (
        <div className="absolute bottom-20 left-0 w-72 bg-[#001219]/90 backdrop-blur-2xl border border-white/10 rounded-3xl p-6 shadow-2xl animate-fade-in-up">
          <h3 className="text-white font-bold text-[14px] uppercase tracking-wider mb-4 opacity-60">
            Scenario Intelligence
          </h3>
          <div className="space-y-3">
            {scenarios.map((s) => (
              <button
                key={s.id}
                disabled={loading}
                onClick={() => switchScenario(s.id)}
                className={`w-full flex items-center justify-between p-3 rounded-2xl border transition-all ${
                  activeScenario === s.id
                    ? "bg-white/10 border-white/20"
                    : "bg-white/5 border-transparent hover:border-white/10"
                }`}
              >
                <div className="flex items-center gap-3">
                  <span
                    className={`material-symbols-outlined text-[20px] ${s.color}`}
                  >
                    {s.icon}
                  </span>
                  <span className="text-white text-[13px] font-medium">
                    {s.label}
                  </span>
                </div>
                {activeScenario === s.id && (
                  <span className="w-2 h-2 rounded-full bg-primary shadow-[0_0_8px_rgba(19,127,236,0.5)]" />
                )}
              </button>
            ))}
          </div>
          <div className="mt-6 pt-4 border-t border-white/5">
            <div className="flex items-center justify-between text-[11px] font-bold uppercase tracking-wider text-white/30">
              <span>Mock Status</span>
              <span className="text-emerald-500">Stateful v3</span>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default ScenarioSwitcher;
