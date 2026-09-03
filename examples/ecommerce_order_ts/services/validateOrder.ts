import { Service, record, string, boolean } from "@neuron/sdk";
import {OrderInput} from "../types";

export const validateOrder = Service("validate-order")
  .version("1.0.0")
  .description("Validate incoming order request")
  .executor("set", { status: "validated", valid: true })
  .inputSchema({
      order: record().required()
  })
  .outputSchema({
    valid: boolean().required(),
    status: string().required(),
  });