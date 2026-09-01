import { Service, record, string, number } from "@neuron/sdk";

export const authorizePayment = Service("authorize-payment")
  .version("1.0.0")
  .description("Authorize payment")
  .executor("set", {
    payment_intent: { id: "pi_abc123", status: "requires_capture" },
  })
  .inputSchema({
    amountCents: number().required(),
    currency: string().required(),
    email: string().email(),
  })
  .outputSchema({
    paymentIntent: record(),
  });
