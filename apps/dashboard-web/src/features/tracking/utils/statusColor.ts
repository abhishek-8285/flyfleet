// utils/statusColor.ts

import { VehicleStatus } from '../types/tracking';

const statusColorMap: Record<string, string> = {
    active: '#22c55e',   // Green
    moving: '#22c55e',   // Green
    idle: '#f59e0b',     // Amber
    stopped: '#ef4444',  // Red
    stale: '#6b7280',    // Gray
    offline: '#6b7280',  // Gray
};

export function statusColor(status: VehicleStatus): string {
    return statusColorMap[status] || statusColorMap.offline;
}

export function statusLabel(status: VehicleStatus): string {
    if (!status) return 'Offline';
    return String(status).charAt(0).toUpperCase() + String(status).slice(1).toLowerCase();
}
