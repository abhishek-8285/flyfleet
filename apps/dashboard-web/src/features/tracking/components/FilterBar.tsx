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
        <div className="flex gap-2 flex-wrap">
            {statuses.map(status => {
                const isActive = statusFilters.length === 0 || statusFilters.includes(status);
                const color = statusColor(status);

                return (
                    <button
                        key={status}
                        onClick={() => toggleStatus(status)}
                        className={`flex items-center gap-1 px-3 py-1 rounded-full text-sm font-medium transition-all ${
                            isActive
                                ? 'bg-white shadow border border-gray-200'
                                : 'bg-gray-100 opacity-50'
                        }`}
                    >
                        <span
                            className="w-2 h-2 rounded-full"
                            style={{ backgroundColor: color }}
                        />
                        <span className="capitalize">{status}</span>
                    </button>
                );
            })}
        </div>
    );
};
