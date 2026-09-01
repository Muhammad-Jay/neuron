import { describe, it, expect } from "vitest";
import { System } from "../src/system.js";
import { Service } from "../src/service.js";
import { Parallel } from "../src/composition.js";
import { connect } from "../src/connection.js";
import { string, number, boolean, record } from "../src/schema.js";
import { createExecutionContext, createSourceContext } from "../src/expression.js";

describe("System", () => {
  it("throws when missing version", () => {
    const sys = System("test").run({ kind: "service", service: "a" });
    expect(() => sys.toManifest()).toThrow('System "test" is missing a version');
  });

  it("throws when missing definition", () => {
    const sys = System("test").version("1.0.0");
    expect(() => sys.toManifest()).toThrow('System "test" has no definition');
  });

  it("throws when service is not registered", () => {
    const a = Service("a");
    const sys = System("test")
      .version("1.0.0")
      .run(a.withInput({}));
    expect(() => sys.toManifest()).toThrow('Service "a" is referenced but not registered');
  });

  it("produces a basic manifest with one service", () => {
    const a = Service("a");
    const sys = System("test")
      .version("1.0.0")
      .register(a)
      .run(a.withInput({}));

    const manifest = sys.toManifest();
    expect(manifest.apiVersion).toBe("neuron/v1");
    expect(manifest.kind).toBe("System");
    expect(manifest.metadata).toEqual({ name: "test", version: "1.0.0" });
    expect(manifest.services).toHaveLength(1);
    expect(manifest.services[0]!.name).toBe("a");
    expect(manifest.connectors).toHaveLength(0);
    expect(manifest.definition).toEqual({ kind: "service", service: "a" });
  });

  it("produces a manifest with a sequence and empty connector (auto-passthrough)", () => {
    const a = Service("a");
    const b = Service("b");

    const sys = System("test")
      .version("1.0.0")
      .registerAll(a, b)
      .run(a.withInput({}).then(b.withInput({})));

    const manifest = sys.toManifest();
    expect(manifest.services).toHaveLength(2);
    expect(manifest.connectors).toHaveLength(1);
    expect(manifest.connectors[0]).toEqual({
      from: "a",
      to: "b",
      mappings: [],
      validations: [],
    });
  });

  it("auto-passthrough maps matching runtime schema fields", () => {
    const a = Service("a").outputSchema({
      content: string(),
      path: string(),
      ignored: number(),
    });
    const b = Service("b").inputSchema({
      content: string().required(),
      path: string().required(),
      count: number(),
    });

    const sys = System("test")
      .version("1.0.0")
      .registerAll(a, b)
      .run(a.withInput({}).then(b));

    expect(sys.toManifest().connectors[0].mappings).toEqual([
      { target: "content", expression: "source.output.content" },
      { target: "path", expression: "source.output.path" },
    ]);
  });

  it("produces a manifest with mappings from withInput bindings", () => {
    const a = Service("a").outputSchema({ data: string(), valid: boolean() });
    const b = Service("b").inputSchema({ data: string().required() });

    const sys = System("test")
      .version("1.0.0")
      .registerAll(a, b)
      .run(
        a.withInput({}).then(
          b.withInput({
            data: a.output.data,
          })
        )
      );

    const manifest = sys.toManifest();
    expect(manifest.connectors[0].mappings).toEqual([
      { target: "data", expression: "source.output.data" },
    ]);
  });

  it("produces validations from then() conditions", () => {
    const a = Service("a").outputSchema({ valid: boolean() });
    const b = Service("b");

    const sys = System("test")
      .version("1.0.0")
      .registerAll(a, b)
      .run(
        a.withInput({}).then(b.withInput({}), {
          when: a.output.valid.equals(true),
          message: "Failed",
        })
      );

    const manifest = sys.toManifest();
    expect(manifest.connectors[0].validations).toEqual([
      { expression: "source.output.valid == true", message: "Failed" },
    ]);
  });

  it("produces connectors for parallel branches", () => {
    const a = Service("a").outputSchema({ email: string() });
    const b = Service("b");
    const c = Service("c");

    const sys = System("test")
      .version("1.0.0")
      .registerAll(a, b, c)
      .run(
        a.withInput({}).then(
          Parallel(
            b.withInput({ email: a.output.email }),
            c.withInput({ email: a.output.email })
          )
        )
      );

    const manifest = sys.toManifest();
    expect(manifest.connectors).toHaveLength(2);
    expect(manifest.connectors[0].to).toBe("b");
    expect(manifest.connectors[1].to).toBe("c");
    expect(manifest.definition).toMatchObject({
      kind: "sequence",
      steps: [
        { kind: "service", service: "a" },
        {
          kind: "parallel",
          branches: [
            { kind: "service", service: "b" },
            { kind: "service", service: "c" },
          ],
        },
      ],
    });
  });

  it("deduplicates services referenced multiple times", () => {
    const a = Service("a");
    const b = Service("b");

    const sys = System("test")
      .version("1.0.0")
      .registerAll(a, b)
      .run(a.withInput({}).then(b.withInput({})).then(a.withInput({})));

    const manifest = sys.toManifest();
    expect(manifest.services).toHaveLength(2);
    expect(manifest.connectors).toHaveLength(2);
  });

  it("supports connect() connections with when() conditions", () => {
    const verify = Service("verify")
      .outputSchema({ verified: boolean(), id: string() });
    const save = Service("save").inputSchema({ id: string().required() });

    const source = createSourceContext<{ verified: boolean }>();
    const verifyToSave = connect<{ verified: boolean; id: string }, { id: string }>((src) => ({ id: src.output.id }))
      .when(source.output.verified.equals(true), "Not verified");

    const sys = System("test")
      .version("1.0.0")
      .registerAll(verify, save)
      .run(
        verify
          .withInput({})
          .then(save.withInput(verifyToSave))
      );

    const manifest = sys.toManifest();
    expect(manifest.connectors[0]).toEqual({
      from: "verify",
      to: "save",
      mappings: [{ target: "id", expression: "source.output.id" }],
      validations: [
        { expression: "source.output.verified == true", message: "Not verified" },
      ],
    });
  });
});

