// components/PlaybackControls.tsx

import React from 'react';
import { useSelector, useDispatch } from 'react-redux';
import { RootState } from '../../../store';
import { setPlaybackState } from '../store/trackingSlice';

const speeds = [1, 2, 4, 8] as const;

export const PlaybackControls: React.FC = () => {
    const dispatch = useDispatch();
    const { isPlaying, playbackSpeed, playbackPosition } = useSelector(
        (state: RootState) => ({
            isPlaying: state.tracking.isPlaying,
            playbackSpeed: state.tracking.playbackSpeed,
            playbackPosition: state.tracking.playbackPosition,
        })
    );
    const historyTrail = useSelector((state: RootState) => state.tracking.historyTrail);

    if (!historyTrail) return null;

    const totalPoints = historyTrail.points.length;

    return (
        <div className="bg-white rounded-lg shadow-lg p-4 space-y-3">
            <div className="flex items-center justify-between">
                <h3 className="font-semibold text-gray-700">Trip Playback</h3>
                <span className="text-xs text-gray-500">
                    {playbackPosition !== null ? `${playbackPosition + 1}/${totalPoints}` : 'Ready'}
                </span>
            </div>

            {/* Playback Controls */}
            <div className="flex items-center gap-2">
                <button
                    onClick={() => dispatch(setPlaybackState({ isPlaying: !isPlaying }))}
                    className="flex-1 bg-blue-600 text-white py-2 px-4 rounded hover:bg-blue-700"
                >
                    {isPlaying ? '⏸ Pause' : '▶ Play'}
                </button>
                <button
                    onClick={() => dispatch(setPlaybackState({ isPlaying: false, position: 0 }))}
                    className="bg-gray-200 text-gray-700 py-2 px-3 rounded hover:bg-gray-300"
                >
                    ⏮
                </button>
            </div>

            {/* Speed Controls */}
            <div className="flex gap-1">
                {speeds.map(speed => (
                    <button
                        key={speed}
                        onClick={() => dispatch(setPlaybackState({ isPlaying, speed }))}
                        className={`flex-1 py-1 px-2 rounded text-sm font-medium ${
                            playbackSpeed === speed
                                ? 'bg-blue-600 text-white'
                                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                        }`}
                    >
                        {speed}x
                    </button>
                ))}
            </div>

            {/* Progress Bar */}
            <div className="w-full bg-gray-200 rounded-full h-1.5">
                <div
                    className="bg-blue-600 h-1.5 rounded-full transition-all"
                    style={{
                        width: playbackPosition !== null
                            ? `${((playbackPosition + 1) / totalPoints) * 100}%`
                            : '0%'
                    }}
                />
            </div>

            {/* Summary Stats */}
            <div className="grid grid-cols-2 gap-2 text-xs text-gray-500">
                <div>
                    <span>Distance:</span>
                    <span className="font-medium ml-1">{historyTrail.summary.total_distance_km.toFixed(1)} km</span>
                </div>
                <div>
                    <span>Avg Speed:</span>
                    <span className="font-medium ml-1">{historyTrail.summary.avg_speed_kmh.toFixed(1)} km/h</span>
                </div>
                <div>
                    <span>Max Speed:</span>
                    <span className="font-medium ml-1">{historyTrail.summary.max_speed_kmh.toFixed(1)} km/h</span>
                </div>
                <div>
                    <span>Duration:</span>
                    <span className="font-medium ml-1">{historyTrail.summary.duration_minutes} min</span>
                </div>
            </div>
        </div>
    );
};
