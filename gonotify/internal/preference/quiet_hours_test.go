package preference

import (
	"testing"
	"time"
)

func parseTZ(t *testing.T, tz string) *time.Location {
	t.Helper()

	if tz == "" {
		return time.UTC
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		t.Fatalf("invalid timezone %q: %v", tz, err)
	}

	return loc
}

func at(t *testing.T, ts string) time.Time {
	t.Helper()

	loc := time.UTC
	tt, err := time.ParseInLocation("2006-01-02T15:04:05", ts, loc)
	if err != nil {
		t.Fatalf("parse time %q: %v", ts, err)
	}

	return tt
}

func atTZ(t *testing.T, ts, tz string) time.Time {
	t.Helper()

	loc := parseTZ(t, tz)
	tt, err := time.ParseInLocation("2006-01-02T15:04:05", ts, loc)
	if err != nil {
		t.Fatalf("parse time %q: %v", ts, err)
	}

	return tt
}

// --- disabled ---

func TestQuietHours_IsActive_Disabled(t *testing.T) {
	qh := QuietHours{Enabled: false, StartTime: "22:00", EndTime: "08:00", Timezone: "UTC"}

	now := at(t, "2026-08-01T03:00:00")
	if qh.IsActive(now) {
		t.Fatal("disabled quiet hours should never be active")
	}
}

// --- normal window (start < end), UTC ---

func TestQuietHours_IsActive_NormalWindow(t *testing.T) {
	qh := QuietHours{Enabled: true, StartTime: "08:00", EndTime: "20:00", Timezone: "UTC"}

	cases := []struct {
		ts   string
		want bool
	}{
		{"2026-08-01T07:59:00", false},
		{"2026-08-01T08:00:00", true},
		{"2026-08-01T12:00:00", true},
		{"2026-08-01T19:59:00", true},
		{"2026-08-01T20:00:00", false},
		{"2026-08-01T23:00:00", false},
	}

	for _, c := range cases {
		got := qh.IsActive(at(t, c.ts))
		if got != c.want {
			t.Errorf("IsActive(%s) = %v, want %v", c.ts, got, c.want)
		}
	}
}

// --- midnight-spanning window (start > end), UTC ---

func TestQuietHours_IsActive_MidnightSpanning(t *testing.T) {
	qh := QuietHours{Enabled: true, StartTime: "22:00", EndTime: "08:00", Timezone: "UTC"}

	cases := []struct {
		ts   string
		want bool
	}{
		{"2026-08-01T07:59:00", true},
		{"2026-08-01T08:00:00", false},
		{"2026-08-01T12:00:00", false},
		{"2026-08-01T21:59:00", false},
		{"2026-08-01T22:00:00", true},
		{"2026-08-01T23:30:00", true},
		{"2026-08-02T00:01:00", true},
		{"2026-08-02T07:59:00", true},
		{"2026-08-02T08:00:00", false},
	}

	for _, c := range cases {
		got := qh.IsActive(at(t, c.ts))
		if got != c.want {
			t.Errorf("IsActive(%s) = %v, want %v", c.ts, got, c.want)
		}
	}
}

// --- End() computation ---

func TestQuietHours_End_NormalWindow(t *testing.T) {
	qh := QuietHours{Enabled: true, StartTime: "08:00", EndTime: "20:00", Timezone: "UTC"}

	now := at(t, "2026-08-01T10:30:00")
	end := qh.End(now)

	want := at(t, "2026-08-01T20:00:00")
	if !end.Equal(want) {
		t.Errorf("End = %v, want %v", end, want)
	}
}

func TestQuietHours_End_MidnightSpanning_FirstSegment(t *testing.T) {
	qh := QuietHours{Enabled: true, StartTime: "22:00", EndTime: "08:00", Timezone: "UTC"}

	now := at(t, "2026-08-01T23:00:00") // in the "after start" segment
	end := qh.End(now)

	want := at(t, "2026-08-02T08:00:00")
	if !end.Equal(want) {
		t.Errorf("End = %v, want %v", end, want)
	}
}

func TestQuietHours_End_MidnightSpanning_SecondSegment(t *testing.T) {
	qh := QuietHours{Enabled: true, StartTime: "22:00", EndTime: "08:00", Timezone: "UTC"}

	now := at(t, "2026-08-01T02:00:00") // in the "before end" segment
	end := qh.End(now)

	want := at(t, "2026-08-01T08:00:00")
	if !end.Equal(want) {
		t.Errorf("End = %v, want %v", end, want)
	}
}

