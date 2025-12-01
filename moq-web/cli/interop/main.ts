import { parseArgs } from "@std/cli/parse-args";
import { Client } from "../../src/client.ts";
import { TrackMux } from "../../src/track_mux.ts";
import { Frame } from "../../src/frame.ts";
import { GroupSequenceFirst } from "../../src/group_stream.ts";

async function main() {
  const args = parseArgs(Deno.args, {
    string: ["addr", "cert-hash"],
    boolean: ["insecure", "debug"],
    default: { addr: "https://localhost:9000", insecure: false, debug: false },
  });

  // Suppress debug logs unless --debug flag is provided
  if (!args.debug) {
    console.debug = () => {};
  }

  const addr = args.addr;
  console.log(`Connecting to server at ${addr}...`);

  // For local development with self-signed certificates
  let transportOptions: WebTransportOptions = {};

  if (args.insecure && args["cert-hash"]) {
    // Use provided certificate hash for localhost development
    const hashBase64 = args["cert-hash"];
    const hashBytes = Uint8Array.from(atob(hashBase64), (c) => c.charCodeAt(0));

    transportOptions = {
      serverCertificateHashes: [{
        algorithm: "sha-256",
        value: hashBytes,
      }],
    };
    console.log("[DEV] Using provided certificate hash for self-signed cert");
  }

  const client = new Client({ transportOptions });
  const mux = new TrackMux();

  // Register publish handler before dialing
  const donePromise = new Promise<void>((resolve) => {
    // Create a context promise that never resolves (unless we want to stop publishing)
    const ctx = new Promise<void>(() => {});

    mux.publishFunc(ctx, "/interop/client", async (tw) => {
      try {
        console.log("Opening group...");
        const [group, err] = await tw.openGroup(GroupSequenceFirst);
        if (err) throw err;
        console.log("...ok");

        console.log("Writing frame to server...");
        const frame = new Frame(new Uint8Array([72, 69, 76, 76, 79])); // "HELLO"
        const writeErr = await group.writeFrame(frame);
        if (writeErr) throw writeErr;
        console.log("...ok");

        await group.close();
        resolve();
      } catch (err) {
        console.error("Error in publish handler:", err);
        resolve(); // Resolve anyway to avoid hanging
      }
    });
  });

  let sess;
  try {
    sess = await client.dial(addr, mux);
    console.log("...ok");
  } catch (err) {
    console.error("...failed\n  Error:", err);
    return;
  }

  try {
    // Step 1: Accept announcements from server
    console.log("Accepting server announcements...");
    const [anns, acceptErr] = await sess.acceptAnnounce("/");
    if (acceptErr) throw acceptErr;
    console.log("...ok");

    console.log("Receiving announcement...");
    // Use a never-resolving promise for signal if we don't want timeout
    const [ann, annErr] = await anns.receive(new Promise(() => {}));
    if (annErr) throw annErr;
    if (!ann) {
      throw new Error("Announcement stream closed");
    }
    console.log("...ok");
    console.log(`Discovered broadcast: ${ann.broadcastPath}`);

    // Step 2: Subscribe to the server's broadcast and receive data
    console.log("Subscribing to server broadcast...");
    const [track, subErr] = await sess.subscribe(ann.broadcastPath, "");
    if (subErr) throw subErr;
    console.log("...ok");

    console.log("Accepting group from server...");
    const [group, groupErr] = await track.acceptGroup(new Promise(() => {}));
    if (groupErr) throw groupErr;
    if (!group) {
      throw new Error("Track closed before group received");
    }
    console.log("...ok");

    console.log("Reading the first frame from server...");
    const frame = new Frame(new Uint8Array(1024));
    const readErr = await group.readFrame(frame);
    if (readErr) throw readErr;

    // Note: frame.data might contain trailing zeros if payload is smaller than 1024
    const payload = new TextDecoder().decode(frame.data).replace(/\0/g, "");
    console.log(`...ok (payload: ${payload})`);

    // Wait for the publish handler to complete
    await donePromise;

    // Wait a bit for server to process everything before closing
    await new Promise((r) => setTimeout(r, 1000));
  } catch (err) {
    console.error("Error during interop:", err);
  } finally {
    console.log("Closing session...");
    await sess.closeWithError(0, "no error");
    console.log("...ok");
  }
}

if (import.meta.main) {
  main().catch((err) => {
    console.error("Fatal error:", err);
    Deno.exit(1);
  });
}
