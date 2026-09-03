import { Service } from "@neuron/sdk";
import type { CreateShipmentInput, CreateShipmentOutput } from "../types";

export const createShipment = Service({
  name: "create-shipment",
  version: "1.0.0",
  description: "Create a shipment for the order",
})
  .executor({ name: "set" })
  .inputSchema<CreateShipmentInput>()
  .outputSchema<CreateShipmentOutput>();
