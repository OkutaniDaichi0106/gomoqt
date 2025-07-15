import { Extensions } from './internal/extensions';

export interface MOQOptions {
	extensions?: Extensions;
	reconnect?: boolean;
	migrate?: (url: URL) => boolean;
	transport?: WebTransportOptions;
}