export type StreamErrorCode = number;

export class StreamError extends Error {
	readonly code: StreamErrorCode;
	readonly remote: boolean;

	constructor(code: StreamErrorCode, message: string, remote: boolean = false) {
		super(message);
		this.code = code;
		this.remote = remote;
		Object.setPrototypeOf(this, StreamError.prototype);
	}
}
