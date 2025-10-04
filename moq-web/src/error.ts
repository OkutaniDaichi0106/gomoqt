export type SessionErrorCode = number;

export type AnnounceErrorCode = number;
export const InternalAnnounceErrorCode: AnnounceErrorCode = 0x00;
export const DuplicatedAnnounceErrorCode: AnnounceErrorCode = 0x1;
export const InvalidAnnounceStatusErrorCode: AnnounceErrorCode = 0x2; // TODO: Is this necessary?
export const UninterestedErrorCode: AnnounceErrorCode = 0x3;
export const BannedPrefixErrorCode: AnnounceErrorCode = 0x4; // TODO: Is this necessary?
export const InvalidPrefixErrorCode: AnnounceErrorCode = 0x5; // TODO: Is this necessary?


export type SubscribeErrorCode = number;

export const InternalSubscribeErrorCode: SubscribeErrorCode = 0x00;
export const InvalidRangeErrorCode: SubscribeErrorCode = 0x01;
export const DuplicateSubscribeIDErrorCode: SubscribeErrorCode = 0x02;
export const TrackNotFoundErrorCode: SubscribeErrorCode = 0x03;
export const UnauthorizedSubscribeErrorCode: SubscribeErrorCode = 0x04; // TODO: Is this necessary?
export const SubscribeTimeoutErrorCode: SubscribeErrorCode = 0x05;



export type GroupErrorCode = number;
export const InternalGroupErrorCode: GroupErrorCode = 0x00;

export const OutOfRangeErrorCode: GroupErrorCode = 0x02;
export const ExpiredGroupErrorCode: GroupErrorCode = 0x03;
export const SubscribeCanceledErrorCode: GroupErrorCode = 0x04; // TODO: Is this necessary?
export const PublishAbortedErrorCode: GroupErrorCode = 0x05;
export const ClosedSessionGroupErrorCode: GroupErrorCode = 0x06;
export const InvalidSubscribeIDErrorCode: GroupErrorCode = 0x07; // TODO: Is this necessary?