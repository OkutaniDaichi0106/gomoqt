import { join } from "@std/path";

// 1. Get mkcert CAROOT
// We use "cmd /c mkcert -CAROOT" on Windows to ensure it runs correctly if it's a batch file,
// but usually calling "mkcert" directly works if it's in PATH.
const cmd = Deno.build.os === "windows" ? "mkcert.exe" : "mkcert";

try {
  const mkcert = new Deno.Command(cmd, {
    args: ["-CAROOT"],
    stdout: "piped",
    stderr: "inherit",
  });
  const output = await mkcert.output();

  if (!output.success) {
    console.error(
      "Error: 'mkcert -CAROOT' failed. Please ensure mkcert is installed and in your PATH.",
    );
    Deno.exit(1);
  }

  const caRoot = new TextDecoder().decode(output.stdout).trim();
  const certPath = join(caRoot, "rootCA.pem");

  console.log(`[Secure Wrapper] Using Root CA from: ${certPath}`);

  // 2. Run the actual interop client with the cert
  const child = new Deno.Command(Deno.execPath(), {
    args: [
      "run",
      "--unstable-net",
      "--allow-all",
      "--cert",
      certPath,
      "cli/interop/main.ts",
      ...Deno.args, // Pass through any extra args
    ],
    stdout: "inherit",
    stderr: "inherit",
    stdin: "inherit",
  });

  const status = await child.spawn().status;
  Deno.exit(status.code);
} catch (err) {
  if (err instanceof Deno.errors.NotFound) {
    console.error("Error: 'mkcert' not found in PATH. Please install mkcert.");
  } else {
    console.error("Error running secure wrapper:", err);
  }
  Deno.exit(1);
}
