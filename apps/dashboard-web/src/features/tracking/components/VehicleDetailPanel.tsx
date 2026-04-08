// components/VehicleDetailPanel.tsx

import React from 'react';
import { useSelector, useDispatch } from 'react-redux';
import { RootState } from '../../../store';
import { setSelectedVehicle } from '../store/trackingSlice';
import { statusColor, statusLabel } from '../utils/statusColor';

export const VehicleDetailPanel: React.FC = () => {
    const dispatch = useDispatch();
    const selectedVehicleId = useSelector((state: RootState) => state.tracking.selectedVehicleId);
    const vehicles = useSelector((state: RootState) => state.tracking.vehicles);

    if (!selectedVehicleId || !vehicles[selectedVehicleId]) {
        return null;
    }

    const vehicle = vehicles[selectedVehicleId];
    const color = statusColor(vehicle.derived_status);

    return (
        <div className="fixed right-0 top-0 h-full w-96 bg-white shadow-xl z-10 overflow-y-auto">
            <div className="p-6">
                <div className="flex justify-between items-center mb-6">
                    <h2 className="text-xl font-bold">{vehicle.registration}</h2>
                    <button
                        onClick={() => dispatch(setSelectedVehicle(null))}
                        className="text-gray-500 hover:text-gray-700"
                    >
                        ✕
                    </button>
                </div>

                <div className="space-y-4">
                    {/* Status Badge */}
                    <div className="flex items-center gap-2">
                        <span
                            className="w-3 h-3 rounded-full"
                            style={{ backgroundColor: color }}
                        />
                        <span className="font-medium">{statusLabel(vehicle.derived_status)}</span>
                    </div>

                    {/* Vehicle Info */}
                    <div className="bg-gray-50 p-4 rounded-lg">
                        <h3 className="font-semibold text-gray-700 mb-2">Vehicle Details</h3>
                        <div className="grid grid-cols-2 gap-2 text-sm">
                            <div>
                                <span className="text-gray-500">Make</span>
                                <p className="font-medium">{vehicle.make}</p>
                            </div>
                            <div>
                                <span className="text-gray-500">Model</span>
                                <p className="font-medium">{vehicle.model}</p>
                            </div>
                            <div>
                                <span className="text-gray-500">Year</span>
                                <p className="font-medium">{vehicle.year}</p>
                            </div>
                            <div>
                                <span className="text-gray-500">Last Update</span>
                                <p className="font-medium">
                                    {new Date(vehicle.last_update).toLocaleTimeString()}
                                </p>
                            </div>
                        </div>
                    </div>

                    {/* Location */}
                    <div className="bg-gray-50 p-4 rounded-lg">
                        <h3 className="font-semibold text-gray-700 mb-2">Current Location</h3>
                        <div className="text-sm space-y-1">
                            <p><span className="text-gray-500">Lat:</span> {vehicle.latitude.toFixed(6)}</p>
                            <p><span className="text-gray-500">Lng:</span> {vehicle.longitude.toFixed(6)}</p>
                            {vehicle.speed_kmh && (
                                <p><span className="text-gray-500">Speed:</span> {vehicle.speed_kmh.toFixed(1)} km/h</p>
                            )}
                            {vehicle.heading && (
                                <p><span className="text-gray-500">Heading:</span> {vehicle.heading.toFixed(0)}°</p>
                            )}
                        </div>
                    </div>

                    {/* Driver Info */}
                    {vehicle.driver_name && (
                        <div className="bg-gray-50 p-4 rounded-lg">
                            <h3 className="font-semibold text-gray-700 mb-2">Driver</h3>
                            <p className="font-medium">{vehicle.driver_name}</p>
                            {vehicle.driver_phone && (
                                <p className="text-sm text-gray-500">{vehicle.driver_phone}</p>
                            )}
                        </div>
                    )}

                    {/* Fuel */}
                    {vehicle.fuel_level !== null && (
                        <div className="bg-gray-50 p-4 rounded-lg">
                            <h3 className="font-semibold text-gray-700 mb-2">Fuel Level</h3>
                            <div className="w-full bg-gray-200 rounded-full h-2.5">
                                <div
                                    className="bg-blue-600 h-2.5 rounded-full"
                                    style={{ width: `${vehicle.fuel_level}%` }}
                                />
                            </div>
                            <p className="text-sm text-gray-500 mt-1">{vehicle.fuel_level.toFixed(1)}%</p>
                        </div>
                    )}

                    {/* Action Buttons */}
                    <div className="flex gap-2">
                        <button className="flex-1 bg-blue-600 text-white py-2 px-4 rounded hover:bg-blue-700">
                            View History
                        </button>
                        <button className="flex-1 bg-gray-200 text-gray-700 py-2 px-4 rounded hover:bg-gray-300">
                            Contact Driver
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
};
