// services/websocketService.ts

import { WSMessage, VehiclePosition, GeofenceEvent } from '../types/tracking';
import { store } from '../../../store';
import {
    updateVehiclePosition,
    setWsConnected,
    setWsReconnectAttempts,
    setError,
    setFallbackToPolling,
    addGeofenceEvent,
} from '../store/trackingSlice';

class TelemetryWebSocket {
    private ws: WebSocket | null = null;
    private reconnectAttempts = 0;
    private readonly MAX_RECONNECT = 5;
    private readonly BASE_DELAY = 1000;
    private token: string = '';
    private orgId: string = '';

    connect(token: string, orgId: string) {
        this.token = token;
        this.orgId = orgId;
        this.reconnectAttempts = 0;

        const wsUrl = `${this.getWsBaseUrl()}/ws/telemetry?token=${token}&org_id=${orgId}`;
        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = this.handleOpen.bind(this);
        this.ws.onmessage = this.handleMessage.bind(this);
        this.ws.onclose = this.handleClose.bind(this);
        this.ws.onerror = this.handleError.bind(this);
    }

    private handleOpen() {
        this.reconnectAttempts = 0;
        store.dispatch(setWsConnected(true));
        store.dispatch(setWsReconnectAttempts(0));
        store.dispatch(setFallbackToPolling(false));
    }

    private handleMessage(event: MessageEvent) {
        try {
            const message: WSMessage = JSON.parse(event.data);

            switch (message.type) {
                case 'position_update':
                    store.dispatch(updateVehiclePosition(message.data as VehiclePosition));
                    break;
                case 'geofence_event':
                    store.dispatch(addGeofenceEvent(message.data as GeofenceEvent));
                    break;
                case 'connection_ack':
                    break;
                case 'error':
                    store.dispatch(setError(message.data as string));
                    break;
            }
        } catch (err) {
            console.error('Failed to parse WS message:', err);
        }
    }

    private handleClose() {
        store.dispatch(setWsConnected(false));
        this.scheduleReconnect();
    }

    private handleError(error: Event) {
        console.error('WebSocket error:', error);
    }

    private scheduleReconnect() {
        if (this.reconnectAttempts >= this.MAX_RECONNECT) {
            store.dispatch(setError('Connection lost. Refresh page.'));
            store.dispatch(setFallbackToPolling(true));
            this.startPolling();
            return;
        }

        const delay = this.BASE_DELAY * Math.pow(2, this.reconnectAttempts);
        this.reconnectAttempts++;
        store.dispatch(setWsReconnectAttempts(this.reconnectAttempts));

        setTimeout(() => {
            this.connect(this.token, this.orgId);
        }, delay);
    }

    private startPolling() {
        const poll = async () => {
            try {
                const response = await fetch('/api/v1/vehicles/locations', {
                    headers: {
                        'Authorization': `Bearer ${this.token}`,
                        'X-Org-ID': this.orgId
                    }
                });
                const data = await response.json();
                // Dispatch positions update
            } catch (error) {
                console.error('Polling failed:', error);
            }
            setTimeout(poll, 10000);
        };
        poll();
    }

    private getWsBaseUrl(): string {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        return `${protocol}//${window.location.host}`;
    }

    disconnect() {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
    }
}

export const telemetryWS = new TelemetryWebSocket();
