// Autopus orchestra result collector — opencode text.complete plugin.
// Captures completed text and writes result JSON + done signal to session dir.
const sessId = process.env.AUTOPUS_SESSION_ID;
if (!sessId) process.exit(0);

import { existsSync, writeFileSync, mkdirSync, unlinkSync, readFileSync } from "fs";
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

  // --- Bidirectional IPC: Ready signal + Input watch loop (SPEC-ORCH-017) ---
  if (validRound) {
    const nextRound = String(Number(validRound) + 1);
    const readyFile = `opencode-round${nextRound}-ready`;
    const inputFile = `opencode-round${nextRound}-input.json`;
    const abortFile = `opencode-round${nextRound}-abort`;

    writeFileSync(join(sessDir, readyFile), "", { mode: 0o600 });

    // Poll for input file (200ms intervals, 120s timeout).
    // @AX:NOTE [AUTO] magic constants 200ms/120s — must match Go-side fileIPCReadyTimeout budget
    let waited = 0;
    const maxWait = 120000;
    const poll = setInterval(() => {
      waited += 200;
      const abortPath = join(sessDir, abortFile);
      const inputPath = join(sessDir, inputFile);

      if (existsSync(abortPath)) {
        clearInterval(poll);
        try { unlinkSync(join(sessDir, readyFile)); } catch {}
        try { unlinkSync(abortPath); } catch {}
        process.exit(0);
      }

      if (existsSync(inputPath)) {
        clearInterval(poll);
        try {
          const raw = readFileSync(inputPath, "utf-8");
          const parsed = JSON.parse(raw);
          const prompt = parsed.prompt || "";
          try { unlinkSync(inputPath); } catch {}
          try { unlinkSync(join(sessDir, readyFile)); } catch {}
          if (prompt) {
            process.stdout.write(prompt);
          }
        } catch {}
        process.exit(0);
      }

      if (waited >= maxWait) {
        clearInterval(poll);
        try { unlinkSync(join(sessDir, readyFile)); } catch {}
        process.exit(0);
      }
    }, 200);
  }
});
