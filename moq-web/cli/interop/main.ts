import { dirname, join } from "@std/path";

console.log("=== MOQ Interop Test (TypeScript Client) ===");

/**
 * Find the repository root by walking up to find .git directory.
 * Returns the repository root directory (gomoqt).
 */
async function findRoot(): Promise<string> {
	let currentPath = import.meta.dirname!; // Current file's directory

	while (true) {
		const gitPath = join(currentPath, ".git");
		try {
			const stat = await Deno.stat(gitPath);
			if (stat.isDirectory) {
				return currentPath; // Found .git directory
			}
		} catch {
			// .git not found, continue searching
		}

		const parent = dirname(currentPath);
		if (parent === currentPath) {
			// Reached filesystem root without finding .git
			throw new Error("Could not find repository root (.git directory)");
		}
		currentPath = parent;
	}
}

// Determine project root and paths (cross-platform, robust)
const rootPath = await findRoot();


// Create context for server lifecycle
const abortController = new AbortController();

// Start server process using 'go run'
console.log("Starting server...");
const serverCmd = new Deno.Command("go", {
	args: ["run", join(rootPath, "cmd/interop/server")],
	stdout: "piped",
	stderr: "piped",
	env: {
		...Deno.env.toObject(),
		"SLOG_LEVEL": "debug",
	},
	signal: abortController.signal,
});
const serverProcess = serverCmd.spawn();

// Forward server stdout/stderr to console
(async () => {
	const reader = serverProcess.stdout.getReader();
	try {
		while (true) {
			const { done, value } = await reader.read();
			if (done) break;
			Deno.stdout.writeSync(value);
		}
	} finally {
		reader.releaseLock();
	}
})();

(async () => {
	const reader = serverProcess.stderr.getReader();
	try {
		while (true) {
			const { done, value } = await reader.read();
			if (done) break;
			Deno.stderr.writeSync(value);
		}
	} finally {
		reader.releaseLock();
	}
})();

// Wait for server to start
console.log("Waiting for server to start...");
await new Promise((resolve) => setTimeout(resolve, 3000));

// Run client and wait for completion
console.log("Starting client...");
let clientStatus: Deno.CommandStatus | null = null;

try {
	const clientCmd = new Deno.Command("deno", {
		args: ["run", "--log-level=info", "--allow-net", "--allow-read", join(rootPath, "moq-web/cli/interop/client/main.ts")],
		cwd: rootPath,
		stdout: "inherit",
		stderr: "inherit",
	});
	const clientProcess = clientCmd.spawn();

	// Wait for client to finish
	clientStatus = await clientProcess.status;

			if (!clientStatus.success) {
				console.error("[ERR] Client failed");
			} else {
				console.log("[OK] Client completed successfully");
			}
} catch (error) {
	console.error("Client execution error:", error);
} finally {
	// Clean up: terminate server
	console.log("Stopping server...");
	abortController.abort();
	
	// Wait for the server process to terminate
	try {
		await serverProcess.status;
	} catch {
		// Process cleanup completed
	}
}

console.log("=== Interop Test Completed ===");
