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
      "icon-image": "vehicle-marker",
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

  // Add Identity Labels Layer
  map.addLayer({
    id: "vehicle-labels",
    type: "symbol",
    source: "vehicles",
    filter: ["!", ["has", "point_count"]],
    layout: {
      "text-field": ["get", "registration"],
      "text-font": ["DIN Offc Pro Medium", "Arial Unicode MS Bold"],
      "text-size": 10,
      "text-offset": [0, 1.5],
      "text-anchor": "top",
    },
    paint: {
      "text-color": "#ffffff",
      "text-halo-color": "#000000",
      "text-halo-width": 1,
    },
  });

  // Generate a local vehicle marker icon (avoids CORS/COEP issues)
  const generateVehicleMarker = () => {
    const size = 64;
    const canvas = document.createElement("canvas");
    canvas.width = size;
    canvas.height = size;
    const ctx = canvas.getContext("2d");
    if (!ctx) return null;

    // Draw Arrow / Marker
    ctx.shadowBlur = 4;
    ctx.shadowColor = "rgba(0,0,0,0.5)";
    ctx.fillStyle = "#3b82f6"; // Primary blue
    ctx.beginPath();
    ctx.moveTo(size / 2, 5);
    ctx.lineTo(size - 10, size - 10);
    ctx.lineTo(size / 2, size - 20);
    ctx.lineTo(10, size - 10);
    ctx.closePath();
    ctx.fill();

    // Add white outline
    ctx.strokeStyle = "#ffffff";
    ctx.lineWidth = 3;
    ctx.stroke();

    return ctx.getImageData(0, 0, size, size);
  };

  const imageData = generateVehicleMarker();
  if (imageData && !map.hasImage("vehicle-marker")) {
    map.addImage("vehicle-marker", imageData);
  }
};

export const updateVehicleSource = (map: maplibregl.Map, vehicles: any[]) => {
  const source = map.getSource("vehicles") as maplibregl.GeoJSONSource;
  if (!source) return;

  const geojson: GeoJSON.FeatureCollection = {
    type: "FeatureCollection",
    features: vehicles
      .filter((v) => v.latitude != null && v.longitude != null && v.gpsEnabled === true)
      .map((v) => ({
        type: "Feature",
        geometry: {
          type: "Point",
          coordinates: [v.longitude, v.latitude],
        },
      properties: {
        id: v.id,
        registration: v.registrationNumber || v.registration || `Asset ${String(v.id).substring(0, 8)}`,
        heading: v.heading || 0,
        speed: v.speed || 0,
        status: v.status || "active",
      },
    })),
  };

  source.setData(geojson);
};
