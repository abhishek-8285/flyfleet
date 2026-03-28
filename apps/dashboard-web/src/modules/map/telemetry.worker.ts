/// <reference lib="webworker" />
/* eslint-disable no-restricted-globals */
export {}

interface VFeature {
    type: "Feature";
    properties: {
        id: string;
        speed: number;
        heading: number;
        status: string; // "moving", "idle"
        timestamp: number;
    };
    geometry: {
        type: "Point";
        coordinates: [number, number];
    };
}

let ws: WebSocket | null = null;

// The Engine's State Map
const vehicleState = new Map<string, VFeature>();

let isDirty = false;
let batchInterval: any = null;

function connect(orgId: string, token: string, wsUrlStr: string) {
    if (ws) ws.close();

    const wsUrl = `${wsUrlStr}/ws?org_id=${orgId}&token=${token || ""}`;
    ws = new WebSocket(wsUrl);

    ws.onopen = () => {
        postMessage({ type: "WS_STATUS", status: "CONNECTED" });
        
        // Start 100ms batcher loop
        if (!batchInterval) {
            batchInterval = setInterval(() => {
                if (isDirty) {
                    flushBatch();
                    isDirty = false;
                }
            }, 100);
        }
    };

    ws.onmessage = (event) => {
        try {
            const data = JSON.parse(event.data);
            if (!data || data.type !== "gps") return;

            const vId = data.vehicle_id;
            if (!vId) return;

            const lat = data.data?.latitude;
            const lng = data.data?.longitude;
            if (lat == null || lng == null) return;
            
            const speed = data.data?.speed || 0;
            const status = speed > 5 ? "moving" : "idle";

            // Mapbox Native Feature format
            const feature: VFeature = {
               type: "Feature",
               properties: {
                   id: vId,
                   speed: speed,
                   heading: data.data?.heading || 0,
                   status: status,
                   timestamp: new Date(data.timestamp || data.data?.time).getTime()
               },
               geometry: {
                   type: "Point",
                   coordinates: [lng, lat]
               }
            };

            vehicleState.set(vId, feature);
            isDirty = true;
        } catch (err) {
            // Drop invalid messages silently to save CPU
        }
    };

    ws.onclose = () => {
        postMessage({ type: "WS_STATUS", status: "DISCONNECTED" });
        if (batchInterval) {
            clearInterval(batchInterval);
            batchInterval = null;
        }
    };

    ws.onerror = () => {
        postMessage({ type: "WS_STATUS", status: "ERROR" });
    };
}

function flushBatch() {
    const features = Array.from(vehicleState.values());
    const geoJson = {
        type: "FeatureCollection",
        features
    };
    
    // Transfer bulk geoJSON over the worker thread boundary
    postMessage({ type: "GEOJSON_BATCH", payload: geoJson });
}

self.addEventListener("message", (e) => {
    const msg = e.data;
    if (msg.type === "CONNECT") {
        connect(msg.orgId, msg.token, msg.wsUrl);
    } else if (msg.type === "DISCONNECT") {
        if (ws) ws.close();
        if (batchInterval) clearInterval(batchInterval);
        batchInterval = null;
    }
});
