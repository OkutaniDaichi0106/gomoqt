import { describe, it, expect } from '@jest/globals';
import {
    GroupPeriod,
    GROUP_PERIOD_MILLISECOND,
    GROUP_PERIOD_SECOND,
    GROUP_PERIOD_MINUTE,
    GROUP_PERIOD_HOUR,
} from './group_period';

describe('GroupPeriod', () => {
    it('exports numeric constants', () => {
        expect(typeof GROUP_PERIOD_MILLISECOND).toBe('number');
        expect(typeof GROUP_PERIOD_SECOND).toBe('number');
        expect(typeof GROUP_PERIOD_MINUTE).toBe('number');
        expect(typeof GROUP_PERIOD_HOUR).toBe('number');
    });

    it('has correct values and relations', () => {
        expect(GROUP_PERIOD_MILLISECOND).toBe(1);
        expect(GROUP_PERIOD_SECOND).toBe(1000);
        expect(GROUP_PERIOD_MINUTE).toBe(60 * GROUP_PERIOD_SECOND);
        expect(GROUP_PERIOD_HOUR).toBe(60 * GROUP_PERIOD_MINUTE);

        // sanity checks
        expect(GROUP_PERIOD_MINUTE).toBe(60 * 1000);
        expect(GROUP_PERIOD_HOUR).toBe(60 * 60 * 1000);
    });

    it('GroupPeriod type is compatible with number at runtime', () => {
        const p: GroupPeriod = GROUP_PERIOD_SECOND;
        // runtime check: typeof is number
        expect(typeof p).toBe('number');
    });
});
