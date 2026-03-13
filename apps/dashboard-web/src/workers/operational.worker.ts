/* eslint-disable no-restricted-globals */

import MiniSearch from "minisearch";

/**
 * OperationalWorker: Handles search indexing and persistent telemetry storage.
 */

interface SearchableVehicle {
  id: string;
  type?: string;
  status?: string;
  lastUpdate: string;
}

const miniSearch = new MiniSearch({
  fields: ["id", "type", "org", "status"],
  storeFields: ["id", "type", "org", "status", "lastUpdate"],
  searchOptions: {
    boost: { id: 2 },
    fuzzy: 0.2,
    prefix: true,
  },
});

// Sharded indexes for specific organizations to improve performance at scale
const shardedIndexes = new Map<string, MiniSearch>();

let db: IDBDatabase | null = null;
const vehicleMap = new Map<string, SearchableVehicle>();

// Initialize IndexedDB for persistence
const DB_NAME = "FlyFleet_Operational_DB";
const DB_VERSION = 1;

function initDB() {
  const request = indexedDB.open(DB_NAME, DB_VERSION);
  request.onupgradeneeded = (e: any) => {
    const db = e.target.result;
    if (!db.objectStoreNames.contains("telemetry_history")) {
      const store = db.createObjectStore("telemetry_history", {
        keyPath: "id_timestamp",
      });
      store.createIndex("timestamp", "ts", { unique: false });
    }
  };
  request.onsuccess = (e: any) => {
    db = e.target.result;
    console.log("Operational DB Initialized");
  };
}

initDB();

self.onmessage = async (event: MessageEvent) => {
  const { type, payload } = event.data;

  if (type === "INDEX_VEHICLES") {
    const updates: SearchableVehicle[] = [];
    payload.forEach((v: any) => {
      vehicleMap.set(v.id, v);
      updates.push(v);

      // Update sharded index if org is present
      if (v.org) {
        if (!shardedIndexes.has(v.org)) {
          shardedIndexes.set(
            v.org,
            new MiniSearch({
              fields: ["id", "type", "status"],
              storeFields: ["id", "type", "status", "lastUpdate"],
            }),
          );
        }
        shardedIndexes.get(v.org)!.add(v);
      }
    });

    miniSearch.addAll(updates);
    persistBatch(payload);
  }

  if (type === "SEARCH") {
    const { query, filters } = payload;
    if (!query && !filters) {
      self.postMessage({ type: "SEARCH_RESULTS", payload: [] });
      return;
    }

    const searchOptions: any = {
      boost: { id: 2 },
      fuzzy: 0.2,
      prefix: true,
    };

    let results: any[] = [];

    // Use sharded index if org filter is active
    if (filters?.org && shardedIndexes.has(filters.org)) {
      results = shardedIndexes
        .get(filters.org)!
        .search(query || "", searchOptions);
    } else {
      if (filters) {
        searchOptions.filter = (result: any) => {
          return Object.entries(filters).every(([key, value]) => {
            if (!value) return true;
            return result[key] === value;
          });
        };
      }
      results = miniSearch.search(query || "", searchOptions);
    }

    self.postMessage({ type: "SEARCH_RESULTS", payload: results.slice(0, 20) });
  }

  if (type === "GET_FILTER_METADATA") {
    const orgs = new Set<string>();
    const statuses = new Set<string>();
    vehicleMap.forEach((v: any) => {
      if (v.org) orgs.add(v.org);
      if (v.status) statuses.add(v.status);
    });
    self.postMessage({
      type: "FILTER_METADATA",
      payload: {
        orgs: Array.from(orgs),
        statuses: Array.from(statuses),
      },
    });
  }

  if (type === "GET_REPLAY_DATA") {
    const { start, end } = payload;
    if (!db) {
      self.postMessage({ type: "REPLAY_DATA", payload: [] });
      return;
    }

    const tx = db.transaction("telemetry_history", "readonly");
    const store = tx.objectStore("telemetry_history");
    const index = store.index("timestamp");
    const results: any[] = [];

    // O(log n) range query using the timestamp index
    const range = IDBKeyRange.bound(start, end);
    const request = index.openCursor(range);

    request.onsuccess = (e: any) => {
      const cursor = e.target.result;
      if (cursor) {
        results.push(cursor.value);
        cursor.continue();
      } else {
        self.postMessage({ type: "REPLAY_DATA", payload: results });
      }
    };
  }

  if (type === "PING") {
    self.postMessage({ type: "PONG" });
  }
};

async function persistBatch(vehicles: SearchableVehicle[]) {
  if (!db) return;

  const tx = db.transaction("telemetry_history", "readwrite");
  const store = tx.objectStore("telemetry_history");

  vehicles.forEach((v) => {
    const entry = {
      ...v,
      ts: new Date(v.lastUpdate).getTime(),
      id_timestamp: `${v.id}_${v.lastUpdate}`,
    };
    store.put(entry);
  });

  tx.oncomplete = () => {
    // Optionally notify main thread of persistence cycle completion
  };
}
