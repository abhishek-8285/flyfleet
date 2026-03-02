import React from "react";
import { useUIStrings } from "../hooks/useUIStrings";

const SecurityFooter: React.FC<{ systemMeta?: any }> = ({ systemMeta }) => {
  const { uiStrings } = useUIStrings("layout_security");
  return (
    <footer className="bg-surface-darker border-t border-border-color px-6 py-2 flex items-center justify-between text-[10px] text-text-muted z-20 mt-auto flex-shrink-0">
      <span>
        © {new Date().getFullYear()}{" "}
        {systemMeta?.company_name || "Enterprise Fleet Manager"}.{" "}
        {systemMeta?.footer_tagline || "Security Module"}{" "}
        {systemMeta?.version || "V--"}
      </span>
      <div className="flex items-center gap-4">
        <span>
          {uiStrings?.encryption_label || "Encryption:"}{" "}
          <span className="text-success">
            {systemMeta?.encryption_status || "AES-256"}
          </span>
        </span>
        <span>
          {uiStrings?.status_label || "System Status:"}{" "}
          <span className="text-success">
            {systemMeta?.status || "Operational"}
          </span>
        </span>
        <span>
          {uiStrings?.time_label || "Server Time:"}{" "}
          <span className="font-mono text-white">
            {systemMeta?.server_time || new Date().toLocaleTimeString()} UTC
          </span>
        </span>
      </div>
    </footer>
  );
};

export default SecurityFooter;
