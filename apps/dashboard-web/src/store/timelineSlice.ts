import { createSlice, PayloadAction } from "@reduxjs/toolkit";

export interface TimelineEvent {
  id: string;
  type: "trip_started" | "alert" | "delivery_complete" | "notification";
  message: string;
  timestamp: string;
  severity: "info" | "warning" | "critical";
  actor?: string;
  metadata?: any;
}

interface TimelineState {
  events: TimelineEvent[];
  stats: any;
  loading: boolean;
  error: string | null;
}

const initialState: TimelineState = {
  events: [],
  stats: null,
  loading: false,
  error: null,
};

const timelineSlice = createSlice({
  name: "timeline",
  initialState,
  reducers: {
    setTimelineEvents: (state, action: PayloadAction<TimelineEvent[]>) => {
      state.events = action.payload;
    },
    addTimelineEvent: (state, action: PayloadAction<TimelineEvent>) => {
      state.events.unshift(action.payload);
      if (state.events.length > 1000) {
        state.events.pop();
      }
    },
    setTimelineLoading: (state, action: PayloadAction<boolean>) => {
      state.loading = action.payload;
    },
    setTimelineError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload;
    },
  },
});

export const {
  setTimelineEvents,
  addTimelineEvent,
  setTimelineLoading,
  setTimelineError,
} = timelineSlice.actions;

export default timelineSlice.reducer;
