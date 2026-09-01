import { Service, string, number, boolean } from "@neuron/sdk";

export const sendConfirmation = Service("send-confirmation")
  .version("1.0.0")
  .description("Send confirmation email")
  .executor("set", { confirmation_sent: true, message_id: "msg_xyz789" })
  .inputSchema({
    trackingNumber: string(),
    carrier: string(),
    email: string().email(),
    grandTotal: number(),
  })
  .outputSchema({
    confirmationSent: boolean(),
  });