func TestQuietHours_End_RespectsTimezone(t *testing.T) {
	// start=22:00, end=08:00 in America/New_York. Now is 23:00 NY time
	// (which is 03:00 UTC the next day) — in the first segment.
	qh := QuietHours{Enabled: true, StartTime: "22:00", EndTime: "08:00", Timezone: "America/New_York"}

	now := atTZ(t, "2026-08-01T23:00:00", "America/New_York")
	end := qh.End(now)

	want := atTZ(t, "2026-08-02T08:00:00", "America/New_York")
	if !end.Equal(want.UTC()) {
		t.Errorf("End = %v (UTC), want %v", end, want.UTC())
	}

	if end.Location() != time.UTC {
		t.Errorf("End should return UTC time, got %v", end.Location())
	}
}

func TestQuietHours_End_DisabledReturnsNow(t *testing.T) {
	qh := QuietHours{Enabled: false, StartTime: "22:00", EndTime: "08:00", Timezone: "UTC"}

	now := at(t, "2026-08-01T03:00:00")
	end := qh.End(now)

	if !end.Equal(now.UTC()) {
		t.Errorf("End of disabled window should return now, got %v want %v", end, now.UTC())
	}
}

func TestQuietHours_IsActive_SpanningTimezone(t *testing.T) {
	// start=22:00, end=08:00 America/New_York (UTC-4 in August).
	// 23:00 NY = 03:00 UTC next day — should be active.
	qh := QuietHours{Enabled: true, StartTime: "22:00", EndTime: "08:00", Timezone: "America/New_York"}

	now := atTZ(t, "2026-08-01T23:00:00", "America/New_York")
	if !qh.IsActive(now) {
		t.Fatal("expected to be in quiet hours at 23:00 NY time")
	}

	// 07:00 NY = 11:00 UTC — should be active (before end).
	now2 := atTZ(t, "2026-08-02T07:00:00", "America/New_York")
	if !qh.IsActive(now2) {
		t.Fatal("expected to be in quiet hours at 07:00 NY time")
	}

	// 09:00 NY = 13:00 UTC — should NOT be active.
	now3 := atTZ(t, "2026-08-02T09:00:00", "America/New_York")
	if qh.IsActive(now3) {
		t.Fatal("expected NOT to be in quiet hours at 09:00 NY time")
	}
}

// --- validation ---

func TestQuietHours_Validate(t *testing.T) {
	t.Run("disabled is valid", func(t *testing.T) {
		qh := QuietHours{Enabled: false, StartTime: "garbage", EndTime: "also", Timezone: "Not/A/Zone"}
		if err := qh.Validate(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("valid", func(t *testing.T) {
		qh := QuietHours{Enabled: true, StartTime: "22:00", EndTime: "08:00", Timezone: "America/New_York"}
		if err := qh.Validate(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("bad start_time", func(t *testing.T) {
		qh := QuietHours{Enabled: true, StartTime: "25:00", EndTime: "08:00", Timezone: "UTC"}
		if err := qh.Validate(); err == nil {
			t.Fatal("expected error for invalid start_time")
		}
	})

	t.Run("bad end_time", func(t *testing.T) {
		qh := QuietHours{Enabled: true, StartTime: "22:00", EndTime: "99:00", Timezone: "UTC"}
		if err := qh.Validate(); err == nil {
			t.Fatal("expected error for invalid end_time")
		}
	})

	t.Run("bad timezone", func(t *testing.T) {
		qh := QuietHours{Enabled: true, StartTime: "22:00", EndTime: "08:00", Timezone: "Not/A/Zone"}
		if err := qh.Validate(); err == nil {
			t.Fatal("expected error for invalid timezone")
		}
	})
}

func TestDefaultQuietHours(t *testing.T) {
	d := DefaultQuietHours()
	if d.Enabled {
		t.Fatal("default quiet hours should be disabled")
	}

	if d.StartTime != "22:00" || d.EndTime != "08:00" || d.Timezone != "UTC" {
		t.Fatalf("unexpected defaults: %+v", d)
	}
}
