import { Service } from "@neuron/sdk";
import type { EnrichCustomerInput, EnrichCustomerOutput } from "../types";

export const enrichCustomer = Service({
  name: "enrich-customer",
  version: "1.0.0",
  description: "Enrich with customer data",
})
  .executor({ name: "set" })
  .inputSchema<EnrichCustomerInput>()
  .outputSchema<EnrichCustomerOutput>();
