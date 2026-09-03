import { Service } from "@neuron/sdk";
import type { CapturePaymentInput, CapturePaymentOutput } from "../types";

export const capturePayment = Service({
  name: "capture-payment",
  version: "1.0.0",
  description: "Capture an authorized payment",
})
  .executor({ name: "set" })
  .inputSchema<CapturePaymentInput>()
  .outputSchema<CapturePaymentOutput>();
