// hooks/useVehiclePositions.ts

import { useSelector } from 'react-redux';
import { RootState } from '../../../store';
import { VehiclePosition } from '../types/tracking';

export function useVehiclePositions(): VehiclePosition[] {
    const vehicles = useSelector((state: RootState) => state.tracking.vehicles);
    const statusFilters = useSelector((state: RootState) => state.tracking.statusFilters);
    const searchQuery = useSelector((state: RootState) => state.tracking.searchQuery);

    let positions = Object.values(vehicles);

    // Apply status filter
    if (statusFilters.length > 0) {
        positions = positions.filter(v => statusFilters.includes(v.derived_status));
    }

    // Apply search filter
    if (searchQuery) {
        const query = searchQuery.toLowerCase();
        positions = positions.filter(v =>
            v.registration.toLowerCase().includes(query) ||
            v.make.toLowerCase().includes(query) ||
            v.model.toLowerCase().includes(query)
        );
    }

    return positions;
}
