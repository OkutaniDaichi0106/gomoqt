// Test file for message/index.ts exports
import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import * as MessageIndex from "./index";
import * as SessionClient from "./session_client";
import * as SessionServer from "./session_server";
import * as SessionUpdate from "./session_update";
import * as AnnouncePlease from "./announce_please";
import * as Announce from "./announce";
import * as Subscribe from "./subscribe";
import * as SubscribeOk from "./subscribe_ok";
import * as SubscribeUpdate from "./subscribe_update";
import * as Group from "./group";
import * as Frame from "./frame";

describe("Message Index", () => {
    it("should export all session client exports", () => {
        expect(MessageIndex).toEqual(expect.objectContaining(SessionClient));
    });
    
    it("should export all session server exports", () => {
        expect(MessageIndex).toEqual(expect.objectContaining(SessionServer));
    });
    
    it("should export all session update exports", () => {
        expect(MessageIndex).toEqual(expect.objectContaining(SessionUpdate));
    });
    
    it("should export all announce please exports", () => {
        expect(MessageIndex).toEqual(expect.objectContaining(AnnouncePlease));
    });
    
    it("should export all announce exports", () => {
        expect(MessageIndex).toEqual(expect.objectContaining(Announce));
    });
    
    it("should export all subscribe exports", () => {
        expect(MessageIndex).toEqual(expect.objectContaining(Subscribe));
    });
    
    it("should export all subscribe ok exports", () => {
        expect(MessageIndex).toEqual(expect.objectContaining(SubscribeOk));
    });
    
    it("should export all subscribe update exports", () => {
        expect(MessageIndex).toEqual(expect.objectContaining(SubscribeUpdate));
    });
    
    it("should export all group exports", () => {
        expect(MessageIndex).toEqual(expect.objectContaining(Group));
    });
    
    it("should export all frame exports", () => {
        expect(MessageIndex).toEqual(expect.objectContaining(Frame));
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
            ...Object.keys(Frame)
        ];
        
        expectedExports.forEach(exportName => {
            expect((MessageIndex as any)[exportName]).toBeDefined();
        });
    });
});
