import {
    ArrowTrendingUpIcon,
    ChartBarIcon,
    CurrencyDollarIcon,
    FunnelIcon
} from '@heroicons/react/24/outline';
import React, { useState } from 'react';
import apiClient from '../../../services/apiClient';

const ForecastingDashboard: React.FC = () => {
    const [selectedRoute, setSelectedRoute] = useState('BLR-MYS');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [experiments, setExperiments] = useState<any[]>([]);

    const [forecast, setForecast] = useState({
        currentDemand: 85,
        predictedDemand: 120,
        suggestedPrice: 1500,
        basePrice: 1200,
        confidence: 92
    });

    const loadForecast = async () => {
        setLoading(true);
        setError(null);
        try {
            const [forecastResponse, expResponse] = await Promise.all([
                apiClient.get('/admin/revenue/forecast', { params: { route_id: selectedRoute } }) as any,
                apiClient.get('/admin/revenue/experiments', { params: { limit: 10 } }) as any,
            ]);
            const data = forecastResponse?.forecast || {};
            setForecast({
                currentDemand: Number(data.current_demand || 0),
                predictedDemand: Number(data.predicted_demand || 0),
                suggestedPrice: Number(data.suggested_price || 0),
                basePrice: Number(data.base_price || 0),
                confidence: Number(data.confidence_pct || 0),
            });
            setExperiments(Array.isArray(expResponse?.experiments) ? expResponse.experiments : []);
        } catch (err: any) {
            setError(err?.message || 'Failed to load forecasting data');
        } finally {
            setLoading(false);
        }
    };

    React.useEffect(() => {
        loadForecast();
    }, []);

    return (
        <div className="space-y-6 p-6 bg-slate-900 min-h-screen text-white">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-2xl font-bold flex items-center gap-2">
                        <ArrowTrendingUpIcon className="h-8 w-8 text-emerald-500" />
                        Revenue Optimization & Forecasting
                    </h1>
                    <p className="text-slate-400">AI-driven demand prediction and dynamic pricing</p>
                </div>
                <div className="flex gap-3">
                    <select
                        className="bg-slate-800 border border-slate-700 rounded-lg px-4 py-2 text-white"
                        value={selectedRoute}
                        onChange={(e) => setSelectedRoute(e.target.value)}
                    >
                        <option value="BLR-MYS">Bangalore - Mysore</option>
                        <option value="MUM-PUN">Mumbai - Pune</option>
                        <option value="DEL-AGR">Delhi - Agra</option>
                    </select>
                    <button onClick={loadForecast} className="px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition">
                        Run Forecast
                    </button>
                </div>
            </div>
            {error && <div className="text-sm text-red-300 bg-red-500/10 border border-red-500/30 rounded px-3 py-2">{error}</div>}

            {/* Metrics Grid */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
                <div className="bg-slate-800 p-6 rounded-xl border border-slate-700">
                    <div className="flex justify-between items-start mb-2">
                        <h3 className="text-slate-400 text-sm font-medium">Predicted Demand (Next 7 Days)</h3>
                        <ChartBarIcon className="h-5 w-5 text-blue-500" />
                    </div>
                    <div className="text-3xl font-bold text-white">{loading ? '--' : forecast.predictedDemand}</div>
                    <div className="text-sm text-emerald-400 mt-1">↑ 15% vs Last Week</div>
                </div>

                <div className="bg-slate-800 p-6 rounded-xl border border-slate-700">
                    <div className="flex justify-between items-start mb-2">
                        <h3 className="text-slate-400 text-sm font-medium">Suggested Price</h3>
                        <CurrencyDollarIcon className="h-5 w-5 text-yellow-500" />
                    </div>
                    <div className="text-3xl font-bold text-white">₹{loading ? '--' : forecast.suggestedPrice}</div>
                    <div className="text-sm text-yellow-400 mt-1">
                        Surge: {forecast.basePrice > 0 ? (forecast.suggestedPrice / forecast.basePrice).toFixed(1) : '0.0'}x
                    </div>
                </div>

                <div className="bg-slate-800 p-6 rounded-xl border border-slate-700">
                    <div className="flex justify-between items-start mb-2">
                        <h3 className="text-slate-400 text-sm font-medium">Model Confidence</h3>
                        <FunnelIcon className="h-5 w-5 text-purple-500" />
                    </div>
                    <div className="text-3xl font-bold text-white">{loading ? '--' : forecast.confidence}%</div>
                    <div className="text-sm text-slate-500 mt-1">Prophet Model v1.2</div>
                </div>

                <div className="bg-slate-800 p-6 rounded-xl border border-slate-700">
                    <div className="flex justify-between items-start mb-2">
                        <h3 className="text-slate-400 text-sm font-medium">Capacity Utilization</h3>
                        <ChartBarIcon className="h-5 w-5 text-red-500" />
                    </div>
                    <div className="text-3xl font-bold text-white">{loading ? '--' : forecast.currentDemand}%</div>
                    <div className="text-sm text-red-400 mt-1">High Demand Alert</div>
                </div>
            </div>

            {/* Charts Placeholder */}
            <div className="bg-slate-800 rounded-xl border border-slate-700 p-6 h-96 flex items-center justify-center text-slate-500">
                [Demand Forecast Chart Placeholder - Recharts or similar would go here]
            </div>

            {/* Pricing Experiments */}
            <div className="bg-slate-800 rounded-xl border border-slate-700 overflow-hidden">
                <div className="p-6 border-b border-slate-700">
                    <h3 className="text-lg font-bold text-white">Active A/B Tests</h3>
                </div>
                <table className="w-full text-left text-sm text-slate-400">
                    <thead className="bg-slate-900/50 uppercase font-medium">
                        <tr>
                            <th className="px-6 py-4">Experiment</th>
                            <th className="px-6 py-4">Variant A</th>
                            <th className="px-6 py-4">Variant B</th>
                            <th className="px-6 py-4">Status</th>
                            <th className="px-6 py-4">Lift</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-700">
                        {experiments.length === 0 && (
                            <tr>
                                <td className="px-6 py-4 text-slate-500" colSpan={5}>
                                    {loading ? 'Loading experiments...' : 'No experiments available'}
                                </td>
                            </tr>
                        )}
                        {experiments.map((exp) => (
                            <tr key={exp.id}>
                                <td className="px-6 py-4 font-medium text-white">{exp.name}</td>
                                <td className="px-6 py-4">{exp.variants?.[0] || '-'}</td>
                                <td className="px-6 py-4">{exp.variants?.[1] || '-'}</td>
                                <td className="px-6 py-4"><span className="px-2 py-1 bg-green-500/10 text-green-500 rounded text-xs">{exp.status || 'UNKNOWN'}</span></td>
                                <td className="px-6 py-4 text-green-400">{exp.metrics?.revenue_lift || '-'}</td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
};

export default ForecastingDashboard;
