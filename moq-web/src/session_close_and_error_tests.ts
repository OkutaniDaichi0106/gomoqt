import { assertEquals } from "@std/assert";
import { Session } from "./session.ts";
import { SessionServerMessage, writeVarint } from "./internal/message/mod.ts";
import { encodeMessageToUint8Array, MockWebTransport } from "./testing/mock_webtransport.ts";
import { DEFAULT_CLIENT_VERSIONS } from "./version.ts";

Deno.test("Session - close calls transport.close", async () => {
	let closedCalled = false;
	class LocalMock extends MockWebTransport {
		close(_closeInfo?: WebTransportCloseInfo) {
			closedCalled = true;
			super.close();
		}
	}

	const serverRsp = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
	const serverBytes = await encodeMessageToUint8Array(async (w) => serverRsp.encode(w));
	const mock = new LocalMock([serverBytes]);

	const session = new Session({ conn: (mock as any) as WebTransport });

	await session.ready;
	await session.close();
	assertEquals(closedCalled, true);
});

Deno.test("Session - acceptAnnounce returns error when openStream fails", async () => {
	// Create a mock that returns a proper session stream on first createBidirectionalStream call,
	// and fails on subsequent createBidirectionalStream calls used by acceptAnnounce.
	let count = 0;
	class SeqMock extends MockWebTransport {
		async createBidirectionalStream() {
			count++;
			if (count === 1) {
				return await super.createBidirectionalStream();
			}
			const err = { source: "session" as const } as any;
			throw err;
		}
	}

	const serverRsp = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
	const serverBytes = await encodeMessageToUint8Array(async (w) => serverRsp.encode(w));
	const mock = new SeqMock([serverBytes]);
	const session = new Session({ conn: (mock as any) as WebTransport });
	await session.ready;
	const [reader, err] = await session.acceptAnnounce("/test/" as any);
	assertEquals(reader, undefined);
	assertEquals(err instanceof Error, true);
});
