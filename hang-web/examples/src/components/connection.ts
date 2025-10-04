import WebTransportWs from "@kixelated/web-transport-ws";
import { withCancelCause, background, watchSignal } from "golikejs/context";

export interface TransportOptions extends WebTransportOptions {
    fallback?: boolean;
    fallbackDelay?: number;
    fallbackUrl?: string | URL;
    signal?: AbortSignal;
}

export async function connect(
    url: string | URL,
    options?: TransportOptions | undefined,
): Promise<WebTransport> {
    const parentCtx = options?.signal ? watchSignal(background(), options.signal) : background();
    const [ctx, cancel] = withCancelCause(parentCtx);

    //
    const webtransport = globalThis.WebTransport ? new WebTransport(url, options) : undefined;

    const attemptTimeout = !webtransport ? 0 : options?.fallbackDelay || 200; // 2 milliseconds timeout for WebTransport connection attempt
    const timer = new Promise<void>((resolve) => setTimeout(resolve, attemptTimeout));

    // Wait for either WebTransport to be ready, timeout, or context cancellation
    // If WebTransport is not supported, this will immediately resolve due to 0 timeout
    let succeed = await Promise.race([
        webtransport?.ready ? webtransport.ready.then(() => true) : Promise.resolve(false),
        timer.then(() => false),
        ctx.done()
    ]);

    if (succeed === true) {
        // WebTransport connected successfully within the timeout
        cancel(undefined); // Cancel the context to clean up any listeners
        return webtransport!;
    }

    const websocket = (globalThis.WebSocket || options?.fallback) ? new WebTransportWs(options?.fallbackUrl || url, options) : undefined;

    if (!webtransport && !websocket) {
        throw new Error("Neither WebTransport nor WebSocket is supported in this environment.");
    }

    succeed = await Promise.race([
        websocket?.ready ? websocket.ready.then(() => true) : Promise.resolve(false),
        ctx.done()
    ]);

    if (succeed === true) {
        // WebSocket connected successfully within the timeout
        cancel(undefined); // Cancel the context to clean up any listeners
        return websocket!;
    }

    // If we reach here, it means neither transport could connect
    const error = ctx.err() || new Error("Connection attempt was aborted or failed.");
    cancel(error);
    throw error;
}