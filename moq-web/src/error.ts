export type SessionErrorCode = number;

export type AnnounceErrorCode = number;

export type SubscribeErrorCode = number;
export const TrackNotFoundErrorCode: SubscribeErrorCode = 2;

export type GroupErrorCode = number;
export const InternalGroupErrorCode: GroupErrorCode = 0x00;

export const OutOfRangeErrorCode: GroupErrorCode = 0x02;
export const ExpiredGroupErrorCode: GroupErrorCode = 0x03;
export const SubscribeCanceledErrorCode: GroupErrorCode = 0x04; // TODO: Is this necessary?
export const PublishAbortedErrorCode: GroupErrorCode = 0x05;
export const ClosedSessionGroupErrorCode: GroupErrorCode = 0x06;
export const InvalidSubscribeIDErrorCode: GroupErrorCode = 0x07; // TODO: Is this necessary?