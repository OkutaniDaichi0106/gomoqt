const path = Deno.args[0];
if (!path) {
	console.error("Usage: deno run check_balance.ts <file>");
	Deno.exit(1);
}
const src = await Deno.readTextFile(path);
const pairs = { "{": "}", "(": ")", "[": "]" };
const opens = new Set(["{", "(", "["]);
const stack = [];
for (let i = 0; i < src.length; i++) {
	const ch = src[i];
	if (opens.has(ch)) stack.push({ ch, i });
	else if (ch === "}" || ch === ")" || ch === "]") {
		const last = stack[stack.length - 1];
		if (last && pairs[last.ch] === ch) stack.pop();
		else {
			console.log("Unmatched closing", ch, "at index", i);
			const line = src.slice(0, i).split("\n").length;
			const col = i - src.lastIndexOf("\n", i);
			console.log("line", line, "col", col);
			Deno.exit(0);
		}
	}
}
if (stack.length) {
	const last = stack[stack.length - 1];
	const line = src.slice(0, last.i).split("\n").length;
	const col = last.i - src.lastIndexOf("\n", last.i);
	console.log("Unclosed", last.ch, "at index", last.i, "line", line, "col", col);
} else {
	console.log("All balanced");
}
