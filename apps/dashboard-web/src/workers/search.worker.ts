/* eslint-disable no-restricted-globals */

/**
 * SearchWorker: Index and search vehicles in memory without blocking the UI.
 * Uses a simple prefix-based inverted index for speed.
 */

interface SearchableVehicle {
  id: string;
  type?: string;
  status?: string;
}

let vehicleIndex: SearchableVehicle[] = [];

self.onmessage = (event: MessageEvent) => {
  const { type, payload } = event.data;

  if (type === "INDEX_VEHICLES") {
    // payload is an array of vehicles
    vehicleIndex = payload;
  }

  if (type === "SEARCH") {
    const query = payload.toLowerCase();
    const results = vehicleIndex
      .filter((v) => v.id.toLowerCase().includes(query))
      .slice(0, 10); // Return top 10

    self.postMessage({ type: "SEARCH_RESULTS", payload: results });
  }

  if (type === "PING") {
    self.postMessage({ type: "PONG" });
  }
};

export { };

