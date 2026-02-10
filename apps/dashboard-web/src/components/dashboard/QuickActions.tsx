import React from 'react';

import { useNavigate } from 'react-router-dom';

interface QuickActionsProps {
    onAddVehicle: () => void;
}

export const QuickActions: React.FC<QuickActionsProps> = ({ onAddVehicle }) => {
    const navigate = useNavigate();

    return (
        <div className="bg-surface-dark rounded-xl p-5 border border-border-color shadow-lg shadow-black/20">
            <h3 className="text-white font-semibold mb-4 flex items-center gap-2">
                <span className="material-symbols-outlined text-primary text-lg">bolt</span>
                Quick Actions
            </h3>
            <div className="grid grid-cols-2 gap-3">
                <button 
                    onClick={onAddVehicle}
                    className="flex flex-col items-center justify-center p-4 bg-surface-darker hover:bg-primary/20 hover:border-primary/50 text-white rounded-lg transition-all group text-center border border-border-color hover:shadow-lg active:scale-95"
                >
                    <span className="material-symbols-outlined mb-2 text-primary group-hover:scale-110 transition-transform">add_circle</span>
                    <span className="text-xs font-semibold">Add Vehicle</span>
                </button>
                <button 
                    onClick={() => navigate('/admin/users')}
                    className="flex flex-col items-center justify-center p-4 bg-surface-darker hover:bg-primary/20 hover:border-primary/50 text-white rounded-lg transition-all group text-center border border-border-color hover:shadow-lg active:scale-95"
                >
                    <span className="material-symbols-outlined mb-2 text-primary group-hover:scale-110 transition-transform">person_add</span>
                    <span className="text-xs font-semibold">User Access</span>
                </button>
                <button 
                    onClick={() => navigate('/admin/monitoring')}
                    className="flex flex-col items-center justify-center p-4 bg-surface-darker hover:bg-primary/20 hover:border-primary/50 text-white rounded-lg transition-all group text-center border border-border-color hover:shadow-lg active:scale-95"
                >
                    <span className="material-symbols-outlined mb-2 text-primary group-hover:scale-110 transition-transform">analytics</span>
                    <span className="text-xs font-semibold">System Health</span>
                </button>
                <button 
                    onClick={() => navigate('/support')}
                    className="flex flex-col items-center justify-center p-4 bg-surface-darker hover:bg-primary/20 hover:border-primary/50 text-white rounded-lg transition-all group text-center border border-border-color hover:shadow-lg active:scale-95"
                >
                    <span className="material-symbols-outlined mb-2 text-primary group-hover:scale-110 transition-transform">support_agent</span>
                    <span className="text-xs font-semibold">Get Support</span>
                </button>
            </div>
        </div>
    );
};
