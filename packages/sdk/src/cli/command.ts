import * as process from "node:process";
import {buildCmdHandler} from "./build";

async function main() {
    const [command] = process.argv.slice(2)

    switch (command) {
        case "build":
            await buildCmdHandler()
            return
        case undefined:
            printHelp()
            process.exit(1);
            return
        default:
            console.log(`Unknown command ${command}`)
            printHelp()
            process.exit(1)
            return
    }
}

function printHelp(){
    console.log(`
    Neuron SDK
    
    Usage:
      neuron-sdk build
    
    Commands:
      Build: build the project
    `)
}

main().catch((error) => {
    console.log("Neuron SDK error")

    if (error instanceof Error){
        console.error(error.message)
    }else {
        console.error(error)
    }

    process.exit(1)
})