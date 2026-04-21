import React from "react";
import { 
  User, 
  Car, 
  MoreVertical, 
  Edit3, 
  Trash2, 
  ChevronRight,
  ShieldCheck,
  IdCard,
  Briefcase
} from "lucide-react";
import { Driver } from "../../types";

type DriverRow = Driver & {
  current_vehicle?: {
    id?: string;
    registration_number?: string;
  };
};

interface DriverTableProps {
  drivers: DriverRow[];
  loading: boolean;
  onDriverClick?: (driver: Driver) => void;
  selectedIds?: Set<string>;
  onSelect?: (id: string) => void;
  onSelectAll?: () => void;
  onEdit?: (id: string) => void;
  onDelete?: (id: string) => void;
}

export const DriverTable: React.FC<DriverTableProps> = ({
  drivers,
  loading,
  onDriverClick,
  selectedIds = new Set(),
  onSelect,
  onSelectAll,
  onEdit,
  onDelete,
}) => {
  const allSelected = drivers.length > 0 && selectedIds.size === drivers.length;

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-left">
        <thead>
          <tr className="text-xs font-semibold text-gray-400 uppercase tracking-wider bg-white/5 border-b border-border-color">
            {onSelect && (
              <th className="px-6 py-4 w-12">
                <input
                  type="checkbox"
                  checked={allSelected}
                  onChange={() => onSelectAll?.()}
                  className="rounded border-white/10 bg-white/5 accent-primary cursor-pointer"
                />
              </th>
            )}
            <th className="px-6 py-4">Driver</th>
            <th className="px-6 py-4">License</th>
            <th className="px-6 py-4">Status</th>
            <th className="px-6 py-4">Assigned Vehicle</th>
            <th className="px-6 py-4 text-right">Actions</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-border-color">
          {loading ? (
            Array(3).fill(0).map((_, i) => (
              <tr key={i} className="animate-pulse">
                <td className="px-6 py-4" colSpan={6}><div className="h-4 bg-white/5 rounded w-full" /></td>
              </tr>
            ))
          ) : drivers.length === 0 ? (
            <tr>
              <td colSpan={6} className="px-6 py-10 text-center text-gray-500">No personnel registered.</td>
            </tr>
          ) : (
            drivers.map((driver) => (
              <tr 
                key={driver.id} 
                className="hover:bg-white/[0.02] transition-colors cursor-pointer group"
                onClick={() => onDriverClick?.(driver)}
              >
                {onSelect && (
                  <td className="px-6 py-4" onClick={(e) => e.stopPropagation()}>
                    <input
                      type="checkbox"
                      checked={selectedIds.has(driver.id)}
                      onChange={() => onSelect(driver.id)}
                      className="rounded border-white/10 bg-white/5 accent-primary cursor-pointer"
                    />
                  </td>
                )}
                <td className="px-6 py-4">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center text-primary group-hover:scale-110 transition-transform font-bold">
                      {driver.name?.[0] || driver.first_name?.[0]}
                    </div>
                    <div>
                      <div className="font-semibold text-white">{driver.name || `${driver.first_name} ${driver.last_name}`}</div>
                      <div className="text-xs text-gray-500 capitalize">{driver.employment_type?.toLowerCase() || 'Full Time'}</div>
                    </div>
                  </div>
                </td>
                <td className="px-6 py-4">
                  <div className="flex flex-col">
                    <span className="text-sm text-gray-300 font-mono">{driver.license_number || "-"}</span>
                    <span className="text-[10px] text-gray-500 uppercase">{driver.license_type || "LMV"}</span>
                  </div>
                </td>
                <td className="px-6 py-4">
                  <div className="flex items-center gap-2">
                    <div className={`w-1.5 h-1.5 rounded-full ${driver.status === 'active' ? 'bg-success' : 'bg-primary'}`} />
                    <span className={`text-xs font-medium uppercase ${driver.status === 'active' ? 'text-success' : 'text-primary'}`}>
                      {driver.status}
                    </span>
                  </div>
                </td>
                <td className="px-6 py-4">
                    {driver.current_vehicle?.registration_number ? (
                        <div className="flex items-center gap-2 text-white/70 text-sm">
                            <Car className="w-4 h-4 text-primary" />
                            {driver.current_vehicle.registration_number}
                        </div>
                    ) : (
                        <span className="text-gray-600 text-xs italic">Unassigned</span>
                    )}
                </td>
                <td className="px-6 py-4">
                  <div className="flex items-center justify-end gap-2">
                    <button 
                      onClick={(e) => { e.stopPropagation(); onEdit?.(driver.id); }}
                      className="p-2 text-gray-500 hover:text-white transition-colors"
                    >
                      <Edit3 className="w-4 h-4" />
                    </button>
                    <button 
                      onClick={(e) => { e.stopPropagation(); onDelete?.(driver.id); }}
                      className="p-2 text-gray-500 hover:text-error transition-colors"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                    <ChevronRight className="w-5 h-5 text-gray-600" />
                  </div>
                </td>
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  );
};
