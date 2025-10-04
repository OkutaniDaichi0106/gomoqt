export const BiStreamTypes = {
    SessionStreamType: 0x00,
    AnnounceStreamType: 0x01,
    SubscribeStreamType: 0x02,
} as const;

export const UniStreamTypes = {
    GroupStreamType: 0x00,
} as const;

export type BiStreamType = typeof BiStreamTypes[keyof typeof BiStreamTypes];
export type UniStreamType = typeof UniStreamTypes[keyof typeof UniStreamTypes];