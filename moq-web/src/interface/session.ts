import * as moqt from "..";

export interface Session {
	readonly ready: Promise<void>;
	readonly mux: moqt.TrackMux;
	acceptAnnounce(
		prefix: moqt.TrackPrefix,
	): Promise<[moqt.AnnouncementReader, undefined] | [undefined, Error]>;
	subscribe(
		path: moqt.BroadcastPath,
		name: moqt.TrackName,
		config?: moqt.TrackConfig,
	): Promise<[moqt.TrackReader, undefined] | [undefined, Error]>;
	close: () => Promise<void>;
	closeWithError: (error: Error) => Promise<void>;
}
