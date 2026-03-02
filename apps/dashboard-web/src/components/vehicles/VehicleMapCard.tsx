import React from "react";
import { useNavigate } from "react-router-dom";

import { useUIStrings } from "../../hooks/useUIStrings";

export const VehicleMapCard: React.FC = () => {
  const navigate = useNavigate();
  const { uiStrings } = useUIStrings("fleet");

  return (
    <div className="bg-surface-dark rounded-xl border border-border-color shadow-sm flex flex-col flex-grow relative group/map overflow-hidden min-h-[400px]">
      <div className="p-3 border-b border-border-color flex justify-between items-center bg-surface-darker z-10 relative">
        <h3 className="text-white font-semibold flex items-center gap-2 text-sm">
          <span className="material-symbols-outlined text-primary text-[18px]">
            location_on
          </span>
          {uiStrings?.current_location || "Current Location & Route"}
        </h3>
        <div className="flex items-center gap-2">
          <span className="text-[10px] text-text-muted">
            {uiStrings?.last_update || "Last update: 3s ago"}
          </span>
          <span className="h-1.5 w-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
        </div>
      </div>
      <div className="relative flex-1 bg-[#0f151b] overflow-hidden">
        <div
          className="absolute inset-0 bg-cover bg-center opacity-60"
          style={{
            backgroundImage:
              'url("https://lh3.googleusercontent.com/aida-public/AB6AXuCeeeLxaUXFSywPWOUz7sSShAB3uQmjmm05fXrxnJMrtOj-77MaaA_rwghU8aFYgdEN25a4Zu_KRt4Bww_5vIegatJal2ZEZwc-jWi6kSXdQy3FKD1_7k4VpFPnvvH8JhhczyN8RmE4dQe7Igd7R_Y3Q2A8cUV3JoEubwD_1yrEB0VImkaOpvLuj7MtoMf4v4Vfp7EmLAR1LIeGUzUqVgEPm__oaIqEyfwohYrEJheP1zpnGNTy0WkwKZgcqRa3hJZyETmgEvoDzR4")',
          }}
        ></div>
        <svg
          className="absolute inset-0 w-full h-full pointer-events-none"
          preserveAspectRatio="none"
          viewBox="0 0 100 100"
        >
          <path
            className="opacity-60"
            d="M10,80 Q30,70 40,50 T70,30"
            fill="none"
            stroke="#137fec"
            strokeDasharray="2"
            strokeWidth="0.5"
          ></path>
          <circle cx="10" cy="80" fill="#64748b" opacity="0.8" r="1"></circle>
          <circle cx="40" cy="50" fill="#64748b" opacity="0.8" r="1"></circle>
        </svg>
        <div className="absolute top-[30%] left-[70%] transform -translate-x-1/2 -translate-y-1/2 group cursor-pointer z-10">
          <div className="w-10 h-10 bg-primary/20 rounded-full flex items-center justify-center animate-pulse">
            <div className="w-4 h-4 bg-primary rounded-full shadow-[0_0_15px_#137fec] ring-2 ring-white"></div>
          </div>
          <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 bg-surface-darker text-xs text-white p-2 rounded shadow-lg whitespace-nowrap border border-border-color z-20">
            <div className="font-bold">65 mph</div>
            <div className="text-text-muted text-[10px]">Heading NW</div>
          </div>
        </div>
        <div className="absolute bottom-4 left-4 right-4 bg-surface-darker/90 backdrop-blur-md border border-border-color rounded-lg p-3 z-20">
          <div className="flex justify-between items-center mb-2">
            <span className="text-xs font-medium text-white">
              {uiStrings?.history_trail || "24h History Trail"}
            </span>
            <span className="text-[10px] text-text-muted">382 miles</span>
          </div>
          <div className="relative w-full h-1 bg-border-color rounded-full">
            <div className="absolute left-0 top-0 h-full w-[70%] bg-primary rounded-full"></div>
            <div className="absolute left-[70%] top-1/2 -translate-y-1/2 w-3 h-3 bg-white rounded-full shadow cursor-pointer hover:scale-110 transition-transform"></div>
          </div>
          <div className="flex justify-between text-[10px] text-text-muted mt-1">
            <span>00:00</span>
            <span>12:00</span>
            <span>{uiStrings?.now || "Now"}</span>
          </div>
        </div>
      </div>
      <div className="bg-surface-dark rounded-xl border border-border-color p-4 shadow-sm">
        <div className="flex items-start gap-3">
          <div className="p-2 bg-border-color/30 rounded text-text-muted">
            <span className="material-symbols-outlined">signpost</span>
          </div>
          <div>
            <h4 className="text-xs text-text-muted uppercase font-semibold">
              {uiStrings?.current_address || "Current Address"}
            </h4>
            <p className="text-white text-sm font-medium mt-1">
              I-95 South, Exit 4, New Haven, CT
            </p>
            <p className="text-xs text-text-muted mt-0.5">
              Lat: 41.3083° N, Lon: 72.9279° W
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};
