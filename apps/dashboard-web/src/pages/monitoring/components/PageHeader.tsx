import React from "react";
import { useUIStrings } from "../../../hooks/useUIStrings";

interface PageHeaderProps {
  title: string;
  subtitle?: React.ReactNode;
  status?: string;
  statusColor?: "success" | "warning" | "danger" | "primary" | "neutral";
  actions?: React.ReactNode;
}

const PageHeader: React.FC<PageHeaderProps> = ({
  title,
  subtitle,
  status,
  statusColor = "success",
  actions,
}) => {
  const { uiStrings } = useUIStrings("monitoring");
  const getStatusColorClass = (color: string) => {
    switch (color) {
      case "success":
        return "text-monitor-success bg-monitor-success/20 border-monitor-success/30";
      case "warning":
        return "text-monitor-warning bg-monitor-warning/20 border-monitor-warning/30";
      case "danger":
        return "text-monitor-danger bg-monitor-danger/20 border-monitor-danger/30";
      case "primary":
        return "text-monitor-primary bg-monitor-primary/20 border-monitor-primary/30";
      default:
        return "text-monitor-text-muted bg-monitor-surface border-monitor-border";
    }
  };

  return (
    <header className="h-16 flex items-center justify-between px-6 border-b border-monitor-border bg-monitor-surface-darker flex-shrink-0 z-20">
      <div className="flex items-center gap-4">
        <div>
          {subtitle && (
            <div className="flex items-center gap-2 text-[10px] text-monitor-text-muted font-medium mb-0.5 uppercase tracking-wide">
              {subtitle}
            </div>
          )}
          <h2 className="text-white text-lg font-bold">{title}</h2>
          {!subtitle && (
            <p className="text-[11px] text-monitor-text-muted mt-0.5">
              {uiStrings?.footer_text ||
                "Enterprise Fleet Platform • Monitoring Module"}
            </p>
          )}
        </div>
        {status && (
          <span
            className={`px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider border flex items-center gap-1 ${getStatusColorClass(statusColor)}`}
          >
            {statusColor === "success" && (
              <span className="w-1.5 h-1.5 rounded-full bg-monitor-success animate-pulse"></span>
            )}
            {status}
          </span>
        )}
      </div>
      <div className="flex items-center gap-3">{actions}</div>
    </header>
  );
};

export default PageHeader;
