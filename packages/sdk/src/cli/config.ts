import {createJiti} from "jiti";
import {ProjectConfig} from "@/manifest";

export type NeuronConfig = {
    entry?: string
    script?: {
        build?: string
    }
    // Project-level configuration merged into the manifest's `config`
    // field so .neuron/manifest.json is the single source of truth.
    config?: ProjectConfig
}

export function defineConfig(config: NeuronConfig): NeuronConfig {
    return config
}

export async function loadConfig(configFile: string | null): Promise<NeuronConfig> {
    if (!configFile) return {}

    try {
        const jiti = createJiti(import.meta.url);
        const module = await jiti.import<{ default?: NeuronConfig }>(configFile);

        return module.default ?? {}
    } catch (error) {
        throw new Error(`Failed to load config file: ${error instanceof Error ? error.message : String(error)}`);
    }
}
