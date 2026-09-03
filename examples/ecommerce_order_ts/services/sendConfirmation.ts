import { Service, string, number, boolean } from "@neuron/sdk";

export const sendConfirmation = Service("send-confirmation")
  .version("1.0.0")
  .description("Send confirmation email")
  .executor("set")
  .inputSchema({
    trackingNumber: string().required(),
    carrier: string().required(),
    email: string().email().required(),
    grandTotal: number().required(),
  })
  .outputSchema({
    confirmationSent: boolean().required(),
  });
