import { MapIcon } from "@heroicons/react/24/outline";
import { useUIStrings } from "../hooks/useUIStrings";

export default function Geofences() {
  const { uiStrings: rawUiStrings } = useUIStrings("geofences");
  const uiStrings = rawUiStrings || {};

  return (
    <div className="p-8 h-full flex flex-col items-center justify-center text-center max-w-2xl mx-auto">
      <div className="w-20 h-20 rounded-3xl bg-primary/10 flex items-center justify-center text-primary mb-6 animate-pulse">
        <MapIcon className="w-10 h-10" />
      </div>
      <h1 className="text-3xl font-bold text-white mb-4">
        {uiStrings.title || "Geofences"}
      </h1>
      <p className="text-text-muted text-lg mb-8">
        {uiStrings.sub ||
          "Geofence management with map drawing will be displayed here"}
      </p>

      <div className="bg-surface-dark/50 border border-border-color rounded-2xl p-6 w-full backdrop-blur-sm">
        <div className="flex items-center gap-3 text-sm text-text-muted">
          <div className="w-2 h-2 rounded-full bg-warning animate-ping" />
          {uiStrings.footnote || "Requires Leaflet Draw integration"}
        </div>
      </div>
    </div>
  );
}
