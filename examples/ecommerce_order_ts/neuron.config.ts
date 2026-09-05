import { defineConfig } from "@neuron/sdk";

export default defineConfig({
    entry: "./system.ts",
    script: {
        build: "echo 'building'"
    },
    config: {
        inspector: {
            enabled: true
        },
        storage: {
            directory: "./home",
            provider: "postgres"
        },
        runtime: {
            execution: {
                mode: "wait"
            }
        }
    }
})