import { Activity, HelpCircle, PlusCircle, Users } from "lucide-react";
import React from "react";
import { Link } from "react-router-dom";
import { useRBAC } from "../../../context/RBACProvider";

const QuickActions: React.FC = () => {
  const { can } = useRBAC();

  const actions = [
    {
      label: "Add Vehicle",
      icon: PlusCircle,
      path: "#",
      color: "bg-blue-600",
      action: () => console.log("Add Vehicle Modal"),
      visible: can("create", "vehicle"),
    },
    {
      label: "User Access",
      icon: Users,
      path: "/admin/users",
      color: "bg-purple-600",
      visible: can("read", "user"),
    },
    {
      label: "System Health",
      icon: Activity,
      path: "/admin/monitoring",
      color: "bg-green-600",
      visible: can("read", "system"),
    },
    {
      label: "Support",
      icon: HelpCircle,
      path: "/support",
      color: "bg-orange-600",
      visible: true,
    },
  ];

  const visibleActions = actions.filter((a) => a.visible);

  return (
    <div className="grid grid-cols-2 gap-4">
      {visibleActions.map((action, idx) => (
        <Link
          key={idx}
          to={action.path}
          onClick={action.action}
          className={`${action.color} flex flex-col items-center justify-center p-6 rounded-xl shadow-lg hover:opacity-90 transition-all text-white`}
        >
          <action.icon size={32} className="mb-3" />
          <span className="font-semibold">{action.label}</span>
        </Link>
      ))}
    </div>
  );
};

export default QuickActions;
