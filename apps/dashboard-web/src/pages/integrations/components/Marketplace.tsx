import {
    ArrowRightIcon,
    ChatBubbleLeftRightIcon,
    CloudIcon,
    Cog6ToothIcon,
    CubeIcon,
    ExclamationCircleIcon,
    MapIcon,
} from "@heroicons/react/24/solid";
import { motion } from "framer-motion";
import React from "react";
import { useUIStrings } from "../../../hooks/useUIStrings";

interface Integration {
  id: string;
  name: string;
  type: string;
  provider: string;
  status: "connected" | "error" | "available";
  description: string;
  lastSync?: string;
}

interface MarketplaceProps {
  onConfigure: (id: string) => void;
}

const Marketplace: React.FC<MarketplaceProps> = ({ onConfigure }) => {
  const [marketplaceIntegrations, setMarketplaceintegrations] = React.useState<
    any[]
  >([]);
  const [loading_marketplaceIntegrations, setLoading_marketplaceIntegrations] =
    React.useState(true);
  const { uiStrings } = useUIStrings("integrations");

  React.useEffect(() => {
    const fetchMarketplaceintegrations = async () => {
      try {
        const res = await fetch(
          `${process.env.REACT_APP_API_URL || "/api/v1"}/integrations/marketplace`,
        );
        if (res.ok) {
          const payload = await res.json();
          setMarketplaceintegrations(payload.data || payload || []);
        }
      } catch (error) {
        console.error("Failed to fetch marketplaceIntegrations", error);
      } finally {
        setLoading_marketplaceIntegrations(false);
      }
    };
    fetchMarketplaceintegrations();
  }, []);

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
      {marketplaceIntegrations.map((item, index) => (
        <motion.div
          key={item.id}
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ delay: index * 0.05 }}
          className="bg-slate-900/40 border border-slate-800 rounded-2xl p-6 hover:border-primary/50 transition-all group relative overflow-hidden backdrop-blur-sm"
        >
          {/* Status Badge */}
          <div className="absolute top-4 right-4">
            {item.status === "connected" && (
              <span className="flex items-center gap-1.5 px-2 py-1 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-[10px] font-bold text-emerald-500 uppercase tracking-tighter">
                <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
                {uiStrings?.status_connected || "Connected"}
              </span>
            )}
            {item.status === "error" && (
              <span className="flex items-center gap-1.5 px-2 py-1 rounded-lg bg-rose-500/10 border border-rose-500/20 text-[10px] font-bold text-rose-500 uppercase tracking-tighter">
                <ExclamationCircleIcon className="w-3 h-3" />
                {uiStrings?.status_error || "Error"}
              </span>
            )}
            {item.status === "available" && (
              <span className="flex items-center gap-1.5 px-2 py-1 rounded-lg bg-slate-800 text-[10px] font-bold text-slate-400 uppercase tracking-tighter">
                {uiStrings?.status_available || "Available"}
              </span>
            )}
          </div>

          {/* Icon & Title */}
          <div className="flex items-start gap-4 mb-4">
            <div
              className={`w-12 h-12 rounded-xl flex items-center justify-center text-white shadow-lg ${
                item.status === "available"
                  ? "bg-slate-800 grayscale"
                  : "bg-slate-800 group-hover:scale-110 transition-transform"
              }`}
            >
              {item.type === "SMS Gateway" && (
                <ChatBubbleLeftRightIcon className="w-6 h-6 text-primary" />
              )}
              {item.type === "Cloud Provider" && (
                <CloudIcon className="w-6 h-6 text-amber-500" />
              )}
              {item.type === "ERP System" && (
                <CubeIcon className="w-6 h-6 text-indigo-500" />
              )}
              {item.type === "Maps API" && (
                <MapIcon className="w-6 h-6 text-emerald-500" />
              )}
            </div>
            <div>
              <h3 className="font-bold text-white group-hover:text-primary transition-colors">
                {item.name}
              </h3>
              <p className="text-[10px] text-slate-500 uppercase font-semibold tracking-wider font-mono">
                {item.type}
              </p>
            </div>
          </div>

          <p className="text-sm text-slate-400 mb-6 line-clamp-2 min-h-[40px]">
            {item.description}
          </p>

          {/* Metadata Grid */}
          <div className="grid grid-cols-2 gap-2 mb-6 p-3 bg-slate-950/50 rounded-xl border border-white/5">
            <div>
              <span className="block text-[10px] text-slate-600 uppercase font-bold mb-0.5">
                {uiStrings?.label_provider || "Provider"}
              </span>
              <span className="text-xs text-white font-medium">
                {item.provider}
              </span>
            </div>
            <div>
              <span className="block text-[10px] text-slate-600 uppercase font-bold mb-0.5">
                {uiStrings?.label_sync_status || "Sync Status"}
              </span>
              <span
                className={`text-xs font-medium ${item.status === "error" ? "text-rose-400" : "text-slate-300"}`}
              >
                {item.lastSync || uiStrings?.label_never || "Never"}
              </span>
            </div>
          </div>

          <div className="flex gap-2">
            {item.status === "available" ? (
              <button className="w-full py-2 bg-white/5 hover:bg-primary text-white text-xs font-semibold rounded-xl border border-white/10 hover:border-primary transition-all">
                {uiStrings?.enable_integration_btn || "Enable Integration"}
              </button>
            ) : (
              <>
                <button
                  onClick={() => onConfigure(item.id)}
                  className="flex-1 py-2 bg-slate-800 hover:bg-slate-700 text-white text-xs font-semibold rounded-xl border border-white/5 flex items-center justify-center gap-2 transition-all"
                >
                  <Cog6ToothIcon className="w-4 h-4" />
                  {uiStrings?.configure_btn || "Configure"}
                </button>
                <button className="px-3 py-2 bg-slate-800 hover:bg-slate-700 text-slate-400 hover:text-white rounded-xl border border-white/5 transition-all">
                  <ArrowRightIcon className="w-4 h-4" />
                </button>
              </>
            )}
          </div>
        </motion.div>
      ))}
    </div>
  );
};

export default Marketplace;
