import React from 'react';
import { Link } from 'react-router-dom';

export const CriticalVehiclesTable: React.FC = () => {
    return (
        <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden flex flex-col">
            <div className="p-4 border-b border-border-color flex justify-between items-center bg-surface-darker">
                <h3 className="text-white font-semibold">Critical Vehicle Status</h3>
                <div className="flex gap-3">
                    <button className="text-xs font-medium text-text-muted hover:text-white flex items-center gap-1 bg-border-color/30 px-2 py-1 rounded transition-colors">
                        <span className="material-symbols-outlined text-[14px]">download</span>
                        Export Selected
                    </button>
                    <button className="text-xs font-medium text-text-muted hover:text-white flex items-center gap-1 bg-border-color/30 px-2 py-1 rounded transition-colors">
                        <span className="material-symbols-outlined text-[14px]">edit_note</span>
                        Batch Status Change
                    </button>
                </div>
            </div>
            <div className="overflow-x-auto">
                <table className="w-full text-left text-sm text-text-muted">
                    <thead className="bg-[#1d2936] text-xs uppercase font-medium">
                        <tr>
                            <th className="px-6 py-3 w-10">
                                <input type="checkbox" className="rounded bg-surface-dark border-border-color text-primary focus:ring-primary w-4 h-4" />
                            </th>
                            <th className="px-6 py-3 cursor-pointer hover:text-white group">
                                <div className="flex items-center gap-1">Vehicle ID <span className="material-symbols-outlined text-[12px] opacity-100 text-primary">arrow_upward</span></div>
                            </th>
                            <th className="px-6 py-3 cursor-pointer hover:text-white group">
                                <div className="flex items-center gap-1">Driver <span className="material-symbols-outlined text-[12px] opacity-0 group-hover:opacity-100">sort</span></div>
                            </th>
                            <th className="px-6 py-3">Status</th>
                            <th className="px-6 py-3">Location</th>
                            <th className="px-6 py-3 cursor-pointer hover:text-white group">
                                <div className="flex items-center gap-1">Last Update <span className="material-symbols-outlined text-[12px] opacity-0 group-hover:opacity-100">arrow_downward</span></div>
                            </th>
                            <th className="px-6 py-3">SLA Aging</th>
                            <th className="px-6 py-3 text-right">Action</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-border-color">
                        <tr className="hover:bg-[#1d2936] transition-colors bg-primary/5">
                            <td className="px-6 py-3">
                                <input type="checkbox" defaultChecked className="rounded bg-surface-dark border-border-color text-primary focus:ring-primary w-4 h-4" />
                            </td>
                            <td className="px-6 py-3 font-medium text-white">#301</td>
                            <td className="px-6 py-3">M. Johnson</td>
                            <td className="px-6 py-3"><span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-500/10 text-red-500 border border-red-500/20">Engine Alert</span></td>
                            <td className="px-6 py-3 text-xs">I-95 South, Exit 4</td>
                            <td className="px-6 py-3 text-xs text-white">2 min ago</td>
                            <td className="px-6 py-3 text-xs text-red-400 font-medium">15 min</td>
                            <td className="px-6 py-3 text-right">
                                <Link to={`/dashboard/vehicles/301`} className="text-primary hover:text-white text-xs">Detail</Link>
                            </td>
                        </tr>
                        <tr className="hover:bg-[#1d2936] transition-colors">
                            <td className="px-6 py-3">
                                <input type="checkbox" className="rounded bg-surface-dark border-border-color text-primary focus:ring-primary w-4 h-4" />
                            </td>
                            <td className="px-6 py-3 font-medium text-white">#442</td>
                            <td className="px-6 py-3">S. Williams</td>
                            <td className="px-6 py-3"><span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-orange-500/10 text-orange-500 border border-orange-500/20">Maintenance</span></td>
                            <td className="px-6 py-3 text-xs">Depot A</td>
                            <td className="px-6 py-3 text-xs text-white">1 hr ago</td>
                            <td className="px-6 py-3 text-xs text-yellow-500">2 hrs</td>
                            <td className="px-6 py-3 text-right">
                                <Link to={`/dashboard/vehicles/442`} className="text-primary hover:text-white text-xs">Detail</Link>
                            </td>
                        </tr>
                        <tr className="hover:bg-[#1d2936] transition-colors">
                            <td className="px-6 py-3">
                                <input type="checkbox" className="rounded bg-surface-dark border-border-color text-primary focus:ring-primary w-4 h-4" />
                            </td>
                            <td className="px-6 py-3 font-medium text-white">#205</td>
                            <td className="px-6 py-3">K. Davis</td>
                            <td className="px-6 py-3"><span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-500/10 text-red-500 border border-red-500/20">Accident</span></td>
                            <td className="px-6 py-3 text-xs">Route 66, MP 12</td>
                            <td className="px-6 py-3 text-xs text-white">5 min ago</td>
                            <td className="px-6 py-3 text-xs text-red-500 font-bold">45 min</td>
                            <td className="px-6 py-3 text-right">
                                <Link to={`/dashboard/vehicles/205`} className="text-primary hover:text-white text-xs">Detail</Link>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
            <div className="bg-surface-darker px-4 py-3 border-t border-border-color text-[11px] text-text-muted flex justify-between items-center">
                <span>Showing 1-50 of 842 records</span>
                <div className="flex gap-2">
                    <button className="p-1 hover:text-white disabled:opacity-50" disabled><span className="material-symbols-outlined text-sm">chevron_left</span></button>
                    <span className="text-white">1</span>
                    <button className="p-1 hover:text-white"><span className="material-symbols-outlined text-sm">chevron_right</span></button>
                </div>
            </div>
        </div>
    );
};
