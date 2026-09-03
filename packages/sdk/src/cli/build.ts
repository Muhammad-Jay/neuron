import { mkdir, writeFile } from "node:fs";
import { pathToFileURL } from "node:url";
import { dirname, resolve } from "node:path"
import {discoverProject} from "./project";
import {loadConfig} from "./config";

export async function buildCmdHandler(): Promise<void> {
    const project = discoverProject()
    const config = await loadConfig(project.configFile)

    console.log("Building...")

    const entryFIlePath = resolve(project.root, config.entry ?? "index.ts")

    console.log("Entry: ", config.entry)

    const manifest = await loadManifest(entryFIlePath);

    await saveManifest(project.outputFile, manifest);
}

async function loadManifest(filePath: string): Promise<unknown> {
    const manifest = await import(pathToFileURL(filePath).href)

    return manifest.default;
}

async function saveManifest(outputPath: string, manifest: unknown): Promise<void> {
    if (!outputPath) {
        throw new Error("output path is missing.")
    }

    mkdir(dirname(outputPath), {recursive: true}, (err) => {
        throw new Error(err?.message)
    })

    writeFile(outputPath, JSON.stringify(manifest), "utf8", (err) => {
        throw new Error(err?.message)
    })
}