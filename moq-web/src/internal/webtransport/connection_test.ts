import { assertEquals, assertInstanceOf } from "@std/assert";
import { WebTransportSession } from "./connection.ts";
import { WebTransportSessionError } from "./error.ts";

class FailingMockWebTransport {
  ready = Promise.resolve();
  closed = Promise.resolve({ closeCode: 123, reason: "fail" });
  incomingBidirectionalStreams = new ReadableStream({ start(_c) {} });
  incomingUnidirectionalStreams = new ReadableStream({ start(_c) {} });
  async createBidirectionalStream() {
    const err = { source: "session" as const } as any; // not an Error
    throw err;
  }
}

Deno.test("SessionImpl.openStream maps WebTransportError session source to SessionError", async () => {
  const session = new WebTransportSession(
    (new FailingMockWebTransport()) as any,
  );
  const [stream, err] = await session.openStream();
  assertEquals(stream, undefined);
  assertInstanceOf(err, WebTransportSessionError);
});

Deno.test("SessionImpl.acceptStream returns error when incoming reader yields done", async () => {
  const mock = {
    incomingBidirectionalStreams: new ReadableStream({
      start(controller) {
        controller.close();
      },
    }),
    incomingUnidirectionalStreams: new ReadableStream({
      start(controller) {
        controller.close();
      },
    }),
    ready: Promise.resolve(),
    closed: Promise.resolve({ closeCode: undefined, reason: undefined }),
  } as any;

  const session = new WebTransportSession(mock);

  const [s1, e1] = await session.acceptStream();
  assertEquals(s1, undefined);
  assertInstanceOf(e1, Error);

  const [s2, e2] = await session.acceptUniStream();
  assertEquals(s2, undefined);
  assertInstanceOf(e2, Error);
});
