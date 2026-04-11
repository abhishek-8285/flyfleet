import { cellToBoundary } from "h3-js";
import { Map } from "maplibre-gl";

/**
 * H3 Cluster Layer: Renders global vehicle density aggregates.
 * Used for LOD 0-1 (Zoom levels 1-10).
 */
export class ClusterLayer {
  private id = "h3-clusters";
  private sourceId = "h3-density-source";
  private map: Map | null = null;

  onAdd(map: Map) {
    this.map = map;

    // Initialize GeoJSON source for H3 cells
    map.addSource(this.sourceId, {
      type: "geojson",
      data: { type: "FeatureCollection", features: [] },
    });

    // Add polygon layer for H3 cells
    map.addLayer({
      id: `${this.id}-polygons`,
      type: "fill",
      source: this.sourceId,
      maxzoom: 12,
      paint: {
        "fill-color": [
          "interpolate",
          ["linear"],
          ["get", "count"],
          0,
          "rgba(30, 41, 59, 0)", // Transparent Slate
          1,
          "rgba(59, 130, 246, 0.2)", // Blue 500
          10,
          "rgba(139, 92, 246, 0.5)", // Violet 500
          100,
          "rgba(249, 115, 22, 0.7)", // Orange 500
          1000,
          "rgba(239, 68, 68, 0.9)", // Red 500
        ],
        "fill-outline-color": "rgba(255, 255, 255, 0.05)",
      },
    });

    // Add text count layer
    map.addLayer({
      id: `${this.id}-counts`,
      type: "symbol",
      source: this.sourceId,
      minzoom: 4,
      maxzoom: 12,
      layout: {
        "text-field": ["get", "count"],
        "text-size": 12,
        "text-font": ["Open Sans Bold", "Arial Unicode MS Bold"],
      },
      paint: {
        "text-color": "#ffffff",
        "text-opacity": ["interpolate", ["linear"], ["zoom"], 4, 0, 6, 1],
      },
    });
  }

  updateData(h3Counts: [string, number][]) {
    if (!this.map || (this.map as any)._destroyed) return;

    const features = h3Counts.map(([cell, count]) => {
      const boundary = cellToBoundary(cell, true); // [[lng, lat], ...]
      return {
        type: "Feature",
        properties: { count },
        geometry: {
          type: "Polygon",
          coordinates: [boundary],
        },
      };
    });

    const source = this.map.getSource(this.sourceId) as any;
    if (source) {
      source.setData({
        type: "FeatureCollection",
        features,
      });
    }
  }

  onRemove() {
    if (this.map) {
      this.map.removeLayer(`${this.id}-polygons`);
      this.map.removeLayer(`${this.id}-counts`);
      this.map.removeSource(this.sourceId);
    }
  }
}
