import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { 
    Plus, 
    Building2, 
    ChevronRight, 
    ShieldCheck,
    User,
    Mail,
    Phone,
    Calendar,
    Users,
    AlertCircle
} from "lucide-react";
import { DataGrid } from "../../components/common/DataGrid";
import apiClient from "../../services/apiClient";

export const CustomerList: React.FC = () => {
    const navigate = useNavigate();
    const [refreshTrigger] = useState(0);
    const [stats, setStats] = useState({ total: 0, active: 0, suspended: 0, pending: 0 });

    useEffect(() => {
        // Fetch snapshot for the summary strip
        apiClient.get('/admin/customers?size=0')
            .then((res: any) => {
                const meta = res?.data?.meta || res?.meta;
                if (meta) {
                    setStats(prev => ({ ...prev, total: meta.total }));
                }
            })
            .catch(err => console.error("Failed to fetch summary stats", err));
    }, [refreshTrigger]);

    const columns = [
        { 
            header: "Customer Entity", 
            field: "name",
            render: (row: any) => (
                <div className="flex items-center gap-4">
                    <div className="w-10 h-10 rounded-xl bg-slate-800/50 flex items-center justify-center text-blue-400 border border-white/5 shadow-inner">
                        <Building2 className="w-5 h-5" />
                    </div>
                    <div>
                        <p className="font-bold text-slate-200 group-hover:text-blue-400 transition-colors">{row.name}</p>
                        <p className="text-[10px] text-slate-500 font-mono uppercase leading-tight tracking-tighter">
                            ID: {row.id.substring(0, 8)}...
                        </p>
                    </div>
                </div>
            )
        },
        { 
            header: "Compliance", 
            field: "gst_number",
            render: (row: any) => (
                row.gst_number ? (
                    <div className="inline-flex items-center gap-1.5 px-2.5 py-1 bg-emerald-500/5 border border-emerald-500/10 rounded-lg text-[10px] font-mono text-emerald-400/80">
                        <ShieldCheck className="w-3 h-3 text-emerald-500" />
                        {row.gst_number}
                    </div>
                ) : (
                    <div className="inline-flex items-center gap-1.5 px-2.5 py-1 bg-amber-500/5 border border-amber-500/10 rounded-lg text-[10px] font-mono text-amber-500/80">
                        <AlertCircle className="w-3 h-3" />
                        MISSING GST
                    </div>
                )
            )
        },
        { 
            header: "Primary Contact", 
            field: "contact_email",
            render: (row: any) => (
                <div className="space-y-1">
                    <div className="flex items-center gap-2 text-xs text-slate-400 font-medium tracking-tight">
                        <Mail className="w-3 h-3 text-slate-600" />
                        {row.contact_email}
                    </div>
                    {row.contact_phone && (
                        <div className="flex items-center gap-2 text-[10px] text-slate-500 font-mono">
                            <Phone className="w-3 h-3 text-slate-700" />
                            {row.contact_phone}
                        </div>
                    )}
                </div>
            )
        },
        {
            header: "Registered On",
            field: "created_at",
            render: (row: any) => (
                <div className="flex items-center gap-2 text-xs text-slate-500 font-mono">
                    <Calendar className="w-3 h-3 text-slate-700" />
                    {new Date(row.created_at).toLocaleDateString()}
                </div>
            )
        },
        { 
            header: "Status", 
            field: "status",
            render: (row: any) => (
                <span className={`px-2.5 py-1 rounded-full text-[10px] font-black uppercase tracking-widest border ${
                    row.status === 'ACTIVE' 
                        ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' 
                        : 'bg-slate-700/50 text-slate-400 border-white/5'
                }`}>
                    {row.status || 'ACTIVE'}
                </span>
            )
        },
        { 
            header: "Actions", 
            field: "actions",
            width: "80px",
            render: (row: any) => (
                <div className="flex items-center gap-2">
                    <button 
                        onClick={() => navigate(`/admin/customers/${row.id}`)}
                        className="p-2 bg-white/5 hover:bg-blue-500/10 text-slate-500 hover:text-blue-400 rounded-xl transition-all border border-transparent hover:border-blue-500/20 shadow-sm"
                        title="View Intelligence"
                    >
                        <ChevronRight className="w-4 h-4" />
                    </button>
                </div>
            )
        }
    ];

    return (
        <div className="p-8 pb-20 space-y-10 animate-in fade-in duration-700">
            {/* Page Header */}
            <div className="flex flex-col md:flex-row md:items-end justify-between gap-8">
                <div className="space-y-3">
                    <div className="flex items-center gap-4">
                        <div className="w-12 h-12 rounded-2xl bg-gradient-to-br from-blue-600 to-indigo-700 flex items-center justify-center text-white shadow-xl shadow-blue-900/20">
                            <Building2 className="w-6 h-6" />
                        </div>
                        <div>
                           <h1 className="text-4xl font-black text-white tracking-tighter">Customer Management</h1>
                           <p className="text-slate-500 text-sm font-medium">Administration of partner entities and billing profiles.</p>
                        </div>
                    </div>
                </div>
                
                <button 
                    onClick={() => navigate("/admin/customers/new")}
                    className="flex items-center gap-2 px-8 py-3.5 bg-blue-600 text-white font-black rounded-[1.25rem] hover:bg-blue-500 transition-all hover:scale-[1.02] active:scale-[0.98] shadow-2xl shadow-blue-900/20 uppercase text-xs tracking-widest"
                >
                    <Plus className="w-5 h-5" />
                    Add New Partner
                </button>
            </div>

            {/* Summary Strip */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
                {[
                    { label: "Total Partners", value: stats.total, icon: Users, color: "blue" },
                    { label: "Active Accounts", value: stats.total, icon: ShieldCheck, color: "emerald" }, // Using total as fallback until real counts endpoint
                    { label: "Compliance Risk", value: 0, icon: AlertCircle, color: "amber" },
                    { label: "Suspended", value: 0, icon: AlertCircle, color: "slate" }
                ].map((stat, i) => (
                    <div key={i} className="bg-slate-900/40 backdrop-blur-md p-6 rounded-[2rem] border border-white/5 space-y-2 group hover:bg-slate-900/60 transition-all">
                        <div className="flex items-center justify-between">
                            <stat.icon className={`w-5 h-5 text-${stat.color}-500 opacity-80`} />
                            <span className="text-[10px] font-black text-white/20 uppercase tracking-[0.2em]">Live Data</span>
                        </div>
                        <div className="text-3xl font-black text-white tracking-tighter">{stat.value}</div>
                        <div className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">{stat.label}</div>
                    </div>
                ))}
            </div>

            {/* DataGrid Implementation */}
            <div className="bg-slate-900/20 rounded-[2.5rem] border border-white/5 overflow-hidden shadow-2xl shadow-black/20">
                <DataGrid 
                    endpoint="/admin/customers"
                    columns={columns}
                    refreshTrigger={refreshTrigger}
                    filters={[
                        { field: "name", label: "Company Name", type: "text" },
                        { field: "status", label: "Status", type: "select", options: [
                            { value: "ACTIVE", label: "Active" },
                            { value: "SUSPENDED", label: "Suspended" },
                            { value: "PENDING", label: "Pending" }
                        ]}
                    ]}
                />
            </div>
        </div>
    );
};
