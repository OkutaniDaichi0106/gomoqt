export * from "./reader.ts";
export * from "./writer.ts";
export * from "./stream.ts";
export * from "./error.ts";
export * from "./len.ts";
export * from "./bytes.ts";
export * from "./buffer_pool.ts";

// Type aliases for convenience
import type { ReceiveStream } from "./reader.ts";
import type { SendStream } from "./writer.ts";

export type Reader = ReceiveStream;
export type Writer = SendStream;
