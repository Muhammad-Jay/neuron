import { exec } from "@neuron/sdk";
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

// The pipeline wires services together. Each step consumes the output of the
// previous service (via `source.output.<field>`) or the original system input
// (via `exec.input.<field>`). The `.then()` second argument declares the guard
// condition that must hold before the next service runs.

export const pipeline = validateOrder
  .withInput({ order: exec.input.order })
  .then(parseOrder, {
    when: validateOrder.output.valid.eq(true),
    message: "Order validation failed",
  })
  .then(
    enrichCustomer.withInput({
      customerId: parseOrder.output.currency,
    })
  )
  .then(
    calculateTotals.withInput({
      items: exec.input.order.items,
      customerTier: enrichCustomer.output.customerData.tier,
      shippingState: enrichCustomer.output.customerData.shippingAddress.state,
      email: enrichCustomer.output.customerData.email,
    })
  )
  .then(
    authorizePayment.withInput({
      amountCents: exec.input.order.total,
      currency: exec.input.order.currency,
      email: enrichCustomer.output.customerData.email,
    })
  )
  .then(
    capturePayment.withInput({
      paymentIntentId: authorizePayment.output.paymentIntent.id,
    }),
    {
      when: authorizePayment.output.paymentIntent.status.eq("requires_capture"),
      message: "Payment not authorized",
    }
  )
  .then(
    createShipment.withInput({
      order: exec.input.order,
      shippingAddress: exec.input.order.shippingAddress,
      email: exec.input.order.customerEmail,
    }),
    {
      when: capturePayment.output.captureResult.status.eq("succeeded"),
      message: "Payment capture failed",
    }
  )
  .then(
    sendConfirmation.withInput({
      trackingNumber: createShipment.output.shipment.trackingNumber,
      carrier: createShipment.output.shipment.carrier,
      email: exec.input.order.customerEmail,
      grandTotal: exec.input.order.total,
    })
  );
