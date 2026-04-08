// types/tracking.ts

export interface VehiclePosition {
    vehicle_id: string;
    org_id: string;
    registration: string;
    make: string;
    model: string;
    year: number;
    vehicle_status: string;
    current_driver_id: string | null;
    driver_name: string | null;
    driver_phone: string | null;
    last_update: string;
    latitude: number;
    longitude: number;
    speed_kmh: number | null;
    heading: number | null;
    ignition: boolean | null;
    fuel_level: number | null;
    derived_status: VehicleStatus;
}

export interface TelemetryPoint {
    time: string;
    vehicle_id: string;
    latitude: number;
    longitude: number;
    speed_kmh: number | null;
    heading: number | null;
    altitude_m: number | null;
    ignition: boolean | null;
    fuel_level: number | null;
}

export interface Geofence {
    id: string;
    org_id: string;
    name: string;
    description: string;
    type: 'circle' | 'polygon' | 'corridor';
    center_lat: number | null;
    center_lng: number | null;
    radius_m: number | null;
    polygon: [number, number][] | null;
    status: 'active' | 'inactive';
    alert_channels: string[];
    created_at: string;
    updated_at: string;
}

export interface GeofenceEvent {
    id: string;
    geofence_id: string;
    geofence_name: string;
    vehicle_id: string;
    registration: string;
    event_type: 'entry' | 'exit';
    occurred_at: string;
    latitude: number;
    longitude: number;
    alert_sent: boolean;
}

export type VehicleStatus = 'moving' | 'idle' | 'stopped' | 'offline';
export type PlaybackSpeed = 1 | 2 | 4 | 8;

export interface WSMessage {
    type: 'position_update' | 'geofence_event' | 'error' | 'subscription_ack' | 'connection_ack';
    data: any;
    timestamp: string;
}

export interface HistoryRequest {
    vehicle_id: string;
    start: string;
    end: string;
    resolution?: number;
}

export interface HistorySummary {
    total_distance_km: number;
    avg_speed_kmh: number;
    max_speed_kmh: number;
    duration_minutes: number;
    stops_count: number;
}
