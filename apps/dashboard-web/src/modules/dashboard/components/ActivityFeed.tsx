import { formatDistanceToNow } from "date-fns";
import { AlertCircle, CheckCircle, Info, PlayCircle } from "lucide-react";
import React from "react";
import { motion } from "framer-motion";
import { TimelineEvent } from "../../../store/timelineSlice";
import { StaggerContainer } from "../../../components/animated/StaggerContainer";

interface ActivityFeedProps {
  events: TimelineEvent[];
}

const ActivityFeed: React.FC<ActivityFeedProps> = ({ events }) => {
  const getIcon = (type: string) => {
    switch (type) {
      case "trip_started":
        return <PlayCircle className="text-blue-400" size={20} />;
      case "alert":
        return <AlertCircle className="text-red-400" size={20} />;
      case "delivery_complete":
        return <CheckCircle className="text-green-400" size={20} />;
      default:
        return <Info className="text-gray-400" size={20} />;
    }
  };

  return (
    <div className="bg-gray-900 border border-gray-800 rounded-xl overflow-hidden">
      <div className="p-4 border-b border-gray-800 flex justify-between items-center">
        <h2 className="text-lg font-bold text-white">Activity Feed</h2>
        <div className="flex items-center gap-1.5">
          <span className="w-2 h-2 rounded-full bg-blue-500 animate-pulse"></span>
          <span className="text-xs text-gray-400 font-medium">
            Live Updates
          </span>
        </div>
      </div>
      <div className="h-[400px] overflow-y-auto p-4 space-y-4">
        {events.length === 0 ? (
          <div className="text-center text-gray-500 py-20">
            No recent activity
          </div>
        ) : (
          <StaggerContainer staggerDelay={0.08} maxDelay={0.4}>
            {events.slice(0, 20).map((event) => (
              <motion.div
                key={event.id}
                whileHover={{ x: 4, backgroundColor: "rgba(255, 255, 255, 0.03)" }}
                transition={{ duration: 0.2 }}
                className="flex gap-4 p-2 rounded-lg cursor-pointer"
              >
                <div className="mt-1">{getIcon(event.type)}</div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm text-gray-200 leading-snug">
                    <span className="font-semibold text-white">
                      {event.actor || "System"}
                    </span>{" "}
                    {event.message}
                  </p>
                  <p className="text-xs text-gray-500 mt-1">
                    {formatDistanceToNow(new Date(event.timestamp), {
                      addSuffix: true,
                    })}
                  </p>
                </div>
              </motion.div>
            ))}
          </StaggerContainer>
        )}
      </div>
      <div className="p-3 border-t border-gray-800 bg-gray-950 text-center">
        <button className="text-sm font-medium text-blue-400 hover:text-blue-300">
          View All Activity
        </button>
      </div>
    </div>
  );
};

export default ActivityFeed;
