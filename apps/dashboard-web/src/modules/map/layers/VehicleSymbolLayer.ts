import maplibregl from "maplibre-gl";

export const addVehicleLayers = (map: maplibregl.Map) => {
  // Add source for vehicles
  map.addSource("vehicles", {
    type: "geojson",
    data: {
      type: "FeatureCollection",
      features: [],
    },
    cluster: true,
    clusterMaxZoom: 14,
    clusterRadius: 50,
  });

  // Add Cluster Layer
  map.addLayer({
    id: "clusters",
    type: "circle",
    source: "vehicles",
    filter: ["has", "point_count"],
    paint: {
      "circle-color": [
        "step",
        ["get", "point_count"],
        "#51bbd6",
        100,
        "#f1f075",
        750,
        "#f28cb1",
      ],
      "circle-radius": ["step", ["get", "point_count"], 20, 100, 30, 750, 40],
      "circle-stroke-width": 2,
      "circle-stroke-color": "#fff",
    },
  });

  // Add Cluster Count Layer
  map.addLayer({
    id: "cluster-count",
    type: "symbol",
    source: "vehicles",
    filter: ["has", "point_count"],
    layout: {
      "text-field": "{point_count_abbreviated}",
      "text-font": ["DIN Offc Pro Medium", "Arial Unicode MS Bold"],
      "text-size": 12,
    },
  });

  // Add Unclustered Vehicle Layer
  map.addLayer({
    id: "unclustered-point",
    type: "symbol",
    source: "vehicles",
    filter: ["!", ["has", "point_count"]],
    layout: {
      "icon-image": "vehicle-marker", // Needs to be loaded separately
      "icon-size": 0.6,
      "icon-rotate": ["get", "heading"],
      "icon-allow-overlap": true,
      "icon-ignore-placement": true,
    },
    paint: {
      "icon-opacity": [
        "case",
        ["boolean", ["feature-state", "hover"], false],
        1,
        0.8,
      ],
    },
  });

  // Load active vehicle icon
  map
    .loadImage("https://img.icons8.com/m_outlined/512/FFFFFF/long-arrow-up.png")
    .then((response: any) => {
      const image = response.data;
      if (image && !map.hasImage("vehicle-marker")) {
        map.addImage("vehicle-marker", image);
      }
    })
    .catch((error: any) => {
      console.error("Failed to load vehicle marker icon", error);
    });
};

export const updateVehicleSource = (map: maplibregl.Map, vehicles: any[]) => {
  const source = map.getSource("vehicles") as maplibregl.GeoJSONSource;
  if (!source) return;

  const geojson: GeoJSON.FeatureCollection = {
    type: "FeatureCollection",
    features: vehicles.map((v) => ({
      type: "Feature",
      geometry: {
        type: "Point",
        coordinates: [v.longitude, v.latitude],
      },
      properties: {
        id: v.id,
        heading: v.heading || 0,
        speed: v.speed || 0,
        status: v.status || "active",
      },
    })),
  };

  source.setData(geojson);
};