describe("System with full ecommerce pipeline", () => {
  it("produces the correct manifest for the order-processing pipeline", () => {
    const exec = createExecutionContext<{
      order: {
        items: unknown[];
        total: number;
        currency: string;
        customerEmail: string;
      };
    }>();
    const source = createSourceContext<{
      status: string;
      valid: boolean;
      currency: string;
      customerData: { tier: string; state: string; email: string };
      paymentIntent: { id: string; status: string };
      captureResult: { status: string };
      shipment: { trackingNumber: string; carrier: string };
    }>();
    const validateOrder = Service("validate-order")
      .executor("set", { status: "validated", valid: true })
      .inputSchema({ order: record().required() })
      .outputSchema({ valid: boolean(), status: string() });

    const parseOrder = Service("parse-order")
      .executor("set", { currency: "USD" })
      .inputSchema({ validationData: record().required() })
      .outputSchema({ currency: string(), items: record() });

    const enrichCustomer = Service("enrich-customer")
      .executor("set", { customer_data: { tier: "gold", email: "c@example.com" } })
      .inputSchema({ customerId: string().required() })
      .outputSchema({ customerData: record() });

    const calculateTotals = Service("calculate-totals")
      .executor("set", { tax_rate: 0.0825, discount_rate: 0.1 })
      .inputSchema({
        items: record().required(),
        customerTier: string(),
        shippingState: string(),
        email: string().email(),
      })
      .outputSchema({ total: number() });

    const authorizePayment = Service("authorize-payment")
      .executor("set", { payment_intent: { id: "pi_abc123", status: "requires_capture" } })
      .inputSchema({
        amountCents: number().required(),
        currency: string().required(),
        email: string().email(),
      })
      .outputSchema({ paymentIntent: record() });

    const capturePayment = Service("capture-payment")
      .executor("set", { capture_result: { status: "succeeded" } })
      .inputSchema({ paymentIntentId: string().required() })
      .outputSchema({ captureResult: record() });

    const createShipment = Service("create-shipment")
      .executor("set", { shipment: { tracking_number: "1Z999" } })
      .inputSchema({
        order: record().required(),
        email: string().email(),
      })
      .outputSchema({ shipment: record() });

    const sendConfirmation = Service("send-confirmation")
      .executor("set", { confirmation_sent: true })
      .inputSchema({
        trackingNumber: string(),
        carrier: string(),
        email: string().email(),
        grandTotal: number(),
      })
      .outputSchema({ confirmationSent: boolean() });

    const sys = System("order-processing")
      .version("1.0.0")
      .registerAll(
        validateOrder, parseOrder, enrichCustomer, calculateTotals,
        authorizePayment, capturePayment, createShipment, sendConfirmation
      )
      .run(
        validateOrder
          .withInput({ order: exec.input.order })
          .then(parseOrder.withInput({ validationData: source.output.status }), {
            when: source.output.valid.equals(true),
            message: "Order validation failed",
          })
          .then(enrichCustomer.withInput({ customerId: source.output.currency }))
          .then(calculateTotals.withInput({
            items: exec.input.order.items,
            customerTier: source.output.customerData.tier,
            shippingState: source.output.customerData.state,
            email: source.output.customerData.email,
          }))
          .then(authorizePayment.withInput({
            amountCents: exec.input.order.total,
            currency: exec.input.order.currency,
            email: source.output.customerData.email,
          }))
          .then(capturePayment.withInput({
            paymentIntentId: source.output.paymentIntent.id,
          }), {
            when: source.output.paymentIntent.status.equals("requires_capture"),
            message: "Payment not authorized",
          })
          .then(createShipment.withInput({
            order: exec.input.order,
            email: exec.input.order.customerEmail,
          }), {
            when: source.output.captureResult.status.equals("succeeded"),
            message: "Payment capture failed",
          })
          .then(sendConfirmation.withInput({
            trackingNumber: source.output.shipment.trackingNumber,
            carrier: source.output.shipment.carrier,
            email: source.output.customerData.email,
            grandTotal: exec.input.order.total,
          }))
      );

    const manifest = sys.toManifest();

    expect(manifest.apiVersion).toBe("neuron/v1");
    expect(manifest.kind).toBe("System");
    expect(manifest.metadata.name).toBe("order-processing");
    expect(manifest.metadata.version).toBe("1.0.0");
    expect(manifest.services).toHaveLength(8);
    expect(manifest.connectors).toHaveLength(7);
    expect(manifest.definition.kind).toBe("sequence");
  });
});
