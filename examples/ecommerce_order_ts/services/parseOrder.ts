import { Service, record, string } from "@neuron/sdk";

export const parseOrder = Service("parse-order")
  .version("1.0.0")
  .description("Parse and normalize order data")
  .executor("set", { currency: "USD" })
  .inputSchema({
    validationData: record().required(),
  })
  .outputSchema({
    currency: string(),
  });
