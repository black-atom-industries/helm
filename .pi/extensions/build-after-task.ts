import type { ExtensionAPI } from "@mariozechner/pi-coding-agent";

/**
 * Post-task build hook.
 * Runs `make build` after the agent finishes responding to a user prompt.
 * Notifies on both success and failure so you know the project still compiles.
 */
export default function (pi: ExtensionAPI) {
  pi.on("agent_end", async (_event, ctx) => {
    ctx.ui.setStatus("build", "make build...");

    try {
      const { code, stdout, stderr } = await pi.exec("make", ["build"], {
        timeout: 60_000,
      });

      if (code === 0) {
        ctx.ui.notify("✓ make build succeeded", "success");
      } else {
        const errText = (stderr || stdout || "").trim();
        const snippet = errText.split("\n").slice(-5).join("\n");
        ctx.ui.notify(`✗ make build failed (exit ${code}): ${snippet}`, "error");
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      ctx.ui.notify(`✗ make build error: ${msg}`, "error");
    } finally {
      ctx.ui.setStatus("build", undefined);
    }
  });
}
