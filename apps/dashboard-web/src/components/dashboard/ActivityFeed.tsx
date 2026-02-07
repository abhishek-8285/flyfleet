import React from 'react';

export const ActivityFeed: React.FC = () => {
    return (
        <div className="bg-surface-dark rounded-xl border border-border-color flex-1 min-h-[400px] flex flex-col">
            <div className="p-4 border-b border-border-color bg-surface-darker flex justify-between items-center">
                <h3 className="text-white font-semibold">Recent Activity</h3>
                <button className="text-xs text-primary hover:text-white">Filter</button>
            </div>
            <div className="p-4 space-y-6 flex-1 overflow-y-auto custom-scrollbar max-h-[500px]">
                {/* Activity Item 1 */}
                <div className="flex gap-4 relative">
                    <div className="flex flex-col items-center">
                        <div className="w-8 h-8 rounded-full bg-primary/20 flex items-center justify-center text-primary z-10 ring-4 ring-surface-dark">
                            <span className="material-symbols-outlined text-[16px]">play_arrow</span>
                        </div>
                        <div className="w-0.5 bg-border-color h-full absolute top-8"></div>
                    </div>
                    <div className="pb-2">
                        <p className="text-sm text-white font-medium">Trip Started</p>
                        <p className="text-xs text-text-muted mt-0.5">Vehicle #402 left Depot C headed to Downtown.</p>
                        <span className="text-[10px] text-[#64748b] mt-1 block">2 mins ago</span>
                    </div>
                </div>

                {/* Activity Item 2 */}
                <div className="flex gap-4 relative">
                    <div className="flex flex-col items-center">
                        <div className="w-8 h-8 rounded-full bg-red-500/20 flex items-center justify-center text-red-500 z-10 ring-4 ring-surface-dark">
                            <span className="material-symbols-outlined text-[16px]">speed</span>
                        </div>
                        <div className="w-0.5 bg-border-color h-full absolute top-8"></div>
                    </div>
                    <div className="pb-2">
                        <p className="text-sm text-white font-medium">Speed Alert</p>
                        <p className="text-xs text-text-muted mt-0.5">Driver John D. exceeded 75mph on Route 66.</p>
                        <span className="text-[10px] text-[#64748b] mt-1 block">15 mins ago</span>
                    </div>
                </div>

                {/* Activity Item 3 */}
                <div className="flex gap-4 relative">
                    <div className="flex flex-col items-center">
                        <div className="w-8 h-8 rounded-full bg-emerald-500/20 flex items-center justify-center text-emerald-500 z-10 ring-4 ring-surface-dark">
                            <span className="material-symbols-outlined text-[16px]">check_circle</span>
                        </div>
                        <div className="w-0.5 bg-border-color h-full absolute top-8"></div>
                    </div>
                    <div className="pb-2">
                        <p className="text-sm text-white font-medium">Delivery Complete</p>
                        <p className="text-xs text-text-muted mt-0.5">Vehicle #112 arrived at destination.</p>
                        <span className="text-[10px] text-[#64748b] mt-1 block">42 mins ago</span>
                    </div>
                </div>

                {/* Activity Item 4 */}
                <div className="flex gap-4 relative">
                    <div className="flex flex-col items-center">
                        <div className="w-8 h-8 rounded-full bg-blue-400/20 flex items-center justify-center text-blue-400 z-10 ring-4 ring-surface-dark">
                            <span className="material-symbols-outlined text-[16px]">local_gas_station</span>
                        </div>
                    </div>
                    <div className="pb-2">
                        <p className="text-sm text-white font-medium">Refueling</p>
                        <p className="text-xs text-text-muted mt-0.5">Vehicle #892 refueling at Shell Station.</p>
                        <span className="text-[10px] text-[#64748b] mt-1 block">1h ago</span>
                    </div>
                </div>
            </div>
            <div className="px-4 py-3 bg-surface-darker border-t border-border-color text-center">
                <a className="text-xs text-primary hover:text-white" href="#">View All Activity</a>
            </div>
        </div>
    );
};
