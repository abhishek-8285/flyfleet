import React, { useEffect, useRef, useState } from "react";

// ── Date Helpers ──
const fmt = (d: Date): string => d.toISOString().split("T")[0];
const today = () => new Date();
const daysAgo = (n: number) => {
  const d = new Date();
  d.setDate(d.getDate() - n);
  return d;
};
const getDaysInMonth = (year: number, month: number) =>
  new Date(year, month + 1, 0).getDate();
const getFirstDayOfMonth = (year: number, month: number) =>
  new Date(year, month, 1).getDay();

type DatePreset = "today" | "yesterday" | "last7" | "last30" | "custom";

const PRESETS: { key: DatePreset; label: string; icon: string }[] = [
  { key: "today", label: "Today", icon: "today" },
  { key: "yesterday", label: "Yesterday", icon: "calendar_today" },
  { key: "last7", label: "Last 7 Days", icon: "date_range" },
  { key: "last30", label: "Last 1 Month", icon: "calendar_month" },
  { key: "custom", label: "Custom", icon: "tune" },
];

interface MiniCalendarProps {
  label: string;
  value: string;
  onChange: (dateStr: string) => void;
  view: { month: number; year: number };
  setView: React.Dispatch<React.SetStateAction<{ month: number; year: number }>>;
  highlightRange?: { start: string; end: string };
}

const MiniCalendar: React.FC<MiniCalendarProps> = ({
  label,
  value,
  onChange,
  view,
  setView,
  highlightRange,
}) => {
  const daysInMonth = getDaysInMonth(view.year, view.month);
  const firstDay = getFirstDayOfMonth(view.year, view.month);

  const handlePrevMonth = () => {
    setView((prev) => {
      if (prev.month === 0) return { month: 11, year: prev.year - 1 };
      return { month: prev.month - 1, year: prev.year };
    });
  };

  const handleNextMonth = () => {
    setView((prev) => {
      if (prev.month === 11) return { month: 0, year: prev.year + 1 };
      return { month: prev.month + 1, year: prev.year };
    });
  };

  const monthNames = [
    "January", "February", "March", "April", "May", "June",
    "July", "August", "September", "October", "November", "December"
  ];

  const days: React.ReactNode[] = [];
  // Padding cells
  for (let i = 0; i < firstDay; i++) {
    days.push(<div key={`pad-${i}`} className="h-8 w-8" />);
  }

  // Days cells
  for (let d = 1; d <= daysInMonth; d++) {
    const dateStr = `${view.year}-${String(view.month + 1).padStart(2, "0")}-${String(d).padStart(2, "0")}`;
    const isSelected = value === dateStr;

    let isInRange = false;
    if (highlightRange?.start && highlightRange?.end) {
      isInRange = dateStr >= highlightRange.start && dateStr <= highlightRange.end;
    }

    days.push(
      <button
        key={d}
        type="button"
        onClick={() => onChange(dateStr)}
        className={`h-8 w-8 rounded-xl text-[11px] font-bold transition-all flex items-center justify-center active:scale-90 ${
          isSelected
            ? "bg-primary text-white shadow-md shadow-primary/20 scale-105"
            : isInRange
            ? "bg-primary/10 text-primary hover:bg-primary/20"
            : "text-on-surface-variant hover:bg-white/5 hover:text-white"
        }`}
      >
        {d}
      </button>
    );
  }

  return (
    <div className="w-[230px] border border-outline/20 rounded-2xl p-3.5 bg-surface-darker/20">
      <div className="flex items-center justify-between mb-2 px-1">
        <span className="text-[10px] font-black uppercase tracking-wider text-on-surface-variant">
          {label}
        </span>
      </div>

      {/* Month Year Header */}
      <div className="flex items-center justify-between mb-3 px-1">
        <button
          type="button"
          onClick={handlePrevMonth}
          className="text-on-surface-variant hover:text-white transition-colors"
        >
          <span className="material-symbols-outlined text-[18px] font-bold">chevron_left</span>
        </button>
        <span className="text-[11.5px] font-black text-white uppercase tracking-wider">
          {monthNames[view.month].slice(0, 3)} {view.year}
        </span>
        <button
          type="button"
          onClick={handleNextMonth}
          className="text-on-surface-variant hover:text-white transition-colors"
        >
          <span className="material-symbols-outlined text-[18px] font-bold">chevron_right</span>
        </button>
      </div>

      {/* Weekdays Header */}
      <div className="grid grid-cols-7 gap-1 text-center mb-1.5">
        {["Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"].map((w) => (
          <span key={w} className="text-[8px] font-black text-on-surface-variant uppercase tracking-widest">
            {w}
          </span>
        ))}
      </div>

      {/* Days Grid */}
      <div className="grid grid-cols-7 gap-1">
        {days}
      </div>
    </div>
  );
};

