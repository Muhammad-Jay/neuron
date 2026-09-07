import { System } from "@neuron/sdk";
import { buildPipeline } from "./pipeline.js";
import type { SystemInput } from "./types.js";

const manifest = System({
  name: "order-processing-ts",
  version: "2.1.0",
  description: "Order processing pipeline",
})
  .inputSchema<SystemInput>()
  .withParams((input) => buildPipeline(input))
  .toManifest();

export default manifest;
