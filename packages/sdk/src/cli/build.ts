import { mkdir, writeFile } from "node:fs/promises";
import { pathToFileURL } from "node:url";
import { dirname, resolve } from "node:path"
import { discoverProject } from "./project";
import { loadConfig } from "./config";

export async function buildCmdHandler(): Promise<void> {
  const project = discoverProject()
  const config = await loadConfig(project.configFile)

  console.log("Building...")

  const entryFilePath = resolve(project.root, config.entry ?? "index.ts")

  console.log("Entry: ", entryFilePath)

  const manifest = await loadManifest(entryFilePath);

  await saveManifest(project.outputFile, manifest);
  console.log("Manifest written to:", project.outputFile);
}

export async function loadManifest(filePath: string): Promise<unknown> {
  const module = await import(pathToFileURL(filePath).href)
  return module.default;
}

export async function saveManifest(outputPath: string, manifest: unknown): Promise<void> {
  if (!outputPath) {
    throw new Error("output path is missing.")
  }

  await mkdir(dirname(outputPath), { recursive: true })
  await writeFile(outputPath, JSON.stringify(manifest, null, 2), "utf8")
}