export interface DateRangePickerProps {
  startDate: string;
  endDate: string;
  onChange: (startDate: string, endDate: string) => void;
  triggerLabel?: string;
  align?: "left" | "right";
}

export const DateRangePicker: React.FC<DateRangePickerProps> = ({
  startDate,
  endDate,
  onChange,
  triggerLabel = "Date Range",
  align = "left",
}) => {
  const [showDatePopup, setShowDatePopup] = useState(false);
  const [datePreset, setDatePreset] = useState<DatePreset | null>(null);
  const [tempStart, setTempStart] = useState(startDate);
  const [tempEnd, setTempEnd] = useState(endDate);
  const [dateError, setDateError] = useState("");
  const datePopupRef = useRef<HTMLDivElement>(null);

  const [startView, setStartView] = useState(() => {
    const d = startDate ? new Date(startDate) : new Date();
    return { month: d.getMonth(), year: d.getFullYear() };
  });

  const [endView, setEndView] = useState(() => {
    const d = endDate ? new Date(endDate) : new Date();
    return { month: d.getMonth(), year: d.getFullYear() };
  });

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (
        datePopupRef.current &&
        !datePopupRef.current.contains(e.target as Node)
      ) {
        setShowDatePopup(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  useEffect(() => {
    if (showDatePopup) {
      setTempStart(startDate);
      setTempEnd(endDate);
      setDateError("");
      setDatePreset(startDate && endDate ? "custom" : null);
    }
  }, [showDatePopup, startDate, endDate]);

  useEffect(() => {
    if (tempStart) {
      const d = new Date(tempStart);
      setStartView({ month: d.getMonth(), year: d.getFullYear() });
    }
    if (tempEnd) {
      const d = new Date(tempEnd);
      setEndView({ month: d.getMonth(), year: d.getFullYear() });
    }
  }, [tempStart, tempEnd]);

  const applyPreset = (key: DatePreset) => {
    setDatePreset(key);
    setDateError("");
    let from = "", to = "";
    switch (key) {
      case "today":
        from = to = fmt(today());
        break;
      case "yesterday":
        from = to = fmt(daysAgo(1));
        break;
      case "last7":
        from = fmt(daysAgo(7));
        to = fmt(today());
        break;
      case "last30":
        from = fmt(daysAgo(30));
        to = fmt(today());
        break;
      case "custom":
        from = tempStart;
        to = tempEnd;
        break;
    }
    setTempStart(from);
    setTempEnd(to);
  };

  const validateDates = (): boolean => {
    if (tempStart && tempEnd && tempEnd < tempStart) {
      setDateError("End date cannot be before start date");
      return false;
    }
    setDateError("");
    return true;
  };

  const applyDateRange = () => {
    if (!validateDates()) return;
    onChange(tempStart, tempEnd);
    setShowDatePopup(false);
  };

  const clearDateSelection = () => {
    setTempStart("");
    setTempEnd("");
    setDatePreset(null);
    setDateError("");
  };

  const dateRangeLabel = () => {
    if (startDate && endDate) return `${startDate} → ${endDate}`;
    if (startDate) return `From ${startDate}`;
    if (endDate) return `Until ${endDate}`;
    return triggerLabel;
  };

  return (
    <div ref={datePopupRef} className="relative flex flex-col">
      <button
        onClick={() => setShowDatePopup(!showDatePopup)}
        className={`flex items-center gap-3 bg-surface-container border text-xs font-bold rounded-xl px-4 py-2.5 hover:bg-surface-container-highest transition-all min-w-[220px] justify-between shadow-sm outline-none ${
          startDate || endDate
            ? "border-primary text-primary"
            : "border-outline text-on-surface-variant"
        }`}
      >
        <span className="flex items-center gap-2">
          <span className="material-symbols-outlined text-[18px]">calendar_today</span>
          <span className="truncate">{dateRangeLabel()}</span>
        </span>
        {(startDate || endDate) && (
          <button
            onClick={(e) => {
              e.stopPropagation();
              onChange("", "");
            }}
            className="text-on-surface-variant hover:text-white transition-colors"
          >
            <span className="material-symbols-outlined text-sm font-black">close</span>
          </button>
        )}
      </button>

      {showDatePopup && (
        <div
          className={`absolute top-[calc(100%+8px)] bg-surface-container border border-outline rounded-2xl shadow-2xl z-50 w-[670px] overflow-hidden ${
            align === "right" ? "right-0" : "left-0"
          }`}
        >
          <div className="px-5 py-4 border-b border-outline bg-surface-dim">
            <h4 className="text-white text-[11px] font-black uppercase tracking-[0.15em]">Timeframe Selection</h4>
          </div>

          <div className="flex">
            {/* Sidebar Presets */}
            <div className="w-[160px] bg-surface-dim border-r border-outline py-3 flex-shrink-0">
              {PRESETS.map((p) => (
                <button
                  key={p.key}
                  onClick={() => applyPreset(p.key)}
                  className={`w-full text-left px-4 py-3.5 text-[10px] font-black uppercase tracking-widest flex items-center gap-3 transition-all ${
                    datePreset === p.key
                      ? "bg-primary text-white"
                      : "text-on-surface-variant hover:bg-surface-container hover:text-white"
                  }`}
                >
                  <span className="material-symbols-outlined text-[18px]">{p.icon}</span>
                  <span className="truncate">{p.label}</span>
                </button>
              ))}
            </div>

            {/* Custom Dual-Calendars Range */}
            <div className="flex-1 p-5 overflow-hidden">
              <div className="flex items-center justify-between mb-4 bg-surface-darker/60 rounded-xl p-3 border border-outline/15">
                <div className="flex flex-col">
                  <span className="text-[7px] font-black text-on-surface-variant uppercase tracking-widest">Start Date</span>
                  <span className="text-[11px] font-mono font-bold text-white">{tempStart || "Not Selected"}</span>
                </div>
                <span className="material-symbols-outlined text-[16px] text-primary">arrow_forward</span>
                <div className="flex flex-col text-right">
                  <span className="text-[7px] font-black text-on-surface-variant uppercase tracking-widest">End Date</span>
                  <span className="text-[11px] font-mono font-bold text-white">{tempEnd || "Not Selected"}</span>
                </div>
              </div>

              <div className="flex gap-4 justify-center">
                <MiniCalendar
                  label="Start Date"
                  value={tempStart}
                  onChange={(date) => {
                    setTempStart(date);
                    setDatePreset("custom");
                    setDateError("");
                  }}
                  view={startView}
                  setView={setStartView}
                  highlightRange={{ start: tempStart, end: tempEnd }}
                />
                <MiniCalendar
                  label="End Date"
                  value={tempEnd}
                  onChange={(date) => {
                    setTempEnd(date);
                    setDatePreset("custom");
                    setDateError("");
                  }}
                  view={endView}
                  setView={setEndView}
                  highlightRange={{ start: tempStart, end: tempEnd }}
                />
              </div>

              {dateError && (
                <p className="text-[9px] text-rose-400 font-bold flex items-center gap-1.5 mt-3 justify-center">
                  <span className="material-symbols-outlined text-[12px]">error</span>
                  {dateError}
                </p>
              )}
            </div>
          </div>

          {/* Footer */}
          <div className="border-t border-outline px-5 py-4 flex items-center justify-between bg-surface-dim">
            <button
              onClick={() => {
                clearDateSelection();
                onChange("", "");
              }}
              className="text-[10px] font-black uppercase tracking-widest text-on-surface-variant hover:text-white transition-all px-4 py-2 hover:bg-surface-container rounded-xl"
            >
              Clear Selection
            </button>
            <button
              onClick={applyDateRange}
              className="bg-primary hover:bg-primary-dark text-white text-[10px] font-black uppercase tracking-widest rounded-xl px-6 py-2.5 transition-all shadow-lg shadow-primary/20 active:scale-95"
            >
              Apply Filter
            </button>
          </div>
        </div>
      )}
    </div>
  );
};
