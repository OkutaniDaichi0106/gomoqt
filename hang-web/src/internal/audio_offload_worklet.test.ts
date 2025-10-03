import { describe, test, expect, it, afterEach, vi } from 'vitest';
import fs from "node:fs";
import path from "node:path";
import { pathToFileURL } from "node:url";
import ts from "typescript";

type OffloadModule = typeof import("./audio_offload_worklet");

function loadWorkletModule(relativePath: string): OffloadModule {
    const fullPath = path.resolve(__dirname, relativePath);
    const source = fs.readFileSync(fullPath, "utf8");
    const { outputText } = ts.transpileModule(source, {
        compilerOptions: {
            module: ts.ModuleKind.CommonJS,
            target: ts.ScriptTarget.ES2020,
        },
    });

    const transformed = outputText.replace(/import\.meta\.url/g, "pathToFileURL(__filename).href");

    const module = { exports: {} as OffloadModule };
    const fn = new Function("exports", "require", "module", "__filename", "__dirname", "pathToFileURL", transformed);
    fn(module.exports, require, module, fullPath, path.dirname(fullPath), pathToFileURL);
    return module.exports;
}

describe("audio_offload_worklet", () => {
    afterEach(() => {
        delete (globalThis as any).AudioWorkletProcessor;
        delete (globalThis as any).registerProcessor;
    });

    it("provides a URL for the offload worklet", () => {
        const module = loadWorkletModule("./audio_offload_worklet.ts");
        const url = module.importUrl();

        expect(url).toMatch(/audio_offload_worklet\.js$/);
    expect(() => new URL(url)).not.toThrow();
    });

    it("registers the offload processor when AudioWorkletProcessor is defined", () => {
        const registerProcessor = vi.fn();
        (globalThis as any).AudioWorkletProcessor = class {
            port = { onmessage: undefined };
        };
        (globalThis as any).registerProcessor = registerProcessor;

        const module = loadWorkletModule("./audio_offload_worklet.ts");

        expect(registerProcessor).toHaveBeenCalledTimes(1);
        const [name, processorCtor] = registerProcessor.mock.calls[0];
        expect(name).toBe("audio-offloader");
        expect(typeof processorCtor).toBe("function");

        const ProcessorCtor = processorCtor as new (options: any) => any;
        const instance = new ProcessorCtor({
            channelCount: 2,
            processorOptions: {
                sampleRate: 48000,
                latency: 50,
            },
        });

        expect(instance).toBeInstanceOf(processorCtor);
        expect(typeof instance.process).toBe("function");
        expect(module.importUrl).toBeDefined();
    });
});
