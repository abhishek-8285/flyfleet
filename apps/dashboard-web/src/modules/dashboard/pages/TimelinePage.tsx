import { format } from "date-fns";
import React from "react";
import { useSelector } from "react-redux";
import { RootState } from "../../../store";

const TimelinePage: React.FC = () => {
  const { events } = useSelector((state: RootState) => state.timeline);

  return (
    <div className="p-6 space-y-6 bg-black min-h-screen text-white">
      <h1 className="text-3xl font-bold">Activity Timeline</h1>

      <div className="relative border-l border-gray-800 ml-4 space-y-8 pb-10">
        {events.length === 0 ? (
          <div className="text-gray-500 pl-8">No historical events found.</div>
        ) : (
          events.map((event: any, idx) => (
            <div key={idx} className="relative pl-8">
              <div className="absolute -left-[5px] top-1.5 h-2 w-2 rounded-full bg-blue-500 ring-4 ring-black"></div>
              <div className="bg-gray-900 border border-gray-800 p-4 rounded-xl max-w-2xl">
                <div className="flex justify-between items-start mb-2">
                  <span className="text-xs font-semibold uppercase tracking-wider text-blue-400">
                    {event.type.replace("_", " ")}
                  </span>
                  <span className="text-xs text-gray-500">
                    {format(new Date(event.timestamp), "MMM d, HH:mm:ss")}
                  </span>
                </div>
                <p className="text-gray-200">{event.message}</p>
                <div className="mt-4 flex gap-4 text-xs text-gray-500">
                  <span>Actor: {event.actor || "System"}</span>
                  {event.id && <span>ID: {event.id}</span>}
                </div>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default TimelinePage;
