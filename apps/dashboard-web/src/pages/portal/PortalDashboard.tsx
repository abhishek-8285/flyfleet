import axios from 'axios';
import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';

const PortalDashboard: React.FC = () => {
    const [stats, setStats] = useState({
        activeShipments: 0,
        pendingInvoices: 0,
        totalSpent: 0
    });
    const [recentShipments, setRecentShipments] = useState<any[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const token = localStorage.getItem('customer_token');
                if (!token) return;

                const config = { headers: { Authorization: `Bearer ${token}` } };

                // Fetch shipments
                const shipmentsRes = await axios.get('/api/v1/customer/shipments', config);
                setRecentShipments(shipmentsRes.data.slice(0, 5)); // Top 5
                setStats(prev => ({ ...prev, activeShipments: shipmentsRes.data.filter((s: any) => s.status !== 'DELIVERED').length }));

                // Fetch invoices
                const invoicesRes = await axios.get('/api/v1/customer/invoices', config);
                setStats(prev => ({
                    ...prev,
                    pendingInvoices: invoicesRes.data.filter((i: any) => i.status === 'PENDING').length,
                    totalSpent: invoicesRes.data.reduce((acc: number, curr: any) => acc + (curr.amount || 0), 0)
                }));

            } catch (err) {
                console.error('Failed to fetch dashboard data', err);
            } finally {
                setLoading(false);
            }
        };

        fetchData();
    }, []);

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-2xl font-bold text-white tracking-tight">Dashboard</h1>
                    <p className="text-text-muted text-sm">Welcome back, Acme Corp</p>
                </div>
                <Link to="/portal/bookings" className="bg-primary hover:bg-primary-dark text-white px-4 py-2 rounded-lg text-sm font-medium transition-colors shadow-lg shadow-primary/20 flex items-center gap-2">
                    <span className="material-symbols-outlined text-[18px]">add</span>
                    New Booking
                </Link>
            </div>

            {/* Stats Cards */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div className="bg-surface-dark border border-border-color p-5 rounded-xl">
                    <div className="flex justify-between items-start mb-4">
                        <div className="p-2 bg-blue-500/10 rounded-lg text-blue-400">
                            <span className="material-symbols-outlined text-[24px]">local_shipping</span>
                        </div>
                        <span className="text-xs font-semibold text-green-400 bg-green-400/10 px-2 py-1 rounded">+12%</span>
                    </div>
                    <div className="text-3xl font-bold text-white mb-1">{stats.activeShipments}</div>
                    <div className="text-xs text-text-muted">Active Shipments</div>
                </div>

                <div className="bg-surface-dark border border-border-color p-5 rounded-xl">
                    <div className="flex justify-between items-start mb-4">
                        <div className="p-2 bg-orange-500/10 rounded-lg text-orange-400">
                            <span className="material-symbols-outlined text-[24px]">pending_actions</span>
                        </div>
                    </div>
                    <div className="text-3xl font-bold text-white mb-1">{stats.pendingInvoices}</div>
                    <div className="text-xs text-text-muted">Pending Invoices</div>
                </div>

                <div className="bg-surface-dark border border-border-color p-5 rounded-xl">
                    <div className="flex justify-between items-start mb-4">
                        <div className="p-2 bg-green-500/10 rounded-lg text-green-400">
                            <span className="material-symbols-outlined text-[24px]">payments</span>
                        </div>
                    </div>
                    <div className="text-3xl font-bold text-white mb-1">${stats.totalSpent.toLocaleString()}</div>
                    <div className="text-xs text-text-muted">Total Spent (YTD)</div>
                </div>
            </div>

            {/* Recent Shipments Table */}
            <div className="bg-surface-dark border border-border-color rounded-xl overflow-hidden">
                <div className="p-5 border-b border-border-color flex justify-between items-center bg-surface-darker">
                    <h3 className="font-semibold text-white">Recent Shipments</h3>
                    <Link to="/portal/shipments" className="text-xs text-primary hover:text-white transition-colors">View All</Link>
                </div>
                <div className="overflow-x-auto">
                    <table className="w-full text-left text-sm text-text-muted">
                        <thead className="bg-[#111a22] text-xs uppercase font-semibold text-text-muted/70">
                            <tr>
                                <th className="px-5 py-3">Tracking ID</th>
                                <th className="px-5 py-3">Origin</th>
                                <th className="px-5 py-3">Destination</th>
                                <th className="px-5 py-3">Status</th>
                                <th className="px-5 py-3">ETA</th>
                                <th className="px-5 py-3">Actions</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-border-color">
                            {loading ? (
                                <tr><td colSpan={6} className="p-4 text-center">Loading...</td></tr>
                            ) : recentShipments.length === 0 ? (
                                <tr><td colSpan={6} className="p-4 text-center">No active shipments found.</td></tr>
                            ) : (
                                recentShipments.map((shipment) => (
                                    <tr key={shipment.id} className="hover:bg-surface-light/5 transition-colors">
                                        <td className="px-5 py-3 font-medium text-white">{shipment.tracking_number}</td>
                                        <td className="px-5 py-3">{shipment.origin?.address || 'N/A'}</td>
                                        <td className="px-5 py-3">{shipment.destination?.address || 'N/A'}</td>
                                        <td className="px-5 py-3">
                                            <span className={`px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wide 
                        ${shipment.status === 'DELIVERED' ? 'bg-green-500/10 text-green-500' :
                                                    shipment.status === 'IN_TRANSIT' ? 'bg-blue-500/10 text-blue-500' : 'bg-gray-500/10 text-gray-500'}`}>
                                                {shipment.status}
                                            </span>
                                        </td>
                                        <td className="px-5 py-3">{shipment.estimated_delivery ? new Date(shipment.estimated_delivery).toLocaleDateString() : '-'}</td>
                                        <td className="px-5 py-3">
                                            <Link to={`/portal/track/${shipment.tracking_number}`} className="text-primary hover:text-white text-xs font-medium">Track</Link>
                                        </td>
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    );
};

export default PortalDashboard;
