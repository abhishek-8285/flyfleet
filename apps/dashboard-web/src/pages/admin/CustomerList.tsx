import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Plus, Search, RefreshCw, User, Mail, Phone, MapPin, Building2 } from "lucide-react";
import { customerV2Service, Customer } from "../../services/customerV2Service";

export const CustomerList: React.FC = () => {
    const navigate = useNavigate();
    const [customers, setCustomers] = useState<Customer[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchQuery, setSearchQuery] = useState("");

    const fetchCustomers = async () => {
        setLoading(true);
        try {
            const { customers: list } = await customerV2Service.listCustomers(100, 0);
            setCustomers(list);
        } catch (error) {
            console.error("Failed to fetch customers", error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchCustomers();
    }, []);

    const filteredCustomers = customers.filter(c => 
        c.name.toLowerCase().includes(searchQuery.toLowerCase()) || 
        c.email?.toLowerCase().includes(searchQuery.toLowerCase())
    );

    return (
        <div className="w-full space-y-6 animate-in fade-in duration-500">
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                <div>
                    <h1 className="text-2xl font-bold text-white tracking-tight">Customer Management</h1>
                    <p className="text-gray-400 text-sm">Manage and track your registered customer base</p>
                </div>
                <button
                    onClick={() => navigate("/admin/customers/new")}
                    className="flex items-center gap-2 px-6 py-2.5 bg-primary text-white rounded-xl hover:bg-primary/90 transition-all font-semibold shadow-lg shadow-primary/20"
                >
                    <Plus className="w-5 h-5" />
                    New Customer
                </button>
            </div>

            {/* Toolbar */}
            <div className="bg-gray-900/50 backdrop-blur-xl border border-white/10 p-4 rounded-2xl flex flex-col md:flex-row gap-4 items-center">
                <div className="relative flex-1 w-full">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
                    <input
                        type="text"
                        placeholder="Search by company name or email..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="w-full bg-white/5 border border-white/10 rounded-xl pl-10 pr-4 py-2 text-white focus:outline-none focus:border-primary transition-colors text-sm"
                    />
                </div>
                <button 
                    onClick={fetchCustomers}
                    className="p-2 text-gray-400 hover:text-white hover:bg-white/5 rounded-lg transition-colors"
                >
                    <RefreshCw className={`w-5 h-5 ${loading ? 'animate-spin text-primary' : ''}`} />
                </button>
            </div>

            {/* Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {loading && Array.from({ length: 6 }).map((_, i) => (
                    <div key={i} className="h-48 bg-gray-900/30 border border-white/5 rounded-2xl animate-pulse" />
                ))}

                {!loading && filteredCustomers.length === 0 && (
                    <div className="col-span-full py-20 text-center">
                        <div className="inline-flex items-center justify-center p-6 bg-gray-900/50 border border-white/10 rounded-full mb-4">
                            <User className="w-12 h-12 text-gray-700" />
                        </div>
                        <h3 className="text-white font-semibold text-lg">No Customers Found</h3>
                        <p className="text-gray-500">Try adjusting your search or add a new customer</p>
                    </div>
                )}

                {!loading && filteredCustomers.map((customer) => (
                    <div 
                        key={customer.id}
                        className="group bg-gray-900/50 backdrop-blur-xl border border-white/10 hover:border-primary/50 p-6 rounded-2xl transition-all hover:shadow-2xl hover:shadow-primary/5 cursor-pointer relative overflow-hidden"
                    >
                        <div className="absolute top-0 right-0 p-4 opacity-0 group-hover:opacity-100 transition-opacity">
                            <span className="px-2 py-1 bg-primary/10 text-primary text-[10px] font-bold rounded tracking-widest uppercase">
                                View Profile
                            </span>
                        </div>

                        <div className="flex items-start gap-4 mb-4">
                            <div className="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center text-primary border border-primary/20">
                                <Building2 className="w-6 h-6" />
                            </div>
                            <div className="flex-1 min-w-0">
                                <h3 className="text-white font-bold truncate group-hover:text-primary transition-colors">{customer.name}</h3>
                                {customer.gstin ? (
                                    <p className="text-[10px] font-mono text-gray-500 uppercase">{customer.gstin}</p>
                                ) : (
                                    <p className="text-[10px] text-gray-600">No GST Registered</p>
                                )}
                            </div>
                        </div>

                        <div className="space-y-2 mb-4 border-y border-white/5 py-3">
                            {customer.email && (
                                <div className="flex items-center gap-2 text-xs text-gray-400">
                                    <Mail className="w-3.5 h-3.5 text-gray-600" />
                                    <span className="truncate">{customer.email}</span>
                                </div>
                            )}
                            {customer.phone && (
                                <div className="flex items-center gap-2 text-xs text-gray-400">
                                    <Phone className="w-3.5 h-3.5 text-gray-600" />
                                    <span>{customer.phone}</span>
                                </div>
                            )}
                        </div>

                        <div className="flex items-center gap-2 text-xs text-gray-500">
                            <MapPin className="w-3.5 h-3.5 flex-shrink-0" />
                            <p className="truncate line-clamp-1">{customer.address || "No address provided"}</p>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};
