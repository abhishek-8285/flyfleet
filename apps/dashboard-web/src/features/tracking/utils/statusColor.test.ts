// utils/statusColor.test.ts

import { statusColor, statusLabel } from './statusColor';

describe('statusColor', () => {
    it('should return green for moving', () => {
        expect(statusColor('moving')).toBe('#22c55e');
    });

    it('should return amber for idle', () => {
        expect(statusColor('idle')).toBe('#f59e0b');
    });

    it('should return red for stopped', () => {
        expect(statusColor('stopped')).toBe('#ef4444');
    });

    it('should return gray for offline', () => {
        expect(statusColor('offline')).toBe('#6b7280');
    });

    it('should return gray for unknown status', () => {
        expect(statusColor('unknown' as any)).toBe('#6b7280');
    });
});

describe('statusLabel', () => {
    it('should capitalize first letter', () => {
        expect(statusLabel('moving')).toBe('Moving');
    });

    it('should handle all statuses', () => {
        expect(statusLabel('idle')).toBe('Idle');
        expect(statusLabel('stopped')).toBe('Stopped');
        expect(statusLabel('offline')).toBe('Offline');
    });
});
