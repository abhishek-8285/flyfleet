import maplibregl from "maplibre-gl";

export const addRouteLayers = (map: maplibregl.Map) => {
  // Source for planned routes
  map.addSource("planned-route", {
    type: "geojson",
    data: {
      type: "FeatureCollection",
      features: [],
    },
  });

  // Planned Route Line
  map.addLayer({
    id: "planned-route-line",
    type: "line",
    source: "planned-route",
    layout: {
      "line-join": "round",
      "line-cap": "round",
    },
    paint: {
      "line-color": "#2563eb",
      "line-width": 4,
      "line-opacity": 0.8,
    },
  });

  // Source for breadcrumbs
  map.addSource("breadcrumbs", {
    type: "geojson",
    data: {
      type: "FeatureCollection",
      features: [],
    },
  });

  // Breadcrumb Trail
  map.addLayer({
    id: "breadcrumb-trail",
    type: "line",
    source: "breadcrumbs",
    layout: {
      "line-join": "round",
      "line-cap": "round",
    },
    paint: {
      "line-color": "#10b981",
      "line-width": 2,
      "line-dasharray": [2, 2],
      "line-opacity": 0.6,
    },
  });
};

export const updatePlannedRoute = (
  map: maplibregl.Map,
  coordinates: [number, number][],
) => {
  const source = map.getSource("planned-route") as maplibregl.GeoJSONSource;
  if (!source) return;

  const geojson: GeoJSON.FeatureCollection = {
    type: "FeatureCollection",
    features: [
      {
        type: "Feature",
        geometry: {
          type: "LineString",
          coordinates,
        },
        properties: {},
      },
    ],
  };

  source.setData(geojson);
};

export const updateBreadcrumbs = (
  map: maplibregl.Map,
  coordinates: [number, number][],
) => {
  const source = map.getSource("breadcrumbs") as maplibregl.GeoJSONSource;
  if (!source) return;

  const geojson: GeoJSON.FeatureCollection = {
    type: "FeatureCollection",
    features: [
      {
        type: "Feature",
        geometry: {
          type: "LineString",
          coordinates,
        },
        properties: {},
      },
    ],
  };

  source.setData(geojson);
};

export const clearRoutes = (map: maplibregl.Map) => {
  const plannedSource = map.getSource(
    "planned-route",
  ) as maplibregl.GeoJSONSource;
  const breadcrumbSource = map.getSource(
    "breadcrumbs",
  ) as maplibregl.GeoJSONSource;

  if (plannedSource) {
    plannedSource.setData({ type: "FeatureCollection", features: [] });
  }
  if (breadcrumbSource) {
    breadcrumbSource.setData({ type: "FeatureCollection", features: [] });
  }
};
