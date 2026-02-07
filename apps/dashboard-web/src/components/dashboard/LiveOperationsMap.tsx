import React from 'react';
import { Link } from 'react-router-dom';
import LiveMap from '../../pages/LiveMap'; // Importing the existing LiveMap component
import { Trip } from '../../types';

interface LiveOperationsMapProps {
    trips: Trip[];
}

export const LiveOperationsMap: React.FC<LiveOperationsMapProps> = ({ trips }) => {
    return (
        <div className="bg-surface-dark rounded-xl border border-border-color shadow-sm flex flex-col h-[500px] overflow-hidden relative group/map">
            <div className="p-4 border-b border-border-color flex justify-between items-center bg-surface-darker z-10 relative">
                <h3 className="text-white font-semibold flex items-center gap-2">
                    <span className="material-symbols-outlined text-primary">public</span>
                    Live Fleet Map
                </h3>
                <div className="flex items-center gap-3">
                    <div className="flex items-center gap-2 mr-2">
                        <label className="text-xs text-text-muted font-medium cursor-pointer">Heatmap</label>
                        <button role="switch" aria-checked="false" className="w-8 h-4 rounded-full bg-border-color relative transition-colors focus:outline-none hover:bg-primary/50">
                            <span className="absolute left-0.5 top-0.5 w-3 h-3 bg-white rounded-full transition-transform"></span>
                        </button>
                    </div>
                    <div className="h-4 w-px bg-border-color"></div>
                    <div className="flex gap-1 bg-border-color/50 p-1 rounded">
                        <button className="px-3 py-1 text-xs font-medium text-white bg-surface-dark rounded shadow-sm">Map</button>
                        <button className="px-3 py-1 text-xs font-medium text-text-muted hover:text-white transition-colors">Satellite</button>
                    </div>
                </div>
            </div>
            
            <div className="relative flex-1 bg-[#0f151b] overflow-hidden">
                <div className="absolute inset-0 bg-cover bg-center opacity-70" style={{ backgroundImage: 'url("https://lh3.googleusercontent.com/aida-public/AB6AXuCeeeLxaUXFSywPWOUz7sSShAB3uQmjmm05fXrxnJMrtOj-77MaaA_rwghU8aFYgdEN25a4Zu_KRt4Bww_5vIegatJal2ZEZwc-jWi6kSXdQy3FKD1_7k4VpFPnvvH8JhhczyN8RmE4dQe7Igd7R_Y3Q2A8cUV3JoEubwD_1yrEB0VImkaOpvLuj7MtoMf4v4Vfp7EmLAR1LIeGUzUqVgEPm__oaIqEyfwohYrEJheP1zpnGNTy0WkwKZgcqRa3hJZyETmgEvoDzR4")' }}></div>
                <svg className="absolute inset-0 w-full h-full pointer-events-none opacity-40 z-0" viewBox="0 0 100 100" preserveAspectRatio="none">
                    <polygon points="20,20 40,20 45,40 25,45" fill="rgba(19, 127, 236, 0.2)" stroke="#137fec" strokeWidth="0.2" strokeDasharray="1" />
                </svg>
                <div className="absolute top-[25%] left-[30%] text-[10px] text-primary font-bold uppercase tracking-wider opacity-60 z-0">Geofence: Zone A</div>

                 {/* Real Map Component (Hidden in this pure visual mockup if strictly following the HTML background image approach, but let's keep it underlying or fallback) */}
                <div className="absolute inset-0 z-0 opacity-0">
                    <LiveMap trips={trips} height="100%" /> 
                </div>

                {/* Overlays */}
                <div className="absolute top-4 left-4 z-20 flex flex-col gap-3">
                    <div className="relative shadow-lg">
                        <input type="text" className="bg-surface-darker border border-border-color text-white text-xs rounded pl-8 pr-3 py-2 w-48 focus:ring-1 focus:ring-primary shadow-sm placeholder-text-muted/70" placeholder="Search map area..." />
                        <span className="material-symbols-outlined absolute left-2 top-1.5 text-base text-text-muted">search</span>
                    </div>
                    <div className="bg-surface-darker/90 backdrop-blur-sm border border-border-color rounded p-2 flex flex-col gap-2 w-32 shadow-lg">
                        <div className="flex items-center justify-between">
                             <span className="text-[10px] text-text-muted font-medium">Geofences</span>
                             <button className="w-6 h-3 rounded-full bg-primary relative transition-colors focus:outline-none">
                                <span className="absolute right-0.5 top-0.5 w-2 h-2 bg-white rounded-full transition-transform"></span>
                             </button>
                        </div>
                        <div className="flex items-center justify-between">
                             <span className="text-[10px] text-text-muted font-medium">Traffic</span>
                             <button className="w-6 h-3 rounded-full bg-border-color relative transition-colors focus:outline-none">
                                <span className="absolute left-0.5 top-0.5 w-2 h-2 bg-white rounded-full transition-transform"></span>
                             </button>
                        </div>
                    </div>
                </div>

                <div className="absolute top-4 right-4 bg-surface-darker/90 backdrop-blur-sm border border-border-color rounded p-3 z-20 shadow-lg">
                    <div className="text-[10px] font-semibold text-text-muted uppercase mb-2">Legend</div>
                    <div className="space-y-1.5">
                        <div className="flex items-center gap-2">
                            <span className="w-2 h-2 rounded-full bg-emerald-500"></span>
                            <span className="text-xs text-white">Active</span>
                        </div>
                        <div className="flex items-center gap-2">
                            <span className="w-2 h-2 rounded-full bg-primary"></span>
                            <span className="text-xs text-white">Moving</span>
                        </div>
                        <div className="flex items-center gap-2">
                            <span className="w-2 h-2 rounded-full bg-yellow-500"></span>
                            <span className="text-xs text-white">Idling</span>
                        </div>
                        <div className="flex items-center gap-2">
                            <span className="w-2 h-2 rounded-full bg-red-500"></span>
                            <span className="text-xs text-white">Alert</span>
                        </div>
                    </div>
                </div>

                <div className="absolute top-1/4 left-1/4 transform -translate-x-1/2 -translate-y-1/2 group cursor-pointer z-10">
                    <div className="w-8 h-8 bg-primary/20 rounded-full flex items-center justify-center animate-pulse">
                        <div className="w-3 h-3 bg-primary rounded-full shadow-[0_0_10px_#137fec]"></div>
                    </div>
                    <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 block bg-surface-darker text-white p-3 rounded-lg shadow-xl border border-border-color z-30 w-60">
                        <div className="flex justify-between items-start mb-2 border-b border-border-color pb-2">
                            <div>
                                <h4 className="font-bold text-sm">Vehicle #104</h4>
                                <span className="text-[10px] text-emerald-400 font-medium">● Online • 45 mph</span>
                            </div>
                        </div>
                        <div className="flex gap-2">
                            <Link to="/dashboard/vehicles/104" className="flex-1 bg-primary hover:bg-primary/90 text-white text-xs font-medium py-1.5 rounded transition-colors text-center">Profile</Link>
                        </div>
                        <div className="absolute bottom-[-6px] left-1/2 -translate-x-1/2 w-3 h-3 bg-surface-darker border-b border-r border-border-color transform rotate-45"></div>
                    </div>
                </div>

                <div className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 group cursor-pointer z-10">
                    <div className="w-8 h-8 bg-emerald-500/20 rounded-full flex items-center justify-center">
                        <div className="w-3 h-3 bg-emerald-500 rounded-full shadow-[0_0_10px_#10b981]"></div>
                    </div>
                </div>

                 <div className="absolute bottom-1/3 right-1/4 transform -translate-x-1/2 -translate-y-1/2 group cursor-pointer z-10">
                    <div className="w-8 h-8 bg-red-500/20 rounded-full flex items-center justify-center animate-bounce">
                        <div className="w-3 h-3 bg-red-500 rounded-full shadow-[0_0_10px_#ef4444]"></div>
                    </div>
                     <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 hidden group-hover:block bg-surface-darker text-xs text-white p-2 rounded shadow-lg whitespace-nowrap border border-border-color z-20">
                        <div className="font-bold text-red-400">Vehicle #301</div>
                        <div className="text-text-muted">Engine Temp Alert</div>
                        <Link to="/dashboard/vehicles/301" className="text-primary hover:underline mt-1 block">Drill down</Link>
                    </div>
                </div>

                <div className="absolute bottom-4 right-4 flex flex-col gap-2 z-20">
                     {/* Custom Zoom Controls */}
                    <button className="w-8 h-8 bg-surface-dark border border-border-color text-white rounded flex items-center justify-center hover:bg-border-color shadow-lg">
                        <span className="material-symbols-outlined text-sm">add</span>
                    </button>
                    <button className="w-8 h-8 bg-surface-dark border border-border-color text-white rounded flex items-center justify-center hover:bg-border-color shadow-lg">
                        <span className="material-symbols-outlined text-sm">remove</span>
                    </button>
                </div>
                 <div className="absolute bottom-1 left-1 px-2 py-1 bg-surface-darker/80 text-[10px] text-text-muted rounded backdrop-blur">
                     © OpenMapTiles | Source: GIS Data
                 </div>
            </div>
        </div>
    );
};
