import { Clock, FastForward, Pause, Play, Rewind } from "lucide-react";
import React from "react";

interface TimeScrubberProps {
  onTimeChange: (timestamp: number) => void;
  isLive: boolean;
  onLiveToggle: (isLive: boolean) => void;
  startTime: number;
  endTime: number;
}

const TimeScrubber: React.FC<TimeScrubberProps> = ({
  onTimeChange,
  isLive,
  onLiveToggle,
  startTime,
  endTime,
}) => {
  const [currentValue, setCurrentValue] = React.useState(endTime);
  const [isPlaying, setIsPlaying] = React.useState(false);

  React.useEffect(() => {
    if (isLive) {
      setCurrentValue(endTime);
      setIsPlaying(false);
    }
  }, [isLive, endTime]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = parseInt(e.target.value, 10);
    setCurrentValue(val);
    onTimeChange(val);
    if (isLive) onLiveToggle(false);
  };

  const formatTime = (ts: number) => {
    return new Date(ts).toLocaleTimeString([], {
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    });
  };

  return (
    <div className="w-full bg-slate-900/90 backdrop-blur-xl border border-white/10 rounded-2xl p-4 shadow-2xl flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Clock className="w-4 h-4 text-blue-400" />
          <span className="text-xs font-bold text-slate-300 uppercase tracking-widest">
            {isLive ? "Live Operations" : "Incident Replay"}
          </span>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-[10px] font-mono text-slate-500">
            {formatTime(currentValue)}
          </span>
          <button
            onClick={() => onLiveToggle(!isLive)}
            className={`px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-tighter transition-all ${
              isLive
                ? "bg-emerald-500/20 text-emerald-400 border border-emerald-500/50"
                : "bg-slate-800 text-slate-500 border border-slate-700"
            }`}
          >
            {isLive ? "● Live" : "Go Live"}
          </button>
        </div>
      </div>

      <div className="flex items-center gap-4">
        <div className="flex items-center gap-1">
          <button className="p-1.5 hover:bg-white/5 rounded-lg text-slate-400 transition-colors">
            <Rewind className="w-4 h-4" />
          </button>
          <button
            onClick={() => setIsPlaying(!isPlaying)}
            className="p-1.5 hover:bg-white/5 rounded-lg text-slate-200 transition-colors"
          >
            {isPlaying ? (
              <Pause className="w-4 h-4" />
            ) : (
              <Play className="w-4 h-4" />
            )}
          </button>
          <button className="p-1.5 hover:bg-white/5 rounded-lg text-slate-400 transition-colors">
            <FastForward className="w-4 h-4" />
          </button>
        </div>

        <div className="flex-1 relative group py-2">
          <input
            type="range"
            min={startTime}
            max={endTime}
            value={currentValue}
            onChange={handleChange}
            className="w-full h-1.5 bg-slate-800 rounded-lg appearance-none cursor-pointer accent-blue-500"
          />
          <div className="absolute -top-6 left-1/2 -translate-x-1/2 bg-slate-800 text-[9px] text-white px-2 py-1 rounded opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none">
            {formatTime(currentValue)}
          </div>
        </div>
      </div>

      <div className="flex justify-between text-[9px] text-slate-600 font-mono uppercase">
        <span>{formatTime(startTime)}</span>
        <span>{formatTime(endTime)}</span>
      </div>
    </div>
  );
};

export default TimeScrubber;
