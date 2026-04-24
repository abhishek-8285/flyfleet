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
  fields: ["id", "type", "tenant_id", "status", "gps_enabled"],
  storeFields: ["id", "type", "tenant_id", "status", "lastUpdate", "gps_enabled"],
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

// Global Error Handler for diagnostics
self.onerror = (message, source, lineno, colno, error) => {
  self.postMessage({
    type: "WORKER_CRASH_DIAGNOSTIC",
    payload: {
      message: typeof message === 'string' ? message : (message as any).message,
      lineno,
      colno,
      stack: error?.stack,
    },
  });
};

self.onunhandledrejection = (e) => {
  self.postMessage({
    type: "WORKER_CRASH_DIAGNOSTIC",
    payload: {
      message: `Unhandled Rejection: ${e.reason}`,
      stack: e.reason?.stack,
    },
  });
};

// Initialize IndexedDB for persistence
const DB_NAME = "FlyFleet_Operational_DB";
const DB_VERSION = 1;

function initDB() {
  const request = indexedDB.open(DB_NAME, DB_VERSION);
  request.onerror = (e) => {
    console.error(`[OperationalWorker] Failed to open IndexedDB:`, e);
    // Don't crash, just log. System can function in memory.
  };
  request.onblocked = () => {
    console.warn("[OperationalWorker] IndexedDB is blocked. Waiting...");
  };
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
    console.log("[OperationalWorker] IndexedDB Initialized");
  };
}

initDB();

self.onmessage = async (event: MessageEvent) => {
  const { type, payload } = event.data;

  if (type === "INDEX_VEHICLES") {
    const updates: SearchableVehicle[] = [];
    payload.forEach((v: any) => {
      const existing = vehicleMap.get(v.id) || {};
      const updated = { ...existing, ...v };
      vehicleMap.set(v.id, updated as SearchableVehicle);
      updates.push(updated as SearchableVehicle);

      // Update sharded index if tenant_id is present
      const tenant_id = updated.tenant_id || updated.tenantId || updated.org;
      if (tenant_id) {
        if (!shardedIndexes.has(tenant_id)) {
          shardedIndexes.set(
            tenant_id,
            new MiniSearch({
              fields: ["id", "type", "status", "gps_enabled"],
              storeFields: ["id", "type", "status", "lastUpdate", "gps_enabled"],
            }),
          );
        }
        const shard = shardedIndexes.get(tenant_id)!;
        if (shard.has(updated.id)) {
          shard.replace(updated);
        } else {
          shard.add(updated);
        }
      }
    });

    updates.forEach((u) => {
      if (miniSearch.has(u.id)) {
        miniSearch.replace(u);
      } else {
        miniSearch.add(u);
      }
    });
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

    // Shared filter logic
    if (filters) {
      searchOptions.filter = (result: any) => {
        return Object.entries(filters).every(([key, value]) => {
          if (!value || key === 'tenant_id' || key === 'org') return true; // tenant is handled by sharding or direct scan
          // Handle boolean strings from the UI (e.g. "true")
          if (key === "gps_enabled") {
            return String(result[key]) === String(value);
          }
          return result[key] === value;
        });
      };
    }

    let results: any[] = [];

    // Optimize for empty query + filters: scan the map directly instead of MiniSearch
    if (!query && filters) {
      results = Array.from(vehicleMap.values())
        .filter((v: any) => {
          // If tenant filter is present, check it first
          const filterTenant = filters.tenant_id || filters.org;
          const vehicleTenant = v.tenant_id || v.tenantId || v.org;
          if (filterTenant && vehicleTenant !== filterTenant) return false;
          // Apply other filters
          return searchOptions.filter(v);
        });
    } else if ((filters?.tenant_id || filters?.org) && shardedIndexes.has(filters?.tenant_id || filters?.org)) {
      // Use sharded index if tenant filter is active
      const activeTenant = filters.tenant_id || filters.org;
      results = shardedIndexes
        .get(activeTenant)!
        .search(query || "", searchOptions);
    } else {
      results = miniSearch.search(query || "", searchOptions);
    }

    self.postMessage({ type: "SEARCH_RESULTS", payload: results.slice(0, 20) });
  }

  if (type === "GET_FILTER_METADATA") {
    const tenants = new Set<string>();
    const statuses = new Set<string>();
    vehicleMap.forEach((v: any) => {
      const tid = v.tenant_id || v.tenantId || v.org;
      if (tid) tenants.add(tid);
      if (v.status) statuses.add(v.status);
    });
    self.postMessage({
      type: "FILTER_METADATA",
      payload: {
        tenants: Array.from(tenants),
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
    // console.log("[OperationalWorker] Received PING, responding with PONG");
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
