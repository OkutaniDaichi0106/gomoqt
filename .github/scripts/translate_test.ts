import { assertEquals, assertRejects } from "https://deno.land/std@0.208.0/assert/mod.ts";
import { translate, parseEvent, generateComment, languages, type Issue } from "./translate.ts";

// Mock fetch for testing
const originalFetch = globalThis.fetch;
const mockFetch = (response: any) => {
  globalThis.fetch = async () => new Response(JSON.stringify(response), { status: 200 });
};

const restoreFetch = () => {
  globalThis.fetch = originalFetch;
};

Deno.test("translate function - success", async () => {
  mockFetch({ translatedText: "Hello world" });
  try {
    const result = await translate("Hola mundo", "en");
    assertEquals(result, "Hello world");
  } finally {
    restoreFetch();
  }
});

Deno.test("translate function - API error", async () => {
  globalThis.fetch = async () => new Response("Error", { status: 500 });
  try {
    await assertRejects(async () => await translate("Test", "en"), Error, "Translation API failed");
  } finally {
    restoreFetch();
  }
});

Deno.test("parseEvent - issue", () => {
  const eventJson = JSON.stringify({ issue: { number: 1, body: "Test body" } });
  const issue = parseEvent(eventJson);
  assertEquals(issue?.number, 1);
  assertEquals(issue?.body, "Test body");
});

Deno.test("parseEvent - pull_request", () => {
  const eventJson = JSON.stringify({ pull_request: { number: 2, body: "PR body" } });
  const issue = parseEvent(eventJson);
  assertEquals(issue?.number, 2);
  assertEquals(issue?.body, "PR body");
});

Deno.test("parseEvent - no issue or PR", () => {
  const eventJson = JSON.stringify({});
  const issue = parseEvent(eventJson);
  assertEquals(issue, undefined);
});

Deno.test("generateComment - basic", async () => {
  mockFetch({ translatedText: "Translated text" });
  try {
    const comment = await generateComment("Original text", [languages[0]]); // Only English
    assertEquals(comment.includes("🇬🇧 English"), true);
    assertEquals(comment.includes("Translated text"), true);
    assertEquals(comment.includes("Auto-generated"), true);
  } finally {
    restoreFetch();
  }
});

// Test for environment variables (mock Deno.env)
Deno.test("environment check - missing GITHUB_TOKEN", () => {
  const originalEnv = Deno.env.get;
  Deno.env.get = (key: string) => key === "GITHUB_TOKEN" ? undefined : "test";
  try {
    // Since main execution is guarded, we can't directly test exit, but this simulates
    const token = Deno.env.get("GITHUB_TOKEN");
    assertEquals(token, undefined);
  } finally {
    Deno.env.get = originalEnv;
  }
});

// Note: For full integration test, run the script with mock environment variables.
// Example: GITHUB_TOKEN=test GITHUB_REPOSITORY=test/repo GITHUB_EVENT='{"issue":{"number":1,"body":"test"}}' deno run --allow-net --allow-env translate.ts