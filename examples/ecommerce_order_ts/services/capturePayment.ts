import { Service, record, string } from "@neuron/sdk";

export const capturePayment = Service("capture-payment")
  .version("1.0.0")
  .description("Capture payment")
  .executor("set", {
    capture_result: { status: "succeeded", amount_received: 7997 },
  })
  .inputSchema({
    paymentIntentId: string().required(),
  })
  .outputSchema({
    captureResult: record(),
  });
