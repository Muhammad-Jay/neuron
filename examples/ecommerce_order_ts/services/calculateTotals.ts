import { Service, record, string, number } from "@neuron/sdk";

export const calculateTotals = Service("calculate-totals")
  .version("1.0.0")
  .description("Calculate totals")
  .executor("set", { tax_rate: 0.0825, discount_rate: 0.1 })
  .inputSchema({
    items: record().required(),
    customerTier: string(),
    shippingState: string(),
    email: string().email(),
  })
  .outputSchema({
    total: number(),
  });
