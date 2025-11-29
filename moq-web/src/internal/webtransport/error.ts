export type SessionErrorCode = number;

export interface SessionErrorInfo {
	closeCode: number;
	reason: string;
}

export class SessionError extends Error {
	readonly _code: SessionErrorCode;

	readonly remote: boolean;

	readonly reason: string;

	get code(): SessionErrorCode {
		return this._code;
	}

	constructor(info: SessionErrorInfo, remote: boolean = false) {
		super(
			`session was closed for reason "${info.reason}" with code ${info.closeCode} by ${
				remote ? "remote" : "local"
			}`,
		);
		this._code = info.closeCode;
		this.remote = remote;
		this.reason = info.reason;
		Object.setPrototypeOf(this, SessionError.prototype);
	}
}

export type StreamErrorCode = number;

export interface StreamErrorInfo {
	readonly source: "stream";
	readonly streamErrorCode: number;
}

export class StreamError extends Error {
	readonly _code: StreamErrorCode;

	readonly remote: boolean;

	get code(): StreamErrorCode {
		return this._code;
	}

	constructor(init: StreamErrorInfo, remote: boolean) {
		super(`stream was reset with code ${init.streamErrorCode}`);
		this._code = init.streamErrorCode;
		this.remote = remote;
		Object.setPrototypeOf(this, StreamError.prototype);
	}

	toJSON(): Record<string, unknown> {
		// Detect direct circular reference shortcuts like `error.self = error` to preserve
		// the original behavior that JSON.stringify throws on circular references.
		for (const key of Object.keys(this)) {
			const v = (this as any)[key];
			if (v === this) {
				throw new TypeError("Converting circular structure to JSON");
			}
		}
		return {
			code: this._code,
			message: undefined,
			remote: this.remote,
		};
	}
}
