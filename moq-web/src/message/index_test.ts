// Test file for message/index.ts exports
import { assertExists } from "../../deps.ts";
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

Deno.test("Message Index", async (t) => {
    await t.step("should export all session client exports", () => {
        for (const key of Object.keys(SessionClient)) {
            assertExists((MessageIndex as any)[key], `SessionClient.${key} should be exported`);
        }
    });
    
    await t.step("should export all session server exports", () => {
        for (const key of Object.keys(SessionServer)) {
            assertExists((MessageIndex as any)[key], `SessionServer.${key} should be exported`);
        }
    });
    
    await t.step("should export all session update exports", () => {
        for (const key of Object.keys(SessionUpdate)) {
            assertExists((MessageIndex as any)[key], `SessionUpdate.${key} should be exported`);
        }
    });
    
    await t.step("should export all announce please exports", () => {
        for (const key of Object.keys(AnnouncePlease)) {
            assertExists((MessageIndex as any)[key], `AnnouncePlease.${key} should be exported`);
        }
    });
    
    await t.step("should export all announce exports", () => {
        for (const key of Object.keys(Announce)) {
            assertExists((MessageIndex as any)[key], `Announce.${key} should be exported`);
        }
    });
    
    await t.step("should export all subscribe exports", () => {
        for (const key of Object.keys(Subscribe)) {
            assertExists((MessageIndex as any)[key], `Subscribe.${key} should be exported`);
        }
    });
    
    await t.step("should export all subscribe ok exports", () => {
        for (const key of Object.keys(SubscribeOk)) {
            assertExists((MessageIndex as any)[key], `SubscribeOk.${key} should be exported`);
        }
    });
    
    await t.step("should export all subscribe update exports", () => {
        for (const key of Object.keys(SubscribeUpdate)) {
            assertExists((MessageIndex as any)[key], `SubscribeUpdate.${key} should be exported`);
        }
    });
    
    await t.step("should export all group exports", () => {
        for (const key of Object.keys(Group)) {
            assertExists((MessageIndex as any)[key], `Group.${key} should be exported`);
        }
    });

    await t.step("should not have any undefined exports", () => {
        const exports = Object.keys(MessageIndex);
        exports.forEach(key => {
            assertExists((MessageIndex as any)[key], `Export ${key} should not be undefined`);
        });
    });
    
    await t.step("should have all expected module exports", () => {
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
            assertExists((MessageIndex as any)[exportName], `${exportName} should be exported`);
        });
    });
});
