import {
    CpuChipIcon,
    PlayCircleIcon,
    SignalIcon,
    VideoCameraIcon
} from '@heroicons/react/24/outline';
import { format } from 'date-fns';
import React, { useEffect, useState } from 'react';
import { api } from '../../services/api';

interface VideoStream {
    id: string;
    vehicle_id: string;
    vehicle_name?: string; // Enriched
    status: 'active' | 'inactive' | 'error';
    url: string;
    started_at: string;
    resolution: string;
    fps: number;
}

const VideoTelematics: React.FC = () => {
    const [streams, setStreams] = useState<VideoStream[]>([]);
    const [loading, setLoading] = useState(true);
    const [selectedStream, setSelectedStream] = useState<VideoStream | null>(null);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        fetchStreams();
        // Poll every 30s
        const interval = setInterval(fetchStreams, 30000);
        return () => clearInterval(interval);
    }, []);

    const fetchStreams = async () => {
        try {
            setLoading(true);
            setError(null);
            const res = await api.listVideoStreams({ status: 'active' });
            const data = (res as any)?.data ?? res;
            const items = data?.streams || data?.items || data?.data || data || [];
            const enriched = (Array.isArray(items) ? items : []).map((s: any) => ({
                ...s,
                vehicle_name: s.vehicle_name || `Vehicle ${s.vehicle_id.substring(0, 8)}`
            }));
            setStreams(enriched);
            if (enriched.length > 0 && !selectedStream) {
                setSelectedStream(enriched[0]);
            }
        } catch (error) {
            console.error("Failed to fetch video streams", error);
            setError('Failed to load video streams.');
            setStreams([]);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="h-full flex flex-col bg-background-dark text-white space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-2xl font-bold tracking-tight">Video Telematics</h1>
                    <p className="text-text-muted text-sm">Real-time fleet monitoring and AI surveillance</p>
                </div>
                <div className="flex gap-2">
                    <span className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-red-500/10 border border-red-500/20 text-red-500 text-xs font-bold uppercase tracking-wider animate-pulse">
                        <span className="w-2 h-2 rounded-full bg-red-500"></span>
                        Live Feed
                    </span>
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 h-[calc(100vh-12rem)]">
                {/* Main Player Area */}
                <div className="lg:col-span-2 bg-surface-dark rounded-xl border border-border-color overflow-hidden flex flex-col relative shadow-2xl">
                    {selectedStream ? (
                        <>
                            <div className="flex-1 bg-black relative flex items-center justify-center group">
                                <div className="absolute inset-0 bg-gradient-to-t from-black/80 to-transparent p-6 flex flex-col justify-end z-20">
                                    <div className="flex justify-between items-end">
                                        <div>
                                            <h3 className="text-xl font-bold">{selectedStream.vehicle_name}</h3>
                                            <div className="flex items-center gap-3 text-sm text-gray-300 mt-1">
                                                <span className="flex items-center gap-1"><SignalIcon className="w-4 h-4 text-emerald-500" /> Excellent</span>
                                                <span className="flex items-center gap-1"><CpuChipIcon className="w-4 h-4" /> {selectedStream.resolution} @ {selectedStream.fps}fps</span>
                                            </div>
                                        </div>
                                        <div className="text-right">
                                            <p className="font-mono text-emerald-400 text-lg">{format(new Date(), 'HH:mm:ss')}</p>
                                            <p className="text-xs text-gray-400">Latency: {((selectedStream as any).latency_ms ?? '--')}ms</p>
                                        </div>
                                    </div>
                                </div>

                                {/* Placeholder Content if no real stream */}
                                <VideoCameraIcon className="w-24 h-24 text-gray-700 opacity-50" />
                                <div className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity bg-black/40 backdrop-blur-sm">
                                    <PlayCircleIcon className="w-16 h-16 text-white cursor-pointer hover:scale-110 transition-transform" />
                                </div>
                            </div>

                            {/* AI Insights Strip */}
                            <div className="h-16 bg-surface-darker border-t border-border-color p-3 flex items-center gap-4 overflow-x-auto">
                                <span className="text-xs font-bold uppercase text-text-muted shrink-0">AI Detection:</span>
                                <span className="bg-yellow-500/10 text-yellow-500 border border-yellow-500/20 px-2 py-1 rounded text-xs">⚠️ Distracted Driving (14:32)</span>
                                <span className="bg-emerald-500/10 text-emerald-500 border border-emerald-500/20 px-2 py-1 rounded text-xs">✓ Lane Keeping</span>
                            </div>
                        </>
                    ) : (
                        <div className="flex-1 flex flex-col items-center justify-center text-text-muted p-8">
                            <VideoCameraIcon className="w-16 h-16 mb-4 opacity-20" />
                            <p>Select a vehicle to view live stream</p>
                        </div>
                    )}
                </div>

                {/* Sidebar List */}
                <div className="bg-surface-dark rounded-xl border border-border-color overflow-hidden flex flex-col">
                    <div className="p-4 border-b border-border-color bg-surface-darker/50 flex justify-between items-center">
                        <h3 className="font-semibold text-sm uppercase tracking-wide text-text-muted">Active Cameras</h3>
                        <span className="bg-primary/20 text-primary text-xs font-bold px-2 py-0.5 rounded-full">{streams.length}</span>
                    </div>
                    <div className="flex-1 overflow-y-auto custom-scrollbar p-2 space-y-2">
                        {loading ? (
                            <div className="p-4 text-text-muted text-sm">Loading cameras...</div>
                        ) : error ? (
                            <div className="p-4 text-red-400 text-sm">{error}</div>
                        ) : streams.map(stream => (
                            <div
                                key={stream.id}
                                onClick={() => setSelectedStream(stream)}
                                className={`p-3 rounded-lg cursor-pointer transition-all border flex gap-3 ${selectedStream?.id === stream.id
                                        ? 'bg-primary/5 border-primary/50 shadow-md'
                                        : 'bg-surface-dark border-transparent hover:bg-surface-darker hover:border-border-color'
                                    }`}
                            >
                                <div className="w-20 h-14 bg-black rounded border border-gray-800 flex items-center justify-center text-gray-600 shrink-0 relative overflow-hidden">
                                    <VideoCameraIcon className="w-6 h-6" />
                                    <div className="absolute top-1 right-1 w-2 h-2 rounded-full bg-red-500 animate-pulse"></div>
                                </div>
                                <div className="flex-1 min-w-0">
                                    <h4 className="font-medium text-sm truncate">{stream.vehicle_name}</h4>
                                    <div className="flex items-center gap-2 mt-1">
                                        <span className="text-[10px] text-text-muted flex items-center gap-1">
                                            <SignalIcon className="w-3 h-3 text-emerald-500" /> Live
                                        </span>
                                        <span className="text-[10px] text-text-muted border-l border-gray-700 pl-2">
                                            {stream.resolution}
                                        </span>
                                    </div>
                                </div>
                            </div>
                        ))}
                        {streams.length === 0 && !loading && !error && (
                            <div className="p-8 text-center text-text-muted text-sm">No active cameras found.</div>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
};

export default VideoTelematics;
