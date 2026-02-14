import React from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { VehicleMapCard } from '../components/vehicles/VehicleMapCard';
import { VehicleSpecsPanel } from '../components/vehicles/VehicleSpecsPanel';
import { VehicleStatsGrid } from '../components/vehicles/VehicleStatsGrid';
import { api } from '../services/api';
import { Vehicle } from '../types';

const VehicleDetails: React.FC = () => {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const [vehicle, setVehicle] = React.useState<Vehicle | null>(null);
    const [loading, setLoading] = React.useState(true);

    React.useEffect(() => {
        const fetchVehicle = async () => {
            if (!id) return;
            setLoading(true);
            try {
                const data = await api.getVehicle(id);
                setVehicle(data);
            } catch (error) {
                console.error("Failed to fetch vehicle", error);
            } finally {
                setLoading(false);
            }
        };
        fetchVehicle();
    }, [id]);

    if (loading) {
        return (
            <div className="flex-1 flex items-center justify-center bg-background-dark">
                <div className="text-primary animate-spin text-4xl">
                    <span className="material-symbols-outlined">sync</span>
                </div>
            </div>
        );
    }

    if (!vehicle) {
        return (
            <div className="flex-1 flex flex-col items-center justify-center bg-background-dark text-text-muted">
                <span className="material-symbols-outlined text-6xl mb-4">error_outline</span>
                <p>Vehicle not found or failed to load.</p>
                <button onClick={() => navigate(-1)} className="mt-4 text-primary hover:underline">Go Back</button>
            </div>
        );
    }

    return (
        <div className="flex-1 flex flex-col min-w-0 bg-background-dark overflow-hidden h-full">
            <header className="h-16 flex items-center justify-between px-6 border-b border-border-color bg-surface-darker flex-shrink-0 z-20">
                <div className="flex items-center gap-4">
                    <button onClick={() => navigate(-1)} className="text-text-muted hover:text-white">
                        <span className="material-symbols-outlined">arrow_back</span>
                    </button>
                    <div>
                        <div className="flex items-center gap-3">
                            <h2 className="text-white text-lg font-bold">Vehicle #{vehicle.registration_number}</h2>
                            <span className={`px-2 py-0.5 rounded border text-xs font-medium ${vehicle.status === 'active' ? 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20' :
                                vehicle.status === 'maintenance' ? 'bg-yellow-500/10 text-yellow-400 border-yellow-500/20' :
                                    'bg-text-muted/10 text-text-muted border-border-color'
                                }`}>
                                {vehicle.status.charAt(0).toUpperCase() + vehicle.status.slice(1)}
                            </span>
                            <span className="bg-primary/10 text-primary text-xs px-2 py-0.5 rounded border border-primary/20 font-medium flex items-center gap-1">
                                <span className="material-symbols-outlined text-[12px] animate-pulse">leak_add</span>
                                Live Telemetry
                            </span>
                        </div>
                        <p className="text-[11px] text-text-muted mt-0.5">{vehicle.year} {vehicle.make} {vehicle.model} • VIN: {vehicle.metadata?.vin ?? 'N/A'}</p>
                    </div>
                </div>
                <div className="flex items-center gap-3">
                    <button className="px-3 py-1.5 bg-surface-dark border border-border-color hover:bg-border-color text-text-muted hover:text-white text-xs font-medium rounded flex items-center gap-2 transition-colors">
                        <span className="material-symbols-outlined text-[16px]">edit</span>
                        Edit Vehicle
                    </button>
                    <button className="px-3 py-1.5 bg-red-500/10 border border-red-500/20 hover:bg-red-500/20 text-red-500 text-xs font-medium rounded flex items-center gap-2 transition-colors">
                        <span className="material-symbols-outlined text-[16px]">delete</span>
                        Decommission
                    </button>
                </div>
            </header>
            <div className="flex-1 overflow-y-auto p-6 custom-scrollbar">
                <div className="grid grid-cols-12 gap-6 h-full min-h-[600px]">
                    <div className="col-span-12 lg:col-span-5 flex flex-col gap-6 h-full">
                        <VehicleMapCard />
                    </div>
                    <div className="col-span-12 lg:col-span-7 flex flex-col h-full gap-6">
                        <VehicleStatsGrid vehicle={vehicle} />
                        <VehicleSpecsPanel vehicle={vehicle} />
                    </div>
                </div>
            </div>
            <footer className="bg-surface-darker border-t border-border-color px-6 py-2 flex items-center justify-between text-[10px] text-text-muted z-20 flex-shrink-0">
                <span>© 2026 FlyFleet Global Operations. All rights reserved.</span>
                <div className="flex items-center gap-4">
                    <span>Data Source: Telematics Gateway V2</span>
                    <span>Server Time: <span className="font-mono text-white">{new Date().toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', hour12: false, timeZone: 'UTC' })} UTC</span></span>
                </div>
            </footer>
        </div>
    );
};

export default VehicleDetails;
