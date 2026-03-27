// Autopus orchestra result collector — opencode text.complete plugin.
// Captures completed text and writes result JSON + done signal to session dir.
const sessId = process.env.AUTOPUS_SESSION_ID;
if (!sessId) process.exit(0);

import { existsSync, writeFileSync, mkdirSync } from "fs";
import { join } from "path";

// Validate session ID to prevent path traversal (alphanumeric, hyphen, underscore only).
if (!/^[a-zA-Z0-9_-]+$/.test(sessId)) process.exit(0);

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

  // Use round-scoped file names when AUTOPUS_ROUND is set (integer-only validation).
  const round = process.env.AUTOPUS_ROUND;
  const validRound = round && /^\d+$/.test(round) ? round : null;
  const suffix = validRound ? `-round${validRound}` : '';
  const resultFile = `opencode${suffix}-result.json`;
  const doneFile = `opencode${suffix}-done`;

  writeFileSync(join(sessDir, resultFile), result, { mode: 0o600 });
  writeFileSync(join(sessDir, doneFile), "", { mode: 0o600 });
});
