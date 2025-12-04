import { Client, Frame, TrackMux, TrackWriter } from "@okudai/moq";
import { background } from "@okudai/golikejs/context";
import { scope } from "@okudai/golikejs";

scope(async (defer) => {
	const client = new Client();
	defer(() => {
		client.close();
	});

	const mux = new TrackMux();

	// Channel to signal that the publish handler has completed
	const doneCh = new Array<() => void>();
	let done = false;

	mux.publishFunc(
		background().done(),
		"/interop/client",
		async (track: TrackWriter) => {
			try {
				console.debug("[Client] Server subscribed, sending data...");

				const [group, trackErr] = await track.openGroup();
				if (trackErr) {
					console.error("[Client] Failed to open group:", trackErr);
					return;
				}
				defer(() => group.close());

				const frame = new Frame(
					new TextEncoder().encode("Hello from moq-ts client"),
				);

				const groupErr = await group.writeFrame(frame);
				if (groupErr) {
					console.error("[Client] Failed to write frame:", groupErr);
					return;
				}

				console.info("[Client] [OK] Data sent to server");
			} catch (e) {
				console.error("[Client] Error in publish:", e);
			} finally {
				// Signal that handler has been invoked
				done = true;
				doneCh.forEach((resolve) => resolve());
			}
		},
	);

	console.debug("[Client] Registering /interop/client handler");

	const session = await client.dial("http://127.0.0.1:9000/", mux);

	console.debug("[Client] Connected to server");

	const [announced, annReqErr] = await session.acceptAnnounce("/");
	if (annReqErr) {
		console.error("[Client] Failed to accept announce:", annReqErr);
		return;
	}

	console.debug("[Client] Starting to accept server announcements...");

	const [announcement, annErr] = await announced.receive(background().done());
	if (annErr) {
		console.error("[Client] Failed to receive announcement:", annErr);
		console.error("[Client] Error details:", annErr);
		return;
	}

	console.debug("[Client] Discovered broadcast:", announcement.broadcastPath);

	const [track, subErr] = await session.subscribe(
		announcement.broadcastPath,
		"",
	);
	if (subErr) {
		console.error("[Client] Failed to subscribe to track:", subErr);
		return;
	}

	console.debug("[Client] Subscribed to a track");

	const [group, groupErr] = await track.acceptGroup(background().done());
	if (groupErr) {
		console.error("[Client] Failed to accept group:", groupErr);
		return;
	}

	console.debug("[Client] Received a group");

	const frame = new Frame(new Uint8Array(1024));
	const readErr = await group.readFrame(frame);
	if (readErr) {
		console.error("[Client] Failed to read frame:", readErr);
		return;
	}

	console.log("[Client] Frame data length:", frame.data.byteLength);
	console.info(
		"[Client] [OK] Received data from server:",
		new TextDecoder().decode(frame.data),
	);

	console.debug("[Client] Operations completed");

	// Wait for the handler to complete (like Go's doneCh)
	if (!done) {
		await Promise.race([
			new Promise<void>((resolve) => doneCh.push(resolve)),
			new Promise<void>((resolve) => setTimeout(() => resolve(), 5000)),
		]);
	}

	// Wait for a longer time before closing to allow server to read the frame
	await new Promise((resolve) => setTimeout(resolve, 2000));

	console.debug("[Client] Closing session...");
	await session.closeWithError(0, "no error");
	console.debug("[Client] ...ok");

	defer(() => {
		Deno.exit(0);
	});
});

