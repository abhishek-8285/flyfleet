import React, { useEffect } from "react";
import { useDispatch, useSelector } from "react-redux";
import { useUIStrings } from "../hooks/useUIStrings";
import { AppDispatch, RootState } from "../store";
import { fetchTrips } from "../store/tripsSlice";
import { Trip } from "../types";

const Trips: React.FC = () => {
  const dispatch = useDispatch<AppDispatch>();
  const { trips, loading } = useSelector((state: RootState) => state.trips);
  const { uiStrings } = useUIStrings("trips");

  useEffect(() => {
    dispatch(fetchTrips({}) as any);
  }, [dispatch]);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-lg text-text-muted">
          {uiStrings?.loading_text || "Loading trips..."}
        </div>
      </div>
    );
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-3xl font-bold text-white">
          {uiStrings?.title || "Trips"}
        </h1>
        <button className="px-4 py-2 bg-primary text-white rounded hover:bg-primary/80 transition-colors font-medium">
          {uiStrings?.btn_create || "Create Trip"}
        </button>
      </div>

      <div className="bg-surface-dark rounded-xl border border-border-color shadow-lg overflow-hidden">
        <table className="min-w-full divide-y divide-border-color">
          <thead className="bg-[#1d2936]">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-semibold text-text-muted uppercase tracking-wider">
                {uiStrings?.th_status || "Execution Status"}
              </th>
              <th className="px-6 py-3 text-left text-xs font-semibold text-text-muted uppercase tracking-wider">
                {uiStrings?.th_vehicle || "Tractor Unit"}
              </th>
              <th className="px-6 py-3 text-left text-xs font-semibold text-text-muted uppercase tracking-wider">
                {uiStrings?.th_driver || "Lead Driver"}
              </th>
              <th className="px-6 py-3 text-left text-xs font-semibold text-text-muted uppercase tracking-wider">
                {uiStrings?.th_start_time || "Ignition Event"}
              </th>
              <th className="px-6 py-3 text-left text-xs font-semibold text-text-muted uppercase tracking-wider">
                {uiStrings?.th_distance || "Net Distance"}
              </th>
              <th className="px-6 py-3 text-left text-xs font-semibold text-text-muted uppercase tracking-wider">
                {uiStrings?.th_actions || "Actions"}
              </th>
            </tr>
          </thead>
          <tbody className="bg-surface-dark divide-y divide-border-color">
            {trips.length === 0 ? (
              <tr>
                <td
                  colSpan={6}
                  className="px-6 py-10 text-center text-text-muted italic"
                >
                  {uiStrings?.no_data || "No active trips found."}
                </td>
              </tr>
            ) : (
              trips.map((trip: Trip) => (
                <tr
                  key={trip.id}
                  className="hover:bg-[#1d2936] transition-colors"
                >
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span
                      className={`px-2 py-1 text-xs font-semibold rounded-md border ${
                        trip.status === "ongoing"
                          ? "bg-blue-500/10 text-blue-400 border-blue-500/20"
                          : trip.status === "completed"
                            ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
                            : "bg-surface-darker text-text-muted border-border-color"
                      }`}
                    >
                      {trip.status}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-white font-medium">
                    {trip.vehicle_id.slice(-8)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-text-muted">
                    {trip.driver_id.slice(-8)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-text-muted text-xs">
                    {new Date(trip.start_time).toLocaleString()}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-white font-medium">
                    {trip.distance_km
                      ? `${trip.distance_km.toFixed(1)} km`
                      : "-"}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm">
                    <button className="text-primary hover:text-white transition-colors">
                      {uiStrings?.btn_details || "View Details"}
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default Trips;
