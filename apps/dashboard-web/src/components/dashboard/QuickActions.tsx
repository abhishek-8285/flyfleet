import React from 'react';
import { useNavigate } from 'react-router-dom';
import MagneticButton from '../animated/MagneticButton';

interface QuickActionsProps {
    onAddVehicle: () => void;
}

export const QuickActions: React.FC<QuickActionsProps> = ({ onAddVehicle }) => {
    const navigate = useNavigate();

    const actions = [
        {
            label: "Add Vehicle",
            icon: "add_circle",
            onClick: onAddVehicle,
        },
        {
            label: "User Access",
            icon: "person_add",
            onClick: () => navigate('/admin/users'),
        },
        {
            label: "System Health",
            icon: "analytics",
            onClick: () => navigate('/admin/monitoring'),
        },
        {
            label: "Get Support",
            icon: "support_agent",
            onClick: () => navigate('/support'),
        },
    ];

    return (
        <div className="bg-surface-dark rounded-xl p-5 border border-border-color shadow-lg shadow-black/20">
            <h3 className="text-white font-semibold mb-4 flex items-center gap-2">
                <span className="material-symbols-outlined text-primary text-lg">bolt</span>
                Quick Actions
            </h3>
            <div className="grid grid-cols-2 gap-3">
                {actions.map((action, idx) => (
                    <MagneticButton
                        key={idx}
                        onClick={action.onClick}
                        variant="secondary"
                        className="w-full flex flex-col items-center justify-center py-4"
                    >
                        <span className="material-symbols-outlined mb-2 text-primary">{action.icon}</span>
                        <span className="text-xs font-semibold">{action.label}</span>
                    </MagneticButton>
                ))}
            </div>
        </div>
    );
};
