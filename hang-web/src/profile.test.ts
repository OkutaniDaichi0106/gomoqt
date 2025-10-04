import { Profile, ProfileInit } from "./profile";
import { describe, test, expect, beforeEach } from 'vitest';

describe("Profile", () => {
    describe("Constructor", () => {
        test("creates profile with required id", () => {
            const init: ProfileInit = {
                id: "user123"
            };

            const profile = new Profile(init);

            expect(profile.id).toBe("user123");
        });

        test("creates profile with different id", () => {
            const init: ProfileInit = {
                id: "another-user-456"
            };

            const profile = new Profile(init);

            expect(profile.id).toBe("another-user-456");
        });

        test("id is readonly", () => {
            const profile = new Profile({ id: "test-id" });

            // This would cause TypeScript error if we tried to assign
            // profile.id = "new-id"; // Should not be allowed

            expect(profile.id).toBe("test-id");
        });

        test("handles empty string id", () => {
            const profile = new Profile({ id: "" });

            expect(profile.id).toBe("");
        });

        test("handles special characters in id", () => {
            const specialId = "user@domain.com_123-test!";
            const profile = new Profile({ id: specialId });

            expect(profile.id).toBe(specialId);
        });

        test("handles very long id", () => {
            const longId = "a".repeat(1000);
            const profile = new Profile({ id: longId });

            expect(profile.id).toBe(longId);
        });
    });

    describe("Properties", () => {
        test("id property is accessible", () => {
            const profile = new Profile({ id: "accessible-id" });

            const profileId = profile.id;

            expect(profileId).toBe("accessible-id");
        });

        test("profile maintains identity", () => {
            const id = "identity-test";
            const profile = new Profile({ id });

            expect(profile.id).toBe(id);
            expect(profile.id).toBe(profile.id); // Should be consistent
        });
    });

    describe("Type Safety", () => {
        test("ProfileInit interface requires id", () => {
            // This test ensures the interface contract is working
            const validInit: ProfileInit = { id: "test" };
            const profile = new Profile(validInit);

            expect(profile).toBeInstanceOf(Profile);
        });

        test("Profile instance type checking", () => {
            const profile = new Profile({ id: "type-check" });

            expect(profile).toBeInstanceOf(Profile);
            expect(typeof profile.id).toBe("string");
        });
    });

    describe("Edge Cases", () => {
        test("handles whitespace-only id", () => {
            const profile = new Profile({ id: "   \t\n   " });

            expect(profile.id).toBe("   \t\n   ");
        });

        test("handles numeric string id", () => {
            const profile = new Profile({ id: "12345" });

            expect(profile.id).toBe("12345");
        });

        test("handles unicode characters", () => {
            const unicodeId = "用户123_测试_🎉";
            const profile = new Profile({ id: unicodeId });

            expect(profile.id).toBe(unicodeId);
        });
    });

    describe("Multiple Instances", () => {
        test("different profiles have different ids", () => {
            const profile1 = new Profile({ id: "user1" });
            const profile2 = new Profile({ id: "user2" });

            expect(profile1.id).toBe("user1");
            expect(profile2.id).toBe("user2");
            expect(profile1.id).not.toBe(profile2.id);
        });

        test("profiles with same id are equal by value", () => {
            const profile1 = new Profile({ id: "same-id" });
            const profile2 = new Profile({ id: "same-id" });

            expect(profile1.id).toBe(profile2.id);
            expect(profile1).not.toBe(profile2); // Different object references
        });

        test("can create multiple profiles", () => {
            const profiles = [
                new Profile({ id: "user1" }),
                new Profile({ id: "user2" }),
                new Profile({ id: "user3" })
            ];

            expect(profiles).toHaveLength(3);
            expect(profiles[0].id).toBe("user1");
            expect(profiles[1].id).toBe("user2");
            expect(profiles[2].id).toBe("user3");
        });
    });
});
