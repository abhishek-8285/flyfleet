import {
    ClockIcon,
    MagnifyingGlassIcon,
    UserCircleIcon,
} from "@heroicons/react/24/outline";
import React, { useEffect, useState } from "react";
import { api } from "../../services/api";

const EntityHistory: React.FC = () => {
  const [entityType, setEntityType] = useState("vehicle");
  const [entityId, setEntityId] = useState("");
  const [history, setHistory] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [uiStrings, setUiStrings] = useState<any>(null);

  useEffect(() => {
    fetch("/api/v1/ui/global_config")
      .then((res) => res.json())
      .then((data) =>
        setUiStrings(
          data.ui_strings?.governance ||
            data.system_metadata?.ui_strings?.governance,
        ),
      )
      .catch((err) => console.error("Failed to fetch UI strings", err));
  }, []);

  const fetchHistory = async () => {
    if (!entityId) return;
    setLoading(true);
    try {
      const data = await api.getEntityHistory(entityType, entityId);
      setHistory(data || []);
    } catch (error) {
      console.error("Failed to fetch entity history", error);
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return (
      date.toLocaleDateString("en-US", {
        month: "short",
        day: "numeric",
        year: "numeric",
      }) +
      " " +
      date.toLocaleTimeString("en-US", { hour: "2-digit", minute: "2-digit" })
    );
  };

  return (
    <div className="space-y-8 max-w-5xl mx-auto">
      {/* Search Header */}
      <div className="bg-slate-900/60 backdrop-blur-xl p-8 rounded-3xl border border-white/5 space-y-6 shadow-2xl">
        <div>
          <h2 className="text-2xl font-semibold text-white mb-2 font-display">
            {uiStrings?.temporal_insights || "Temporal Insights"}
          </h2>
          <p className="text-slate-400 font-light">
            {uiStrings?.temporal_desc ||
              "Trace the evolutionary history of any system entity across time"}
          </p>
        </div>

        <div className="flex flex-col md:flex-row gap-4">
          <div className="flex-1 space-y-2">
            <label className="text-xs font-bold text-slate-500 uppercase tracking-widest px-1">
              {uiStrings?.entity_domain || "Entity Domain"}
            </label>
            <select
              value={entityType}
              onChange={(e) => setEntityType(e.target.value)}
              className="w-full px-4 py-3 bg-slate-950 border border-slate-800 rounded-2xl text-white appearance-none focus:ring-2 focus:ring-blue-500/50 transition-all font-light"
            >
              <option value="vehicle">
                {uiStrings?.dom_vehicle || "Vehicle (Fleet Portfolio)"}
              </option>
              <option value="user">
                {uiStrings?.dom_user || "User (Identity Engine)"}
              </option>
              <option value="device">
                {uiStrings?.dom_device || "IoT Device (Edge Computing)"}
              </option>
              <option value="organization">
                {uiStrings?.dom_org || "Organization (Core Entity)"}
              </option>
            </select>
          </div>
          <div className="flex-[2] space-y-2">
            <label className="text-xs font-bold text-slate-500 uppercase tracking-widest px-1">
              {uiStrings?.uuid_label || "Universal Identifier (UUID)"}
            </label>
            <div className="relative">
              <MagnifyingGlassIcon className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-500" />
              <input
                type="text"
                placeholder={uiStrings?.id_placeholder || "Enter entity ID..."}
                value={entityId}
                onChange={(e) => setEntityId(e.target.value)}
                className="w-full pl-12 pr-4 py-3 bg-slate-950 border border-slate-800 rounded-2xl text-white placeholder:text-slate-600 focus:ring-2 focus:ring-blue-500/50 transition-all font-light"
              />
            </div>
          </div>
          <div className="flex flex-col justify-end">
            <button
              onClick={fetchHistory}
              disabled={!entityId || loading}
              className="px-8 py-3 bg-blue-600 hover:bg-blue-500 text-white rounded-2xl font-semibold transition-all shadow-lg shadow-blue-500/20 disabled:opacity-50 h-[50px] mt-auto"
            >
              {loading
                ? uiStrings?.searching_btn || "Searching..."
                : uiStrings?.analyze_btn || "Analyze Trace"}
            </button>
          </div>
        </div>
      </div>

      {/* Timeline View */}
      <div className="relative pl-8 space-y-8 before:absolute before:left-[11px] before:top-2 before:bottom-2 before:w-[2px] before:bg-slate-800 before:content-['']">
        {history.length === 0 ? (
          <div className="py-20 text-center space-y-4">
            <div className="w-16 h-16 bg-slate-900 border border-slate-800 rounded-2xl flex items-center justify-center mx-auto mb-4">
              <ClockIcon className="w-8 h-8 text-slate-500" />
            </div>
            <p className="text-slate-500 font-light italic">
              {uiStrings?.enter_id ||
                "Enter an entity ID to view its temporal history"}
            </p>
          </div>
        ) : (
          history.map((event, index) => (
            <div
              key={event.id}
              className="relative group animate-in slide-in-from-left duration-500"
              style={{ animationDelay: `${index * 100}ms` }}
            >
              {/* Dot */}
              <div className="absolute -left-[27px] top-2 w-4 h-4 rounded-full bg-slate-950 border-2 border-blue-500 shadow-[0_0_10px_rgba(59,130,246,0.5)] z-10 group-hover:scale-125 transition-transform" />

              {/* Content Card */}
              <div className="bg-slate-900/40 backdrop-blur-xl p-6 rounded-3xl border border-white/5 hover:border-blue-500/30 transition-all shadow-xl">
                <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-6">
                  <div className="space-y-1">
                    <div className="flex items-center gap-3">
                      <span className="text-white font-semibold">
                        {uiStrings?.transformation || "State Transformation"}
                      </span>
                      <span className="px-2 py-0.5 bg-slate-800 rounded text-[10px] text-slate-400 font-mono tracking-tighter">
                        {uiStrings?.version_prefix || "VER."}{" "}
                        {history.length - index}
                      </span>
                    </div>
                    <p className="text-xs text-slate-500">
                      {formatDate(event.timestamp)}
                    </p>
                  </div>
                  <div className="flex items-center gap-2 bg-slate-950/50 px-3 py-1.5 rounded-xl border border-slate-800">
                    <UserCircleIcon className="w-4 h-4 text-blue-400" />
                    <span className="text-xs text-slate-300 font-medium">
                      {uiStrings?.actor_prefix || "Actor:"} {event.actor}
                    </span>
                  </div>
                </div>

                <div className="bg-slate-950/80 rounded-2xl border border-slate-800 p-4 font-mono">
                  <pre className="text-xs text-slate-400 overflow-x-auto whitespace-pre-wrap">
                    {JSON.stringify(event.change_data, null, 2)}
                  </pre>
                </div>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default EntityHistory;
