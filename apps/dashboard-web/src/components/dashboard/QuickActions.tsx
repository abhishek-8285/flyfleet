import React from 'react';

interface QuickActionsProps {
    onAddVehicle: () => void;
}

export const QuickActions: React.FC<QuickActionsProps> = ({ onAddVehicle }) => {
    return (
        <div className="bg-surface-dark rounded-xl p-5 border border-border-color">
            <h3 className="text-white font-semibold mb-4">Quick Actions</h3>
            <div className="grid grid-cols-2 gap-3">
                <button 
                    onClick={onAddVehicle}
                    className="flex flex-col items-center justify-center p-4 bg-surface-darker hover:bg-primary hover:text-white rounded-lg transition-all group text-center border border-transparent hover:shadow-lg hover:shadow-primary/20"
                >
                    <span className="material-symbols-outlined mb-2 text-primary group-hover:text-white">add_circle</span>
                    <span className="text-sm font-medium text-white">Add Vehicle</span>
                </button>
                <button className="flex flex-col items-center justify-center p-4 bg-surface-darker hover:bg-primary hover:text-white rounded-lg transition-all group text-center border border-transparent hover:shadow-lg hover:shadow-primary/20">
                    <span className="material-symbols-outlined mb-2 text-primary group-hover:text-white">person_add</span>
                    <span className="text-sm font-medium text-white">Assign Driver</span>
                </button>
                <button className="flex flex-col items-center justify-center p-4 bg-surface-darker hover:bg-primary hover:text-white rounded-lg transition-all group text-center border border-transparent hover:shadow-lg hover:shadow-primary/20">
                    <span className="material-symbols-outlined mb-2 text-primary group-hover:text-white">assignment</span>
                    <span className="text-sm font-medium text-white">Log Report</span>
                </button>
                <button className="flex flex-col items-center justify-center p-4 bg-surface-darker hover:bg-primary hover:text-white rounded-lg transition-all group text-center border border-transparent hover:shadow-lg hover:shadow-primary/20">
                    <span className="material-symbols-outlined mb-2 text-primary group-hover:text-white">schedule</span>
                    <span className="text-sm font-medium text-white">Schedule</span>
                </button>
            </div>
        </div>
    );
};
