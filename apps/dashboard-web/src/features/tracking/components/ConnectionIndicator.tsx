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
        <div className="flex items-center gap-3 px-4 py-2 bg-surface-dim border border-outline rounded-2xl shadow-lg">
            <div className="relative">
                <span className={`w-2 h-2 rounded-full block ${color} ${status === 'connected' ? 'animate-pulse' : ''}`} />
                {status === 'connected' && (
                    <span className={`absolute inset-0 w-2 h-2 rounded-full ${color} animate-ping opacity-50`} />
                )}
            </div>
            <span className="text-[10px] font-black uppercase tracking-widest text-white leading-none">{label}</span>
            <span className="material-symbols-outlined text-on-surface-variant text-[12px] opacity-40">
                {status === 'connected' ? 'wifi' : status === 'polling' ? 'sync' : 'wifi_off'}
            </span>
        </div>
    );
};
