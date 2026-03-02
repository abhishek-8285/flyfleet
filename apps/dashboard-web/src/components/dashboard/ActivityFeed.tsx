import React, { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { useUIStrings } from "../../hooks/useUIStrings";
import { api } from "../../services/api";
import { TimelineEvent } from "../../types";

export const ActivityFeed: React.FC = () => {
  const { uiStrings } = useUIStrings("comp_dashboard");
  const [events, setEvents] = useState<TimelineEvent[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchEvents = async () => {
    try {
      const data = await api.getTimeline();
      setEvents(data.data || []);
    } catch (error) {
      console.error("Failed to fetch timeline events", error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchEvents();
    const interval = setInterval(fetchEvents, 15000); // 15s refresh for real-time feel
    return () => clearInterval(interval);
  }, []);

  const getIcon = (type: string) => {
    switch (type.toLowerCase()) {
      case "trip_started":
        return "play_arrow";
      case "alert":
        return "speed";
      case "delivery_complete":
        return "check_circle";
      case "notification":
        return "notifications";
      default:
        return "info";
    }
  };

  const getIconColor = (severity: string) => {
    switch (severity.toLowerCase()) {
      case "critical":
        return "bg-red-500/20 text-red-500";
      case "warning":
        return "bg-amber-500/20 text-amber-500";
      case "info":
        return "bg-primary/20 text-primary";
      default:
        return "bg-blue-400/20 text-blue-400";
    }
  };

  const formatTimestamp = (ts: string) => {
    const date = new Date(ts);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);

    if (diffMins < 1) return "Just now";
    if (diffMins < 60) return `${diffMins} mins ago`;
    const diffHours = Math.floor(diffMins / 60);
    if (diffHours < 24) return `${diffHours}h ago`;
    return date.toLocaleDateString();
  };

  return (
    <div className="bg-surface-dark rounded-xl border border-border-color flex-1 min-h-[400px] flex flex-col">
      <div className="p-4 border-b border-border-color bg-surface-darker flex justify-between items-center">
        <h3 className="text-white font-semibold flex items-center gap-2">
          <span className="material-symbols-outlined text-primary text-lg">
            history
          </span>
          {uiStrings?.title_activity || "Operation Stream"}
        </h3>
        <div className="flex items-center gap-2">
          {loading && (
            <span className="w-2 h-2 bg-primary rounded-full animate-ping"></span>
          )}
          <button className="text-xs text-primary hover:text-white bg-primary/10 px-2 py-1 rounded border border-primary/20">
            Real-time
          </button>
        </div>
      </div>
      <div className="p-4 space-y-6 flex-1 overflow-y-auto custom-scrollbar max-h-[500px]">
        {events.length === 0 && !loading ? (
          <div className="flex flex-col items-center justify-center h-full text-text-muted py-10">
            <span className="material-symbols-outlined text-4xl mb-2 opacity-20">
              inbox
            </span>
            <p className="text-xs">
              {uiStrings?.msg_no_data || "No recent activity detected"}
            </p>
          </div>
        ) : (
          events.map((event, idx) => (
            <div key={event.id || idx} className="flex gap-4 relative">
              <div className="flex flex-col items-center">
                <div
                  className={`w-8 h-8 rounded-full flex items-center justify-center z-10 ring-4 ring-surface-dark ${getIconColor(event.severity)}`}
                >
                  <span className="material-symbols-outlined text-[16px]">
                    {getIcon(event.type)}
                  </span>
                </div>
                {idx !== events.length - 1 && (
                  <div className="w-0.5 bg-border-color h-full absolute top-8"></div>
                )}
              </div>
              <div className="pb-2">
                <p className="text-sm text-white font-medium">{event.title}</p>
                <p className="text-xs text-text-muted mt-0.5">
                  {event.description}
                </p>
                <span className="text-[10px] text-[#64748b] mt-1 block">
                  {formatTimestamp(event.timestamp)}
                </span>
              </div>
            </div>
          ))
        )}
      </div>
      <div className="px-4 py-3 bg-surface-darker border-t border-border-color text-center">
        <Link
          className="text-xs text-primary hover:text-white"
          to="/dashboard/timeline"
        >
          {uiStrings?.btn_view_all || "View All Activity"}
        </Link>
      </div>
    </div>
  );
};
