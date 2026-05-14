import React, { useState, useMemo } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import { 
  Search, Tag, Layers, Eye, 
  MapPin, User, Video, ListFilter, Target, ChevronDown
} from 'lucide-react';
import { RootState } from '../../../store';
import { setSelectedVehicle } from '../../../store/telemetrySlice';
import { api } from '../../../services/api';

/**
 * AssetDispatchList: High-precision dispatch sidebar mirroring the reference design.
 */
export const AssetDispatchList: React.FC = () => {
  const dispatch = useDispatch();
  const telemetryVehicles = useSelector((state: RootState) => state.telemetry.vehicles);
  const selectedId = useSelector((state: RootState) => state.telemetry.selectedVehicleId);
  const [registryVehicles, setRegistryVehicles] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');

  React.useEffect(() => {
    const fetchRegistry = async () => {
      try {
        const res = await api.getVehicles({ limit: 100 });
        const data = res?.data || res;
        setRegistryVehicles(Array.isArray(data) ? data : data?.vehicles || []);
      } catch (err) {
        console.error('Failed to fetch registry vehicles:', err);
      } finally {
        setLoading(false);
      }
    };
    fetchRegistry();
  }, []);

  const combinedAssets = useMemo(() => {
    return registryVehicles.map(v => {
      const live = telemetryVehicles[v.id] || {};
      return {
        ...v,
        ...live,
        id: v.id,
        registrationNumber: v.registration_number || live.registrationNumber,
        driver_name: v.current_driver?.name || v.current_driver_name || live.driver_name,
      };
    });
  }, [registryVehicles, telemetryVehicles]);

  const filteredAssets = combinedAssets.filter(v => 
    v.registrationNumber?.toLowerCase().includes(searchQuery.toLowerCase()) ||
    v.id.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="flex flex-col h-full bg-white dark:bg-[#0b1219] font-sans overflow-hidden border-r border-gray-200 dark:border-white/5 transition-colors duration-300">
      {/* Search & Action Header */}
      <div className="p-4 bg-white dark:bg-transparent space-y-4">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input 
            type="text"
            placeholder={`Search ${registryVehicles.length} assets`}
            className="w-full bg-[#f6f7f9] dark:bg-white/5 border-none rounded-lg pl-10 pr-4 py-2.5 text-sm text-gray-800 dark:text-white placeholder:text-gray-400 focus:ring-0 transition-all font-medium"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>

        <div className="flex items-center justify-between text-gray-700 dark:text-gray-400">
          <div className="flex gap-5">
            <button className="flex items-center gap-1.5 text-[13px] font-bold hover:text-black dark:hover:text-white">
              <Tag className="w-5 h-5 fill-black dark:fill-white text-black dark:text-white" /> Tags <ChevronDown className="w-4 h-4 text-gray-400" />
            </button>
            <button className="flex items-center gap-1.5 text-[13px] font-bold hover:text-black dark:hover:text-white">
              Attributes <ChevronDown className="w-4 h-4 text-gray-400" />
            </button>
            <button className="flex items-center gap-1.5 text-[13px] font-bold hover:text-black dark:hover:text-white">
              More <ChevronDown className="w-4 h-4 text-gray-400" />
            </button>
          </div>
          <Eye className="w-5 h-5 text-gray-800 dark:text-gray-200 cursor-pointer hover:text-primary transition-colors" />
        </div>

        <button className="flex items-center gap-1 text-[13px] font-bold text-gray-700 dark:text-gray-400 hover:text-black dark:hover:text-white">
          Sort <ChevronDown className="w-4 h-4 text-gray-400" />
        </button>
      </div>

      {/* Asset List */}
      <div className="flex-1 overflow-y-auto custom-scrollbar border-t border-gray-100 dark:border-white/5">
        {loading ? (
          <div className="p-8 text-center text-gray-400 animate-pulse text-xs font-bold uppercase tracking-widest">Initialising Fleet...</div>
        ) : filteredAssets.length === 0 ? (
          <div className="p-12 text-center text-gray-400 text-sm font-medium">No assets found</div>
        ) : (
          <div className="divide-y divide-gray-100 dark:divide-white/[0.03]">
            {filteredAssets.map((asset) => {
              const isSelected = selectedId === asset.id;
              const speed = Math.round(asset.speed ?? 0);
              const isMoving = speed > 2;
              const hasTelemetry = !!asset.latitude;
              
              return (
                <div 
                  key={asset.id}
                  onClick={() => dispatch(setSelectedVehicle(asset.id))}
                  className={`group px-6 py-5 cursor-pointer transition-all relative ${
                    isSelected ? 'bg-blue-50/20 dark:bg-primary/[0.07]' : 'hover:bg-gray-50/50 dark:hover:bg-white/[0.02]'
                  }`}
                >
                  <div className="flex justify-between items-start mb-2.5">
                    <div className="flex items-center gap-2">
                      <span className="text-[16px] font-black text-[#1a1a1a] dark:text-white tracking-tight">
                        {asset.registrationNumber || asset.id.substring(0, 8)}
                      </span>
                      <Video className="w-4.5 h-4.5 text-gray-400 fill-gray-400/20" />
                    </div>
                    
                    {hasTelemetry && (
                      <div className={`px-2.5 py-0.5 rounded-md text-[11px] font-black border flex items-center gap-1.5 shadow-sm transition-all ${
                        isMoving 
                          ? 'bg-emerald-50 dark:bg-emerald-500/10 text-emerald-600 dark:text-emerald-500 border-emerald-100 dark:border-emerald-500/20 shadow-emerald-500/5' 
                          : 'bg-emerald-50/50 dark:bg-white/5 text-emerald-500/80 dark:text-gray-500 border-emerald-100/50 dark:border-white/5'
                      }`}>
                        {speed} MPH
                      </div>
                    )}
                  </div>

                  <div className="space-y-2.5">
                    <div className="flex items-start gap-2.5">
                      {isMoving ? (
                        <MapPin className="w-4 h-4 text-emerald-500 mt-0.5 flex-shrink-0" />
                      ) : (
                        <Target className="w-4 h-4 text-blue-500 mt-0.5 flex-shrink-0" />
                      )}
                      <span className="text-[12px] font-bold text-gray-500 dark:text-gray-400 leading-tight">
                        {asset.latitude && asset.longitude 
                          ? `${asset.latitude.toFixed(5)}, ${asset.longitude.toFixed(5)} (Active)`
                          : 'Awaiting Position...'}
                      </span>
                    </div>
                    
                    <div className="flex items-center gap-2.5">
                      <User className="w-4 h-4 text-gray-400 flex-shrink-0" />
                      <span className="text-[12px] font-bold text-gray-400 dark:text-gray-500">
                        {asset.driver_name || 'Pilot Unassigned'}
                      </span>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* Pagination Bar */}
      <div className="p-5 bg-white dark:bg-[#0b1219] border-t border-gray-50 dark:border-white/5 flex items-center justify-center gap-5 transition-colors">
        <div className="w-8 h-8 rounded-full bg-[#5c6b7a] dark:bg-primary flex items-center justify-center text-white text-[11px] font-black shadow-lg">1</div>
        {[2, 3, 4, 5, 6, 7].map(p => (
          <button key={p} className="text-gray-400 dark:text-gray-600 text-[12px] font-bold hover:text-black dark:hover:text-white transition-colors">{p}</button>
        ))}
        <button className="text-gray-400 dark:text-gray-600 text-[12px] font-black hover:text-black dark:hover:text-white uppercase tracking-widest ml-2">Next</button>
      </div>
    </div>
  );
};

export default AssetDispatchList;
