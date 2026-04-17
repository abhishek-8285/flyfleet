// components/FilterBar.tsx

import React from 'react';
import { useSelector, useDispatch } from 'react-redux';
import { RootState } from '../../../store';
import { setStatusFilters } from '../store/trackingSlice';
import { VehicleStatus } from '../types/tracking';
import { statusColor } from '../utils/statusColor';

const statuses: VehicleStatus[] = ['moving', 'idle', 'stopped', 'offline'];

export const FilterBar: React.FC = () => {
    const dispatch = useDispatch();
    const statusFilters = useSelector((state: RootState) => state.tracking.statusFilters);

    const toggleStatus = (status: VehicleStatus) => {
        const newFilters = statusFilters.includes(status)
            ? statusFilters.filter(s => s !== status)
            : [...statusFilters, status];
        dispatch(setStatusFilters(newFilters));
    };

    return (
        <div className="flex gap-4 flex-wrap items-center">
            {statuses.map(status => {
                const isActive = statusFilters.length === 0 || statusFilters.includes(status);
                const color = statusColor(status);

                return (
                    <button
                        key={status}
                        onClick={() => toggleStatus(status)}
                        className={`group flex items-center gap-3 px-5 py-2.5 rounded-2xl text-[10px] font-black uppercase tracking-[0.2em] transition-all border ${
                            isActive
                                ? 'bg-surface-dim border-outline text-white shadow-lg'
                                : 'bg-transparent border-transparent text-on-surface-variant opacity-40 hover:opacity-100 hover:bg-surface-container-highest'
                        }`}
                    >
                        <div className="relative flex items-center justify-center">
                            <span
                                className="w-2 h-2 rounded-full z-10"
                                style={{ backgroundColor: color }}
                            />
                            {isActive && (
                                <span 
                                    className="absolute w-4 h-4 rounded-full animate-ping opacity-20"
                                    style={{ backgroundColor: color }}
                                />
                            )}
                        </div>
                        <span className="group-hover:translate-x-0.5 transition-transform">{status}</span>
                    </button>
                );
            })}
        </div>
    );
};
