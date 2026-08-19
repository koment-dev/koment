import { spawn } from "node:child_process";
import { readFileSync } from "node:fs";

const packageVersion = JSON.parse(
  readFileSync(new URL("./package.json", import.meta.url), "utf8")
).version;

export function connectMCP(child) {
  const pending = new Map();
  let nextID = 1;
  let buffer = "";
  let terminalError;

  function failConnection(error) {
    if (terminalError) return;
    terminalError = error;
    for (const waiter of pending.values()) waiter.reject(error);
    pending.clear();
  }

  child.stderr.on("data", (chunk) => {
    const text = chunk.toString("utf8");
    for (const line of text.split("\n")) {
      if (line) process.stderr.write(`[koment-mcp] ${line}\n`);
    }
  });

  child.stdout.on("data", (chunk) => {
    buffer += chunk.toString("utf8");
    let newline;
    while ((newline = buffer.indexOf("\n")) !== -1) {
      const line = buffer.slice(0, newline);
      buffer = buffer.slice(newline + 1);
      if (!line) continue;
      let message;
      try {
        message = JSON.parse(line);
      } catch (error) {
        failConnection(new Error("koment MCP emitted invalid JSON", { cause: error }));
        child.kill();
        return;
      }
      if (!("id" in message)) {
        continue;
      }
      const id = message.id;
      const waiter = pending.get(id);
      if (!waiter) {
        failConnection(new Error(`koment MCP returned unexpected response id ${id}`));
        child.kill();
        return;
      }
      pending.delete(id);
      if (message.error) {
        waiter.reject(new Error(message.error.message || "mcp error"));
      } else {
        waiter.resolve(message.result ?? {});
      }
    }
  });

  child.on("error", failConnection);
  child.stdin.on("error", failConnection);
  child.stdout.on("error", failConnection);
  child.stderr.on("error", failConnection);
  child.on("exit", (code, signal) => {
    const outcome = signal ? `signal=${signal}` : `code=${code}`;
    failConnection(new Error(`koment mcp exited (${outcome})`));
  });

  function request(method, params) {
    if (terminalError) return Promise.reject(terminalError);
    const id = nextID++;
    const payload = JSON.stringify({ jsonrpc: "2.0", id, method, params: params || {} }) + "\n";
    return new Promise((resolve, reject) => {
      pending.set(id, { resolve, reject });
      child.stdin.write(payload, (err) => {
        if (err) failConnection(err);
      });
    });
  }

  function notify(method, params) {
    if (terminalError) return Promise.reject(terminalError);
    const payload = JSON.stringify({ jsonrpc: "2.0", method, params: params || {} }) + "\n";
    return new Promise((resolve, reject) => {
      child.stdin.write(payload, (error) => {
        if (error) {
          failConnection(error);
          reject(error);
          return;
        }
        resolve();
      });
    });
  }

  return {
    child,
    request,
    initialize() {
      return request("initialize", {
        protocolVersion: "2024-11-05",
        capabilities: {},
        clientInfo: { name: "@koment/opencode-koment", version: packageVersion },
      });
    },
    notify,
    async callTool(name, args) {
      const result = await request("tools/call", { name, arguments: args || {} });
      if (Array.isArray(result.content)) {
        for (const block of result.content) {
          if (block && block.type === "text" && block.text) {
            process.stderr.write(`[koment-mcp:${name}] ${block.text}\n`);
          }
        }
      }
      return result;
    },
    close() {
      if (terminalError) return Promise.reject(terminalError);
      return new Promise((resolve, reject) => {
        child.stdin.end((error) => {
          if (error) reject(error);
          else resolve();
        });
      });
    },
  };
}

function startMCP(directory) {
  return connectMCP(
    spawn("koment", ["mcp", "--write"], {
      cwd: directory,
      env: process.env,
      stdio: ["pipe", "pipe", "pipe"],
    })
  );
}

function run(cmd, args, { cwd, stdin } = {}) {
  return new Promise((resolve, reject) => {
    const child = spawn(cmd, args, {
      cwd,
      env: process.env,
      stdio: ["pipe", "pipe", "pipe"],
    });
    let stdout = "";
    let stderr = "";
    child.stdout.on("data", (chunk) => (stdout += chunk.toString("utf8")));
    child.stderr.on("data", (chunk) => (stderr += chunk.toString("utf8")));
    child.on("error", reject);
    if (stdin !== undefined && stdin !== null) {
      child.stdin.end(stdin);
    } else {
      child.stdin.end();
    }
    child.on("close", (code) => {
      if (code === 0) resolve({ stdout, stderr });
      else {
        const error = new Error("koment " + args.join(" ") + " exited with code " + code);
        error.stdout = stdout;
        error.stderr = stderr;
        error.code = code;
        reject(error);
      }
    });
  });
}

function deny(reason) {
  throw new Error(reason);
}

export default async ({ directory }) => {
  let mcp;
  try {
    mcp = startMCP(directory);
    await mcp.initialize();
    await mcp.notify("notifications/initialized");
  } catch (err) {
    throw new Error(
      `koment plugin failed to start the MCP server in ${directory}:\n` +
        (err && err.message ? err.message : String(err)) +
        "\nInstall the \`koment\` binary on PATH: https://github.com/koment-dev/koment/releases"
    );
  }

  return {
    "tool.execute.before": async (input, output) => {
      const tool = input.tool;
      if (tool !== "edit" && tool !== "write") return;
      const args = output.args ?? {};
      const filePath = args.filePath ?? args.path ?? args.file ?? "";
      const content = args.content ?? args.newContent ?? args.text ?? "";
      if (!filePath || typeof content !== "string") return;
      try {
        const result = await mcp.callTool("koment_pre_tool", {
          tool_name: "opencode_edit",
          filePath,
          content,
        });
        const structured = result.structuredContent || {};
        if (structured.decision === "deny") {
          deny(structured.reason || "koment policy denied this edit");
        }
      } catch (err) {
        deny("koment pre-tool MCP call failed: " + (err && err.message ? err.message : String(err)));
      }
    },
    dispose: async () => {
      const failures = [];
      for (const command of [
        ["check"],
        ["comments", "check"],
        ["agents", "check"],
      ]) {
        try {
          await run("koment", command, { cwd: directory });
        } catch (err) {
          failures.push(
            "koment " + command.join(" ") + ":\n" + (err.stderr || err.message || String(err))
          );
        }
      }
      try {
        await mcp.close();
      } catch (err) {
        failures.push("close koment mcp:\n" + (err.message || String(err)));
      }
      if (failures.length > 0) {
        deny("koment policy gate failed:\n" + failures.join("\n"));
      }
    },
  };
};
