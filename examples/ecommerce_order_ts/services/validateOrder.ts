import { Service, record, boolean, string } from "@neuron/sdk";

export const validateOrder = Service("validate-order")
  .version("1.0.0")
  .description("Validate incoming order request")
  .executor("set", { status: "validated", valid: true })
  .inputSchema({
    order: record().required(),
  })
  .outputSchema({
    valid: boolean(),
    status: string(),
  });