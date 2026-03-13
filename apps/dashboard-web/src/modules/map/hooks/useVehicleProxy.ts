import { useSelector } from "react-redux";
import { RootState } from "../../../store/index";

/**
 * Optimized hook for Zero-Churn telemetry access.
 * Reads high-frequency attributes directly from SharedArrayBuffer.
 */
export const useVehicleProxy = (
  vehicleId: string,
  renderBufferSab: SharedArrayBuffer | null,
) => {
  const index = useSelector(
    (state: RootState) => state.telemetry.indices[vehicleId],
  );
  const metadata = useSelector(
    (state: RootState) => state.telemetry.vehicles[vehicleId],
  );

  if (index === undefined || !renderBufferSab) {
    return metadata;
  }

  const buffer = new Float32Array(renderBufferSab);
  const stride = 8;
  const offset = index * stride;

  // We return a proxy or a fresh object that prods the buffer.
  // For standard React components, we might still want a fresh object on render
  // but it avoids Redux subscription overhead for every position change.
  return {
    ...metadata,
    // Latest values from binary buffer
    longitude: buffer[offset + 3], // curr_lng
    latitude: buffer[offset + 4], // curr_lat
    heading: buffer[offset + 5], // curr_hdg
  };
};
