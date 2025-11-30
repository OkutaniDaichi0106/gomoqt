// =============================================================================
// Session Error Codes
// =============================================================================

export const SessionErrorCode = {
	/** No error occurred */
	NoError: 0x0,
	/** Internal error */
	InternalError: 0x1,
	/** Unauthorized */
	Unauthorized: 0x2,
	/** Protocol violation */
	ProtocolViolation: 0x3,
	/** Duplicate track alias */
	DuplicateTrackAlias: 0x4,
	/** Parameter length mismatch */
	ParameterLengthMismatch: 0x5,
	/** Too many subscribers */
	TooManySubscribers: 0x6,
	/** GOAWAY timeout */
	GoAwayTimeout: 0x10,
} as const;

export type SessionErrorCode = number;

export class SessionError extends Error {
	readonly code: SessionErrorCode;

	constructor(code: SessionErrorCode, message: string) {
		super(message);
		this.name = "SessionError";
		this.code = code;
	}
}

// =============================================================================
// Announce Error Codes
// =============================================================================

export const AnnounceErrorCode = {
	/** Internal error */
	InternalError: 0x00,
	/** Duplicated announcement */
	DuplicatedAnnounce: 0x01,
	/** Invalid announce status */
	InvalidAnnounceStatus: 0x02,
	/** Uninterested */
	Uninterested: 0x03,
	/** Banned prefix */
	BannedPrefix: 0x04,
	/** Invalid prefix */
	InvalidPrefix: 0x05,
} as const;

export type AnnounceErrorCode = number;

export class AnnounceError extends Error {
	readonly code: AnnounceErrorCode;

	constructor(code: AnnounceErrorCode, message: string) {
		super(message);
		this.name = "AnnounceError";
		this.code = code;
	}
}

// =============================================================================
// Subscribe Error Codes
// =============================================================================

export const SubscribeErrorCode = {
	/** Internal error */
	InternalError: 0x00,
	/** Invalid range */
	InvalidRange: 0x01,
	/** Duplicate subscribe ID */
	DuplicateSubscribeID: 0x02,
	/** Track not found */
	TrackNotFound: 0x03,
	/** Unauthorized */
	Unauthorized: 0x04,
	/** Subscribe timeout */
	SubscribeTimeout: 0x05,
} as const;

export type SubscribeErrorCode = number;

export class SubscribeError extends Error {
	readonly code: SubscribeErrorCode;

	constructor(code: SubscribeErrorCode, message: string) {
		super(message);
		this.name = "SubscribeError";
		this.code = code;
	}
}

// =============================================================================
// Group Error Codes
// =============================================================================

export const GroupErrorCode = {
	/** Internal error */
	InternalError: 0x00,
	/** Out of range */
	OutOfRange: 0x02,
	/** Expired group */
	ExpiredGroup: 0x03,
	/** Subscribe canceled */
	SubscribeCanceled: 0x04,
	/** Publish aborted */
	PublishAborted: 0x05,
	/** Closed session */
	ClosedSession: 0x06,
	/** Invalid subscribe ID */
	InvalidSubscribeID: 0x07,
} as const;

export type GroupErrorCode = number;

export class GroupError extends Error {
	readonly code: GroupErrorCode;

	constructor(code: GroupErrorCode, message: string) {
		super(message);
		this.name = "GroupError";
		this.code = code;
	}
}
