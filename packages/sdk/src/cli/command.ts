import * as process from "node:process";
import { buildCmdHandler } from "./build";

async function main(): Promise<void> {
    const [command] = process.argv.slice(2)

    switch (command) {
        case "build":
            await buildCmdHandler()
            return
        case "version":
        case "--version":
        case "-v":
            console.log(getVersion())
            return
        case "help":
        case "--help":
        case "-h":
            printHelp()
            return
        case undefined:
            printHelp()
            return
        default:
            console.error(`Unknown command: "${command}"`)
            printHelp()
            process.exit(1)
    }
}

function getVersion(): string {
    return "0.1.0"
}

function printHelp(): void {
    console.log(`
Neuron SDK

Usage:
  neuron-sdk <command> [options]

Commands:
  build     Compile the project into a manifest (.neuron/manifest.json)
  version   Print the SDK version
  help      Show this help message
`)
}

main().catch((error) => {
    console.error("Neuron SDK error")

    if (error instanceof Error) {
        console.error(error.message)
    } else {
        console.error(error)
    }

    process.exit(1)
})
