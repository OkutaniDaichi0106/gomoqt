import type { Extensions } from './internal/extensions.ts';

export interface MOQOptions {
	versions?: Set<bigint>;
	extensions?: Extensions;
	reconnect?: boolean;
	// migrate?: (url: URL) => boolean;
	transportOptions?: WebTransportOptions;
}