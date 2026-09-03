import { existsSync } from "node:fs";
import { resolve } from "node:path";
import * as process from "node:process";

export type Project = {
  root: string;
  configFile: string | null;
  entryFile: string;
  outputFile: string;
}

export function discoverProject(cwd: string = process.cwd()): Project {
  const root = resolve(cwd);

  const candidates = [
    "neuron.config.ts",
    "neuron.config.js",
    "neuron.config.mjs",
  ]

  let configFile: string | null = null

  for (const fileName of candidates) {
    const file = resolve(root, fileName)

    if (existsSync(file)) {
      configFile = file;
    }
  }



  const entryFile = resolve(root, "index.ts")
  const outputFile = resolve(root, ".neuron", "manifest.json")

  return {
    root,
    configFile,
    entryFile,
    outputFile,
  }
}
