import { describe, test, expect, it, afterEach, vi } from 'vitest';
import fs from "node:fs";
import path from "node:path";
import { pathToFileURL } from "node:url";
import ts from "typescript";

type WorkletModule = typeof import("./audio_hijack_worklet");

function loadWorkletModule(relativePath: string): WorkletModule {
    const fullPath = path.resolve(__dirname, relativePath);
    const source = fs.readFileSync(fullPath, "utf8");
    const { outputText } = ts.transpileModule(source, {
        compilerOptions: {
            module: ts.ModuleKind.CommonJS,
            target: ts.ScriptTarget.ES2020,
        },
    });

    const transformed = outputText.replace(/import\.meta\.url/g, "pathToFileURL(__filename).href");

    const module = { exports: {} as WorkletModule };
    const fn = new Function("exports", "require", "module", "__filename", "__dirname", "pathToFileURL", transformed);
    fn(module.exports, require, module, fullPath, path.dirname(fullPath), pathToFileURL);
    return module.exports;
}

describe("audio_hijack_worklet", () => {
    afterEach(() => {
        delete (globalThis as any).AudioWorkletProcessor;
        delete (globalThis as any).registerProcessor;
    });

    it("provides a URL for the worklet script", () => {
        const module = loadWorkletModule("./audio_hijack_worklet.ts");
        const url = module.importWorkletUrl();

        expect(url).toMatch(/audio_hijack_worklet\.js$/);
        expect(() => new URL(url)).not.toThrow();
    });

    it("registers the hijack processor when AudioWorkletProcessor is available", () => {
        const registerProcessor = vi.fn();
        (globalThis as any).AudioWorkletProcessor = class {};
        (globalThis as any).registerProcessor = registerProcessor;

        loadWorkletModule("./audio_hijack_worklet.ts");

        expect(registerProcessor).toHaveBeenCalledTimes(1);
        const [name, processorCtor] = registerProcessor.mock.calls[0];
        expect(name).toBe("AudioHijacker");
        expect(typeof processorCtor).toBe("function");
        expect(Object.getPrototypeOf((processorCtor as Function).prototype)).toBe((globalThis as any).AudioWorkletProcessor.prototype);
    });
});
