import { defineConfig } from "@neuron/sdk";

export default defineConfig({
    entry: "./system.ts",
    script: {
        build: "echo 'building'"
    }
})