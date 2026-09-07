import { Service } from "@neuron/sdk";
import type { SendConfirmationInput, SendConfirmationOutput } from "../types";

export const sendConfirmation = Service({
  name: "send-confirmation",
  version: "1.0.0",
  description: "Send order confirmation email",
})
  .executor({ name: "neuron:core:set" })
  .inputSchema<SendConfirmationInput>()
  .outputSchema<SendConfirmationOutput>();
