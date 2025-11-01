import { assertEquals } from "@std/assert";
import { AnnounceInitMessage } from "./announce_init.ts";
import { createIsolatedStreams } from "./test-utils_deno_test.ts";

Deno.test("AnnounceInitMessage", async (t) => {
	await t.step("should encode and decode with empty suffixes array", async () => {
		const suffixes: string[] = [];
		const { writer, reader, cleanup } = createIsolatedStreams();

		try {
			// Encode the message
			const message = new AnnounceInitMessage({ suffixes });
			const encodeErr = await message.encode(writer);
			assertEquals(encodeErr, undefined);

			// Close writer to signal end of stream
			await writer.close();

			// Decode the message
			const decodedMessage = new AnnounceInitMessage({});
			const decodeErr = await decodedMessage.decode(reader);
			assertEquals(decodeErr, undefined);
			assertEquals(decodedMessage.suffixes, suffixes);
		} finally {
			await cleanup();
		}
	});

	await t.step("should encode and decode with single suffix", async () => {
		const suffixes = ["test-suffix"];
		const { writer, reader, cleanup } = createIsolatedStreams();

		try {
			// Encode the message
			const message = new AnnounceInitMessage({ suffixes });
			const encodeErr = await message.encode(writer);
			assertEquals(encodeErr, undefined);

			// Close writer to signal end of stream
			await writer.close();

			// Decode the message
			const decodedMessage = new AnnounceInitMessage({});
			const decodeErr = await decodedMessage.decode(reader);
			assertEquals(decodeErr, undefined);
			assertEquals(decodedMessage.suffixes, suffixes);
		} finally {
			await cleanup();
		}
	});

	await t.step("should encode and decode with multiple suffixes", async () => {
		const suffixes = ["suffix1", "suffix2", "suffix3"];
		const { writer, reader, cleanup } = createIsolatedStreams();

		try {
			// Encode the message
			const message = new AnnounceInitMessage({ suffixes });
			const encodeErr = await message.encode(writer);
			assertEquals(encodeErr, undefined);

			// Close writer to signal end of stream
			await writer.close();

			// Decode the message
			const decodedMessage = new AnnounceInitMessage({});
			const decodeErr = await decodedMessage.decode(reader);
			assertEquals(decodeErr, undefined);
			assertEquals(decodedMessage.suffixes, suffixes);
		} finally {
			await cleanup();
		}
	});

	await t.step("should handle special characters in suffixes", async () => {
		const suffixes = [
			"suffix-with-dashes",
			"suffix_with_underscores",
			"suffix/with/slashes",
			"suffix with spaces",
		];
		const { writer, reader, cleanup } = createIsolatedStreams();

		try {
			// Encode the message
			const message = new AnnounceInitMessage({ suffixes });
			const encodeErr = await message.encode(writer);
			assertEquals(encodeErr, undefined);

			// Close writer to signal end of stream
			await writer.close();

			// Decode the message
			const decodedMessage = new AnnounceInitMessage({});
			const decodeErr = await decodedMessage.decode(reader);
			assertEquals(decodeErr, undefined);
			assertEquals(decodedMessage.suffixes, suffixes);
		} finally {
			await cleanup();
		}
	});

	await t.step("should create instance with constructor", () => {
		const suffixes = ["test1", "test2"];
		const message = new AnnounceInitMessage({ suffixes });

		assertEquals(message.suffixes, suffixes);
	});

	await t.step("should handle empty strings in suffixes array", async () => {
		const suffixes = ["", "valid-suffix", ""];
		const { writer, reader, cleanup } = createIsolatedStreams();

		try {
			// Encode the message
			const message = new AnnounceInitMessage({ suffixes });
			const encodeErr = await message.encode(writer);
			assertEquals(encodeErr, undefined);

			// Close writer to signal end of stream
			await writer.close();

			// Decode the message
			const decodedMessage = new AnnounceInitMessage({});
			const decodeErr = await decodedMessage.decode(reader);
			assertEquals(decodeErr, undefined);
			assertEquals(decodedMessage.suffixes, suffixes);
		} finally {
			await cleanup();
		}
	});
});
