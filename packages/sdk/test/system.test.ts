import { describe, it, expect } from "vitest";
import { System } from "../src/system.js";
import { Service, Input, Output } from "../src/service.js";
import { Parallel } from "../src/composition.js";
import { map } from "../src/mapping.js";
import { validate } from "../src/validation.js";
import { source, exec } from "../src/expr.js";

describe("System", () => {
  it("throws when missing version", () => {
    const sys = System("test").run({
      kind: "service",
      service: "a",
    });
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
      .run(a.node());
    expect(() => sys.toManifest()).toThrow(
      'Service "a" is referenced but not registered'
    );
  });

  it("produces a basic manifest with one service", () => {
    const a = Service("a");
    const sys = System("test")
      .version("1.0.0")
      .register(a)
      .run(a.node());

    const manifest = sys.toManifest();
    expect(manifest.apiVersion).toBe("neuron/v1");
    expect(manifest.kind).toBe("System");
    expect(manifest.metadata).toEqual({ name: "test", version: "1.0.0" });
    expect(manifest.services).toHaveLength(1);
    expect(manifest.services[0]!.name).toBe("a");
    expect(manifest.connectors).toHaveLength(0);
    expect(manifest.definition).toEqual({ kind: "service", service: "a" });
  });

  it("produces a manifest with a sequence", () => {
    const a = Service("a").version("1.0.0");
    const b = Service("b").version("1.0.0");

    const sys = System("test")
      .version("1.0.0")
      .registerAll(a, b)
      .run(a.then(b));

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

  it("produces a manifest with connectors and mappings", () => {
    const a = Service("a");
    const b = Service("b");

    const sys = System("test")
      .version("1.0.0")
      .registerAll(a, b)
      .run(
        a.then(b, {
          mappings: [map("data", source.output())],
          validations: [
            validate(source.output("valid").eq(true), "Failed"),
          ],
        })
      );

    const manifest = sys.toManifest();
    expect(manifest.connectors[0]).toEqual({
      from: "a",
      to: "b",
      mappings: [{ target: "data", expression: "source.output" }],
      validations: [
        { expression: "source.output.valid == true", message: "Failed" },
      ],
    });
  });

  it("produces a manifest with parallel branches", () => {
    const a = Service("a");
    const b = Service("b");
    const c = Service("c");

    const sys = System("test")
      .version("1.0.0")
      .registerAll(a, b, c)
      .run(a.then(Parallel(b, c)));

    const manifest = sys.toManifest();
    expect(manifest.services).toHaveLength(3);
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

  it("collects services in order from tree", () => {
    const a = Service("a");
    const b = Service("b");
    const c = Service("c");

    const sys = System("test")
      .version("1.0.0")
      .registerAll(a, b, c)
      .run(a.then(Parallel(b, c)));

    const manifest = sys.toManifest();
    expect(manifest.services.map((s) => s.name)).toEqual(["a", "b", "c"]);
  });

  it("deduplicates services", () => {
    const a = Service("a");
    const b = Service("b");

    const sys = System("test")
      .version("1.0.0")
      .registerAll(a, b)
      .run(a.then(b).then(a));

    const manifest = sys.toManifest();
    expect(manifest.services).toHaveLength(2);
  });

  it("sets description and config", () => {
    const sys = System("test")
      .version("1.0.0")
      .description("A test system")
      .config({ environment: "development" })
      .run({ kind: "service", service: "a" })
      .register(Service("a"));

    const manifest = sys.toManifest();
    expect(manifest.metadata.description).toBe("A test system");
  });
});

describe("System with full ecommerce pipeline", () => {
  it("produces the correct manifest for 8-service pipeline", () => {
    const validateOrder = Service("validate-order")
      .executor("set")
      .config({ status: "validated", valid: true })
      .input({ order: exec.input("order") })
      .execution({ timeout: "5s" });

    const parseOrder = Service("parse-order")
      .executor("set")
      .config({ currency: "USD" })
      .input({ validation_data: source.output("validation_data") })
      .execution({ timeout: "5s" });

    const enrichCustomer = Service("enrich-customer")
      .executor("set")
      .config({
        customer_data: { tier: "gold", email: "customer@example.com" },
      })
      .input({ customer_id: source.output("validation_data.order.customer_id") })
      .execution({ timeout: "5s" });

    const calculateTotals = Service("calculate-totals")
      .executor("set")
      .config({ tax_rate: 0.0825, discount_rate: 0.1 })
      .input({
        items: exec.input("order.items"),
        customer_tier: source.output("customer_data.tier"),
        shipping_state: source.output("customer_data.shipping_address.state"),
        email: source.output("customer_data.email"),
      })
      .execution({ timeout: "5s" });

    const authorizePayment = Service("authorize-payment")
      .executor("set")
      .config({ payment_intent: { id: "pi_abc123", status: "requires_capture" } })
      .input({
        amount_cents: exec.input("order.total"),
        currency: exec.input("order.currency"),
        email: source.output("email"),
      })
      .execution({ timeout: "5s" });

    const capturePayment = Service("capture-payment")
      .executor("set")
      .config({ capture_result: { status: "succeeded", amount_received: 7997 } })
      .input({ payment_intent_id: source.output("payment_intent.id") })
      .execution({ timeout: "5s" });

    const createShipment = Service("create-shipment")
      .executor("set")
      .config({
        shipment: {
          tracking_number: "1Z999AA10123456784",
          carrier: "UPS",
        },
      })
      .input({
        order: exec.input("order"),
        shipping_address: exec.input("order.shipping_address"),
        email: exec.input("order.customer_email"),
      })
      .execution({ timeout: "5s" });

    const sendConfirmation = Service("send-confirmation")
      .executor("set")
      .config({ confirmation_sent: true, message_id: "msg_xyz789" })
      .input({
        tracking_number: source.output("shipment.tracking_number"),
        carrier: source.output("shipment.carrier"),
        email: source.output("email"),
      })
      .execution({ timeout: "5s" });

    const sys = System("order-processing")
      .version("1.0.0")
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
      .run(
        validateOrder
          .then(parseOrder, {
            mappings: [map("validation_data", source.output())],
            validations: [
              validate(source.output("valid").eq(true), "Order validation failed"),
            ],
          })
          .then(enrichCustomer, {
            mappings: [
              map(
                "customer_id",
                source.output("validation_data.order.customer_id")
              ),
            ],
          })
          .then(calculateTotals, {
            mappings: [
              map("items", exec.input("order.items")),
              map("customer_tier", source.output("customer_data.tier")),
              map(
                "shipping_state",
                source.output("customer_data.shipping_address.state")
              ),
              map("email", source.output("customer_data.email")),
            ],
          })
          .then(authorizePayment, {
            mappings: [
              map("amount_cents", exec.input("order.total")),
              map("currency", exec.input("order.currency")),
              map("email", source.output("email")),
            ],
          })
          .then(capturePayment, {
            mappings: [
              map("payment_intent_id", source.output("payment_intent.id")),
            ],
            validations: [
              validate(
                source.output("payment_intent.status").eq("requires_capture"),
                "Payment not authorized"
              ),
            ],
          })
          .then(createShipment, {
            mappings: [
              map("order", exec.input("order")),
              map("shipping_address", exec.input("order.shipping_address")),
              map("email", exec.input("order.customer_email")),
            ],
            validations: [
              validate(
                source.output("capture_result.status").eq("succeeded"),
                "Payment capture failed"
              ),
            ],
          })
          .then(sendConfirmation, {
            mappings: [
              map(
                "tracking_number",
                source.output("shipment.tracking_number")
              ),
              map("carrier", source.output("shipment.carrier")),
              map("email", source.output("email")),
              map("grand_total", exec.input("order.total")),
            ],
          })
      );

    const manifest = sys.toManifest();

    expect(manifest.apiVersion).toBe("neuron/v1");
    expect(manifest.kind).toBe("System");
    expect(manifest.metadata.name).toBe("order-processing");
    expect(manifest.metadata.version).toBe("1.0.0");

    expect(manifest.services).toHaveLength(8);
    expect(manifest.services.map((s) => s.name)).toEqual([
      "validate-order",
      "parse-order",
      "enrich-customer",
      "calculate-totals",
      "authorize-payment",
      "capture-payment",
      "create-shipment",
      "send-confirmation",
    ]);

    expect(manifest.connectors).toHaveLength(7);

    expect(manifest.connectors[0]).toEqual({
      from: "validate-order",
      to: "parse-order",
      mappings: [{ target: "validation_data", expression: "source.output" }],
      validations: [
        {
          expression: "source.output.valid == true",
          message: "Order validation failed",
        },
      ],
    });

    expect(manifest.connectors[4]).toEqual({
      from: "authorize-payment",
      to: "capture-payment",
      mappings: [
        { target: "payment_intent_id", expression: "source.output.payment_intent.id" },
      ],
      validations: [
        {
          expression: "source.output.payment_intent.status == 'requires_capture'",
          message: "Payment not authorized",
        },
      ],
    });

    expect(manifest.definition.kind).toBe("sequence");
  });
});
