import { Service } from "@neuron/sdk";
import type { ValidateOrderInput, ValidateOrderOutput } from "../types";

export const validateOrder = Service({
  name: "validate-order",
  version: "1.0.0",
  description: "Validate incoming order request",
})
  .executor({ name: "neuron:core:set" })
  .inputSchema<ValidateOrderInput>()
  .outputSchema<ValidateOrderOutput>();
