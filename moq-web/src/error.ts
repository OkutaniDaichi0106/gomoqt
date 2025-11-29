// =============================================================================
// Session Error Codes
// =============================================================================

import * as webtransport from "./internal/webtransport/error.ts";

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

export class SessionError extends webtransport.SessionError {
	override get code(): SessionErrorCode {
		return super.code as SessionErrorCode;
	}

	override get message(): string {
		return SessionError.textOf(this.code);
	}

	static textOf(code: SessionErrorCode): string {
		switch (code) {
			case SessionErrorCode.NoError:
				return "no error";
			case SessionErrorCode.InternalError:
				return "internal error";
			case SessionErrorCode.Unauthorized:
				return "unauthorized";
			case SessionErrorCode.ProtocolViolation:
				return "protocol violation";
			case SessionErrorCode.DuplicateTrackAlias:
				return "duplicate track alias";
			case SessionErrorCode.ParameterLengthMismatch:
				return "parameter length mismatch";
			case SessionErrorCode.TooManySubscribers:
				return "too many subscribers";
			case SessionErrorCode.GoAwayTimeout:
				return "goaway timeout";
			default:
				return `unknown session error (${code})`;
		}
	}

	constructor(code: SessionErrorCode, remote: boolean) {
		super({ closeCode: code, reason: SessionError.textOf(code) }, remote);
		this.name = "SessionError";
		Object.setPrototypeOf(this, SessionError.prototype);
		// Remove the instance message property from Error so that the prototype getter
		// `message` defined in this subclass is used instead.
		try {
			delete (this as any).message;
		} catch (_) { /* ignore */ }
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

export class AnnounceError extends webtransport.StreamError {
	override get code(): AnnounceErrorCode {
		return super.code as AnnounceErrorCode;
	}

	override get message(): string {
		return AnnounceError.textOf(this.code);
	}

	static textOf(code: AnnounceErrorCode): string {
		switch (code) {
			case AnnounceErrorCode.InternalError:
				return "internal error";
			case AnnounceErrorCode.DuplicatedAnnounce:
				return "duplicated announce";
			case AnnounceErrorCode.InvalidAnnounceStatus:
				return "invalid announce status";
			case AnnounceErrorCode.Uninterested:
				return "uninterested";
			case AnnounceErrorCode.BannedPrefix:
				return "banned prefix";
			case AnnounceErrorCode.InvalidPrefix:
				return "invalid prefix";
			default:
				return `unknown announce error (${code})`;
		}
	}

	constructor(code: AnnounceErrorCode, remote: boolean) {
		super({ source: "stream", streamErrorCode: code }, remote);
		this.name = "AnnounceError";
		Object.setPrototypeOf(this, AnnounceError.prototype);
		try {
			delete (this as any).message;
		} catch (_) { /* ignore */ }
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

export class SubscribeError extends webtransport.StreamError {
	override get code(): SubscribeErrorCode {
		return super.code as SubscribeErrorCode;
	}

	override get message(): string {
		return SubscribeError.textOf(this.code);
	}

	static textOf(code: SubscribeErrorCode): string {
		switch (code) {
			case SubscribeErrorCode.InternalError:
				return "internal error";
			case SubscribeErrorCode.InvalidRange:
				return "invalid range";
			case SubscribeErrorCode.DuplicateSubscribeID:
				return "duplicate subscribe id";
			case SubscribeErrorCode.TrackNotFound:
				return "track not found";
			case SubscribeErrorCode.Unauthorized:
				return "unauthorized";
			case SubscribeErrorCode.SubscribeTimeout:
				return "subscribe timeout";
			default:
				return `unknown subscribe error (${code})`;
		}
	}

	constructor(code: SubscribeErrorCode, remote: boolean) {
		super({ source: "stream", streamErrorCode: code }, remote);
		this.name = "SubscribeError";
		Object.setPrototypeOf(this, SubscribeError.prototype);
		try {
			delete (this as any).message;
		} catch (_) { /* ignore */ }
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

export class GroupError extends webtransport.StreamError {
	override get code(): GroupErrorCode {
		return super.code as GroupErrorCode;
	}

	override get message(): string {
		return GroupError.textOf(this.code);
	}

	static textOf(code: GroupErrorCode): string {
		switch (code) {
			case GroupErrorCode.InternalError:
				return "internal error";
			case GroupErrorCode.OutOfRange:
				return "out of range";
			case GroupErrorCode.ExpiredGroup:
				return "expired group";
			case GroupErrorCode.SubscribeCanceled:
				return "subscribe canceled";
			case GroupErrorCode.PublishAborted:
				return "publish aborted";
			case GroupErrorCode.ClosedSession:
				return "closed session";
			case GroupErrorCode.InvalidSubscribeID:
				return "invalid subscribe id";
			default:
				return `unknown group error (${code})`;
		}
	}

	constructor(code: GroupErrorCode, remote: boolean) {
		super({ source: "stream", streamErrorCode: code }, remote);
		this.name = "GroupError";
		Object.setPrototypeOf(this, GroupError.prototype);
		try {
			delete (this as any).message;
		} catch (_) { /* ignore */ }
	}
}
