import { ExclamationTriangleIcon, XMarkIcon } from '@heroicons/react/24/solid';
import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

interface SOSAlert {
    id: string;
    vehicle_reg_number: string;
    driver_name: string;
    timestamp: string;
    severity: 'LOW' | 'MEDIUM' | 'CRITICAL';
}

export const SOSOverlay: React.FC = () => {
    const [activeAlert, setActiveAlert] = useState<SOSAlert | null>(null);
    const navigate = useNavigate();

    useEffect(() => {
        // Mock listener for WebSocket events
        const simulateIncident = () => {
            setTimeout(() => {
                setActiveAlert({
                    id: 'new-incident-123',
                    vehicle_reg_number: 'KA-01-HH-1234',
                    driver_name: 'Rajesh Kumar',
                    timestamp: new Date().toISOString(),
                    severity: 'CRITICAL'
                });

                // Play sound
                const audio = new Audio('/sounds/emergency-siren.mp3');
                audio.play().catch(e => console.log('Audio play failed', e));

            }, 5000); // Simulate an alert after 5 seconds
        };

        // simulateIncident(); // Uncomment to test

        // Cleanup
        return () => { };
    }, []);

    if (!activeAlert) return null;

    return (
        <div className="fixed inset-0 z-[9999] flex items-end justify-center pointer-events-none pb-12 px-4">
            <div className="bg-red-600 text-white rounded-lg shadow-2xl max-w-lg w-full pointer-events-auto ring-4 ring-red-500 ring-opacity-50 animate-bounce-subtle">
                <div className="p-4 flex items-start gap-4">
                    <div className="bg-white/20 p-2 rounded-full animate-pulse">
                        <ExclamationTriangleIcon className="h-8 w-8 text-white" />
                    </div>
                    <div className="flex-1">
                        <h3 className="text-lg font-bold">EMERGENCY SOS TRIGGERED</h3>
                        <p className="text-red-100 mt-1">
                            Vehicle <strong>{activeAlert.vehicle_reg_number}</strong> ({activeAlert.driver_name}) reported a
                            <span className="font-bold bg-white/20 px-1 mx-1 rounded">{activeAlert.severity}</span>
                            incident.
                        </p>
                        <div className="mt-4 flex gap-3">
                            <button
                                onClick={() => {
                                    navigate('/admin/emergency');
                                    setActiveAlert(null);
                                }}
                                className="bg-white text-red-700 px-4 py-2 rounded font-bold hover:bg-gray-100 transition flex-1"
                            >
                                View Details
                            </button>
                            <button
                                onClick={() => setActiveAlert(null)}
                                className="border border-white/30 text-white px-4 py-2 rounded hover:bg-white/10 transition"
                            >
                                Dismiss
                            </button>
                        </div>
                    </div>
                    <button onClick={() => setActiveAlert(null)} className="text-white/60 hover:text-white">
                        <XMarkIcon className="h-6 w-6" />
                    </button>
                </div>
            </div>

            {/* Red overlay flash effect */}
            <div className="fixed inset-0 bg-red-500/10 z-[-1] animate-pulse pointer-events-none"></div>
        </div>
    );
};
