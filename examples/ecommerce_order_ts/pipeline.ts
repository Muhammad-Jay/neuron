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
 * order can be bound to the very first service in the chain.
 */
export function buildPipeline(input: Expressionify<{ order: SystemInput["order"] }>) {
  return validateOrder
    .withInput({
      order: input.order,
    })
    .next(
      parseOrder.withInput({
        validationData: validateOrder.output,
      })
    )
    .next(
      enrichCustomer.withInput({
        customerId: parseOrder.output.order.customerId,
      })
    )
    .next(
      calculateTotals.withInput({
        items: parseOrder.output.order.items,
        customerTier: enrichCustomer.output.customerData.tier,
        shippingState: enrichCustomer.output.customerData.shippingAddress.state,
        email: enrichCustomer.output.customerData.email,
      })
    )
    .next(
      authorizePayment.withInput({
        amountCents: parseOrder.output.order.total,
        currency: parseOrder.output.order.currency,
        email: enrichCustomer.output.customerData.email,
      })
    )
    .next(
      capturePayment.withInput({
        paymentIntentId: authorizePayment.output.paymentIntent.id,
      }),
      {
        when: authorizePayment.output.paymentIntent.status.equals("requires_capture"),
        message: "Payment not authorized",
      }
    )
    .next(
      createShipment.withInput({
        order: parseOrder.output.order,
        shippingAddress: parseOrder.output.order.shippingAddress,
        email: enrichCustomer.output.customerData.email,
      }),
      {
        when: capturePayment.output.captureResult.status.equals("succeeded"),
        message: "Payment capture failed",
      }
    )
    .next(
      sendConfirmation.withInput({
        trackingNumber: createShipment.output.shipment.trackingNumber,
        carrier: createShipment.output.shipment.carrier,
        email: enrichCustomer.output.customerData.email,
        grandTotal: parseOrder.output.order.total,
      })
    );
}
