// Autopus orchestra result collector — opencode text.complete plugin.
// Captures completed text and writes result JSON + done signal to session dir.
const sessId = process.env.AUTOPUS_SESSION_ID;
if (!sessId) process.exit(0);

import { existsSync, writeFileSync, mkdirSync } from "fs";
import { join } from "path";

const sessDir = join("/tmp/autopus", sessId);
if (!existsSync(sessDir)) process.exit(0);

const chunks: Buffer[] = [];
process.stdin.on("data", (chunk) => chunks.push(chunk));
process.stdin.on("end", () => {
  const input = Buffer.concat(chunks).toString();
  let text = "";
  try {
    const data = JSON.parse(input);
    text = data.text || "";
  } catch {
    text = input;
  }
  if (!text) process.exit(0);

  const result = JSON.stringify({
    output: text,
    exit_code: 0,
  });

  writeFileSync(join(sessDir, "opencode-result.json"), result, { mode: 0o600 });
  writeFileSync(join(sessDir, "opencode-done"), "", { mode: 0o644 });
});
