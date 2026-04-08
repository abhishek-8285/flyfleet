// components/ConnectionIndicator.tsx

import React from 'react';
import { useSelector } from 'react-redux';
import { RootState } from '../../../store';

export const ConnectionIndicator: React.FC = () => {
    const wsConnected = useSelector((state: RootState) => state.tracking.wsConnected);
    const fallbackToPolling = useSelector((state: RootState) => state.tracking.fallbackToPolling);
    const reconnectAttempts = useSelector((state: RootState) => state.tracking.wsReconnectAttempts);

    let status: 'connected' | 'reconnecting' | 'polling' | 'error';
    let color: string;
    let label: string;

    if (wsConnected) {
        status = 'connected';
        color = 'bg-green-500';
        label = 'Live';
    } else if (fallbackToPolling) {
        status = 'polling';
        color = 'bg-yellow-500';
        label = 'Polling (10s)';
    } else if (reconnectAttempts > 0) {
        status = 'reconnecting';
        color = 'bg-yellow-500';
        label = `Reconnecting (${reconnectAttempts})`;
    } else {
        status = 'error';
        color = 'bg-red-500';
        label = 'Disconnected';
    }

    return (
        <div className="flex items-center gap-2 px-3 py-1 bg-white rounded-full shadow-sm">
            <span className={`w-2 h-2 rounded-full ${color} ${status === 'connected' ? 'animate-pulse' : ''}`} />
            <span className="text-xs text-gray-600">{label}</span>
        </div>
    );
};
