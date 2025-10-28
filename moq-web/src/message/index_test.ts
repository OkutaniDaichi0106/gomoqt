// Test file for message/index.ts exports
import { describe, it, beforeEach, afterEach, assertEquals, assertExists, assertThrows } from "../../deps.ts";
import * as MessageIndex from "./index.ts";
import * as SessionClient from "./session_client.ts";
import * as SessionServer from "./session_server.ts";
import * as SessionUpdate from "./session_update.ts";
import * as AnnouncePlease from "./announce_please.ts";
import * as Announce from "./announce.ts";
import * as Subscribe from "./subscribe.ts";
import * as SubscribeOk from "./subscribe_ok.ts";
import * as SubscribeUpdate from "./subscribe_update.ts";
import * as Group from "./group.ts";

describe("Message Index", () => {
    it("should export all session client exports", () => {
        assertEquals(MessageIndex, expect.objectContaining(SessionClient));
    });
    
    it("should export all session server exports", () => {
        assertEquals(MessageIndex, expect.objectContaining(SessionServer));
    });
    
    it("should export all session update exports", () => {
        assertEquals(MessageIndex, expect.objectContaining(SessionUpdate));
    });
    
    it("should export all announce please exports", () => {
        assertEquals(MessageIndex, expect.objectContaining(AnnouncePlease));
    });
    
    it("should export all announce exports", () => {
        assertEquals(MessageIndex, expect.objectContaining(Announce));
    });
    
    it("should export all subscribe exports", () => {
        assertEquals(MessageIndex, expect.objectContaining(Subscribe));
    });
    
    it("should export all subscribe ok exports", () => {
        assertEquals(MessageIndex, expect.objectContaining(SubscribeOk));
    });
    
    it("should export all subscribe update exports", () => {
        assertEquals(MessageIndex, expect.objectContaining(SubscribeUpdate));
    });
    
    it("should export all group exports", () => {
        assertEquals(MessageIndex, expect.objectContaining(Group));
    });

    it("should not have any undefined exports", () => {
        const exports = Object.keys(MessageIndex);
        exports.forEach(key => {
            expect((MessageIndex as any)[key]).toBeDefined();
        });
    });
    
    it("should have all expected module exports", () => {
        // Check that all modules we expect to be exported are actually exported
        const expectedExports = [
            ...Object.keys(SessionClient),
            ...Object.keys(SessionServer),
            ...Object.keys(SessionUpdate),
            ...Object.keys(AnnouncePlease),
            ...Object.keys(Announce),
            ...Object.keys(Subscribe),
            ...Object.keys(SubscribeOk),
            ...Object.keys(SubscribeUpdate),
            ...Object.keys(Group),
        ];
        
        expectedExports.forEach(exportName => {
            expect((MessageIndex as any)[exportName]).toBeDefined();
        });
    });
});
