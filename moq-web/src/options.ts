import type { Extensions } from "./extensions.ts";
import { Version } from "./version.ts";

export interface MOQOptions {
	versions?: Set<Version>;
	extensions?: Extensions;
	reconnect?: boolean;
	// migrate?: (url: URL) => boolean;
	transportOptions?: WebTransportOptions;
}
