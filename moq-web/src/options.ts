import type { Extensions } from './internal/extensions';

export interface MOQOptions {
	versions?: Set<bigint>;
	extensions?: Extensions;
	reconnect?: boolean;
	// migrate?: (url: URL) => boolean;
	transportOptions?: WebTransportOptions;
}