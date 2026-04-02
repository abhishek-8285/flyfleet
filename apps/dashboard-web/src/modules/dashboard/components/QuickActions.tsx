import { Activity, HelpCircle, PlusCircle, Users } from "lucide-react";
import React from "react";
import { useNavigate } from "react-router-dom";
import { useRBAC } from "../../../context/RBACProvider";

// Fallback component for missing MagneticButton import
const MagneticButton: React.FC<{ children: React.ReactNode; onClick?: () => void; variant?: "primary" | "secondary"; className?: string }> = ({ children, onClick, variant, className }) => {
  const variantClasses = variant === "primary" ? "bg-blue-600 hover:bg-blue-700" : "bg-gray-800 hover:bg-gray-700";
  return (
    <button onClick={onClick} className={`${variantClasses} text-white rounded-xl transition-colors ${className}`}>
      {children}
    </button>
  );
};

const QuickActions: React.FC = () => {
  const { can } = useRBAC();
  const navigate = useNavigate();

  const actions = [
    {
      label: "Add Vehicle",
      icon: PlusCircle,
      path: "/admin/vehicles/new",
      color: "primary" as const,
      action: () => console.log("Add Vehicle Modal"),
      visible: can("create", "vehicle"),
    },
    {
      label: "User Access",
      icon: Users,
      path: "/admin/users",
      color: "secondary" as const,
      visible: can("read", "user"),
    },
    {
      label: "System Health",
      icon: Activity,
      path: "/admin/monitoring",
      color: "secondary" as const,
      visible: can("read", "system"),
    },
    {
      label: "Support",
      icon: HelpCircle,
      path: "/support",
      color: "secondary" as const,
      visible: true,
    },
  ];

  const visibleActions = actions.filter((a) => a.visible);

  return (
    <div className="grid grid-cols-2 gap-4">
      {visibleActions.map((action, idx) => (
        <MagneticButton
          key={idx}
          onClick={() => {
            if (action.action) action.action();
            else navigate(action.path);
          }}
          variant={action.color}
          className="w-full flex flex-col items-center justify-center py-6"
        >
          <action.icon size={32} className="mb-3" />
          <span className="font-semibold">{action.label}</span>
        </MagneticButton>
      ))}
    </div>
  );
};

export default QuickActions;
