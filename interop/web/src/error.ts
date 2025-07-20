export type SessionErrorCode = number;

export type AnnounceErrorCode = number;

export type SubscribeErrorCode = number;
export const TrackNotFoundErrorCode: SubscribeErrorCode = 2;

export type GroupErrorCode = number;
export const PublishAbortedErrorCode: GroupErrorCode = 1;