import { type Expressionify } from "@neuron/sdk";
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
import type { SystemInput } from "./types.js";

/**
 * Builds the order-processing pipeline from the system input.
 *
 * `SystemInput` (execution context) is passed in as `input`, so the original
 * order is reachable from any step via `input.order` (compiled to
 * `execution.input.order`). Each service also forwards the order through its
 * own output, so in-flight maps read it back from the previous step's output
 * (`source.output.order`).
 *
 * Every executor is `neuron:core:set`, which echoes the service input (plus
 * config), so a service only outputs the data its input carried. Bindings are
 * therefore kept to fields the previous step actually emits, and gateway
 * conditions test order data rather than domain objects no mock step produces.
 */
export function buildPipeline(input: Expressionify<{ order: SystemInput["order"] }>) {
  return validateOrder
    .withInput({
      order: input.order,
    })
    .next(
      parseOrder.withInput({
        order: validateOrder.output.order,
        validationData: validateOrder.output,
      })
    )
    .next(
      enrichCustomer.withInput({
        order: parseOrder.output.order,
        customerId: parseOrder.output.order.customerId,
      })
    )
    .next(
      calculateTotals.withInput({
        order: parseOrder.output.order,
        items: parseOrder.output.order.items,
        email: parseOrder.output.order.customerEmail,
      })
    )
    .next(
      authorizePayment.withInput({
        order: parseOrder.output.order,
        amountCents: parseOrder.output.order.total,
        currency: parseOrder.output.order.currency,
        email: parseOrder.output.order.customerEmail,
      })
    )
    .next(
      capturePayment.withInput({
        order: authorizePayment.output.order,
        amountCents: authorizePayment.output.amountCents,
      }),
      {
        when: authorizePayment.output.amountCents.greaterThanOrEqualTo(1000),
        message: "Payment not authorized",
      }
    )
    .next(
      createShipment.withInput({
        order: input.order,
        shippingAddress: input.order.shippingAddress,
        email: input.order.customerEmail,
      }),
      {
        when: capturePayment.output.amountCents.greaterThanOrEqualTo(1000),
        message: "Payment capture failed",
      }
    )
    .next(
      sendConfirmation.withInput({
        order: createShipment.output.order,
        email: createShipment.output.email,
        grandTotal: createShipment.output.order.total,
      })
    );
}