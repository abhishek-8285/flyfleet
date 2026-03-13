import React from "react";
import { FleetList } from "../../../modules/fleet/components/FleetList";

/**
 * VehicleList: Entry point for the Fleet Command registry.
 * Now wraps the high-performance FleetList module which uses worker-based search.
 */
export const VehicleList: React.FC = () => {
  return (
    <div className="h-full">
      <FleetList />
    </div>
  );
};
