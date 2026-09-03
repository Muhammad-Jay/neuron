import { pathToFileURL } from "node:url"

export type NeuronConfig = {
    entry?: string
    script?: {
        build?: string
    }
}

export function defineConfig(config: NeuronConfig): NeuronConfig {
    return config
}

export async function loadConfig(configFile: string | null): Promise<NeuronConfig> {
    if (!configFile) return {}

    const module = await import(pathToFileURL(configFile).href)

    return module.default ?? {}
}