import { Service } from "@neuron/sdk";
import type { ParseOrderInput, ParseOrderOutput } from "../types";

export const parseOrder = Service({
  name: "parse-order",
  version: "1.0.0",
  description: "Parse and normalize order data",
})
  .executor({ name: "set" })
  .inputSchema<ParseOrderInput>()
  .outputSchema<ParseOrderOutput>();
