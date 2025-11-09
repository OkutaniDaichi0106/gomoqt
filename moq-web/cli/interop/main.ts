const projectRoot = new URL("../../..", import.meta.url).pathname.slice(1); // Remove leading slash on Windows

const cmd = new Deno.Command("go", {
    args: ["run", `${projectRoot}/cmd/interop/server`],
    cwd: projectRoot,
});

const { code } = await cmd.output();
Deno.exit(code);