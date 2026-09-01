import { System } from "@neuron/sdk";
import { pipeline } from "./pipeline.js";
import {
  validateOrder,
  parseOrder,
  enrichCustomer,
  calculateTotals,
  authorizePayment,
  capturePayment,
  createShipment,
  sendConfirmation,
} from "./services/index.js";

export const manifest = System("order-processing")
  .version("1.0.0")
  .description("Order processing pipeline")
  .registerAll(
    validateOrder,
    parseOrder,
    enrichCustomer,
    calculateTotals,
    authorizePayment,
    capturePayment,
    createShipment,
    sendConfirmation
  )
  .run(pipeline)
  .toManifest();
