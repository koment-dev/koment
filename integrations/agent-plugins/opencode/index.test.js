import assert from "node:assert/strict";
import { EventEmitter } from "node:events";
import { readFileSync } from "node:fs";
import { PassThrough } from "node:stream";
import test from "node:test";

import { connectMCP } from "./index.js";

function fakeChild() {
  const child = new EventEmitter();
  child.stdin = new PassThrough();
  child.stdout = new PassThrough();
  child.stderr = new PassThrough();
  child.wasKilled = false;
  child.kill = () => {
    child.wasKilled = true;
    return true;
  };
  return child;
}

function nextChunk(stream) {
  return new Promise((resolve) => stream.once("data", resolve));
}

test("initialize identifies the installed package version", async () => {
  const child = fakeChild();
  const client = connectMCP(child);
  const written = nextChunk(child.stdin);
  const initialized = client.initialize();
  const request = JSON.parse((await written).toString("utf8"));
  const packageVersion = JSON.parse(readFileSync(new URL("./package.json", import.meta.url))).version;

  assert.equal(request.method, "initialize");
  assert.equal(request.params.clientInfo.version, packageVersion);
  child.stdout.write(JSON.stringify({ jsonrpc: "2.0", id: request.id, result: {} }) + "\n");
  await initialized;
});

test("fragmented responses resolve the matching request", async () => {
  const child = fakeChild();
  const client = connectMCP(child);
  const written = nextChunk(child.stdin);
  const initialized = client.initialize();
  const request = JSON.parse((await written).toString("utf8"));
  const response = JSON.stringify({ jsonrpc: "2.0", id: request.id, result: { ready: true } }) + "\n";

  child.stdout.write(response.slice(0, 7));
  child.stdout.write(response.slice(7));
  assert.deepEqual(await initialized, { ready: true });
});

test("malformed server output rejects instead of hanging", async () => {
  const child = fakeChild();
  const client = connectMCP(child);
  const initialized = client.initialize();

  child.stdout.write("not-json\n");
  await assert.rejects(initialized, /emitted invalid JSON/);
  assert.equal(child.wasKilled, true);
});

test("an unexpected response id terminates the connection", async () => {
  const child = fakeChild();
  const client = connectMCP(child);
  const initialized = client.initialize();

  child.stdout.write(JSON.stringify({ jsonrpc: "2.0", id: 999, result: {} }) + "\n");
  await assert.rejects(initialized, /unexpected response id 999/);
  assert.equal(child.wasKilled, true);
});

test("process exit rejects every outstanding request", async () => {
  const child = fakeChild();
  const client = connectMCP(child);
  const initialized = client.initialize();

  child.emit("exit", 7, null);
  await assert.rejects(initialized, /exited \(code=7\)/);
  await assert.rejects(client.notify("notifications/initialized"), /exited \(code=7\)/);
});

test("close waits for stdin to finish", async () => {
  const child = fakeChild();
  const client = connectMCP(child);
  let finished = false;
  child.stdin.on("finish", () => {
    finished = true;
  });

  await client.close();
  assert.equal(finished, true);
});
