import React from 'react';

const Reports: React.FC = () => {
    const reportCategories = [
        { title: "Fuel Consumption", icon: "local_gas_station", desc: "Monthly fuel efficiency and costs analysis." },
        { title: "Safety Scores", icon: "verified_user", desc: "Driver behavior and safety compliance metrics." },
        { title: "Maintenance ROI", icon: "request_quote", desc: "Total cost of ownership and repair analysis." },
        { title: "Utilization Report", icon: "bar_chart", desc: "Vehicle uptime and idle time distribution." },
        { title: "Route Efficiency", icon: "alt_route", desc: "Planned vs actual distance and time logs." },
        { title: "Carbon Footprint", icon: "eco", desc: "Fleet-wide emissions and sustainability goals." }
    ];

    return (
        <div className="flex-1 overflow-y-auto p-6 custom-scrollbar bg-background-dark">
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 max-w-[1600px] mx-auto">
                {reportCategories.map((report, idx) => (
                    <div 
                        key={idx} 
                        className="bg-surface-dark p-6 rounded-xl border border-border-color hover:border-primary/50 transition-all group cursor-pointer shadow-lg hover:shadow-primary/5 shadow-black/20"
                    >
                        <div className="flex items-center gap-4 mb-4">
                            <div className="w-12 h-12 rounded-lg bg-primary/10 flex items-center justify-center text-primary group-hover:bg-primary group-hover:text-white transition-all transform group-hover:scale-110 duration-300">
                                <span className="material-symbols-outlined text-2xl">{report.icon}</span>
                            </div>
                            <h3 className="text-white font-bold tracking-tight">{report.title}</h3>
                        </div>
                        <p className="text-sm text-text-muted mb-6 leading-relaxed">{report.desc}</p>
                        <button className="text-xs font-semibold text-primary flex items-center gap-1 group-hover:gap-2 transition-all">
                            Generate Report 
                            <span className="material-symbols-outlined text-sm">arrow_forward</span>
                        </button>
                    </div>
                ))}
            </div>
            <footer className="mt-12 text-center text-xs text-[#64748b] pb-6 flex flex-col items-center gap-1">
                <span>© 2026 FlyFleet Enterprise. All rights reserved.</span>
                <span className="text-[10px]">v2.4.1 | Server Time: <span className="font-mono">{new Date().getUTCHours()}:{new Date().getUTCMinutes().toString().padStart(2, '0')} UTC</span></span>
            </footer>
        </div>
    );
};

export default Reports;
