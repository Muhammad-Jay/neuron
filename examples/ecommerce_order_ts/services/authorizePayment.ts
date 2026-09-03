import { Service } from "@neuron/sdk";
import type { AuthorizePaymentInput, AuthorizePaymentOutput } from "../types";

export const authorizePayment = Service({
  name: "authorize-payment",
  version: "1.0.0",
  description: "Authorize payment for the order",
})
  .executor({ name: "set" })
  .inputSchema<AuthorizePaymentInput>()
  .outputSchema<AuthorizePaymentOutput>();
