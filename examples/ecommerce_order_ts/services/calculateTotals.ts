import { Service } from "@neuron/sdk";
import type { CalculateTotalsInput, CalculateTotalsOutput } from "../types";

export const calculateTotals = Service({
  name: "calculate-totals",
  version: "1.0.0",
  description: "Calculate order totals with tax and discounts",
})
  .executor({ name: "set" })
  .inputSchema<CalculateTotalsInput>()
  .outputSchema<CalculateTotalsOutput>();
