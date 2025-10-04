export * from "./room";

import { defineRoom } from "./room";

export function defineAll(): void {
    defineRoom();

    // Add more element definitions here as needed
}