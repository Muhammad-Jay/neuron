import { mkdir, writeFile } from "node:fs/promises";
import { dirname, resolve } from "node:path"
import { discoverProject } from "./project";
import { loadConfig } from "./config";
import {createJiti} from "jiti";
import {SystemManifest} from "@/manifest";
import * as process from "node:process";

export async function buildCmdHandler(projectDir: string = process.cwd()): Promise<void> {
  const project = discoverProject(projectDir)
  const config = await loadConfig(project.configFile)

  console.log("Building...")

  const entryFilePath = resolve(project.root, config.entry ?? "index.ts")

  console.log("Entry: ", entryFilePath)

  const manifest = await loadManifest(entryFilePath);

  await saveManifest(project.outputFile, manifest);
  console.log("Manifest written to:", project.outputFile);
}

export async function loadManifest(filePath: string): Promise<SystemManifest | undefined> {
  try {
    const jiti = createJiti(import.meta.url);
    const module = await jiti.import<{default?: SystemManifest}>(filePath)
    return module.default;
  }catch (error) {
    throw error
  }
}

export async function saveManifest(outputPath: string, manifest: SystemManifest | undefined): Promise<void> {
  if (!outputPath) {
    throw new Error("output path is missing.")
  }

  await mkdir(dirname(outputPath), { recursive: true })
  await writeFile(outputPath, JSON.stringify(manifest, null, 2), "utf8")
}
