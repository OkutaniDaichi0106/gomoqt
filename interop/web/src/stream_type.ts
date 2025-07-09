const BiStreamTypes: Record<string, number> = {
    SessionStreamType: 0x00,
    AnnounceStreamType: 0x01,
    SubscribeStreamType: 0x02,
} as const;

const UniStreamTypes: Record<string, number> = {
    GroupStreamType: 0x00,
} as const;