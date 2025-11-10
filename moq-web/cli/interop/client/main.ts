import { Frame, Client, TrackMux } from "@okudai/moq";
import { background } from "@okudai/golikejs/context";
import { scope } from "@okudai/golikejs";

scope(async (defer)=> {
    defer(() => {
        Deno.exit(0)
    })

    const client = new Client()
    defer(() => {
        client.close()
    })

    const mux = new TrackMux()

    mux.publishFunc(background().done(), "/interop/client", async (track) => {
        scope(async (defer) => {
            const [group, trackErr] = await track.openGroup(1)
            if (trackErr) {
                console.error("[Client] Failed to open group:", trackErr)
                return
            }
            defer(() => group.close())

            console.log("[Client] Opened group")

            const frame = new Frame(new TextEncoder().encode("Hello from moq-ts client"))

            const groupErr = await group.writeFrame(frame)
            if (groupErr) {
                console.error("[Client] Failed to write frame:", groupErr)
                return
            }

            return
        })
    })

    const session = await client.dial("http://moqt.example.com:9000/", mux)

    const [announced, annReqErr] = await session.acceptAnnounce("/")
    if (annReqErr) {
        console.error("[Client] Failed to accept announce:", annReqErr)
        return
    }

    const [announcement, annErr] = await announced.receive(background().done())
    if (annErr) {
        console.error("[Client] Failed to receive announcement:", annErr)
        return
    }

    console.log("[Client] Received announcement for path:", announcement.broadcastPath)

    const [track, subErr] = await session.subscribe(announcement.broadcastPath, "")
    if (subErr) {
        console.error("[Client] Failed to subscribe to track:", subErr)
        return
    }

    const [group, groupErr] = await track.acceptGroup(background().done())
    if (groupErr) {
        console.error("[Client] Failed to accept group:", groupErr)
        return
    }
    console.log("[Client] Accepted group")

    const frame = new Frame(new Uint8Array())
    const readErr = await group.readFrame(frame)
    if (readErr) {
        console.error("[Client] Failed to read frame:", readErr)
        return
    }
})
