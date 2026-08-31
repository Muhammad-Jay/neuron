import {
  System,
  Service,
  Parallel,
  map,
  inputMapping,
  outputMapping,
  validate,
  source,
  exec,
} from "@neuron/sdk";

// --- Services ---

const validateOrder = Service("validate-order")
  .description("Validate incoming order request")
  .version("1.0.0")
  .executor("set", {
    status: "validated",
    valid: true,
  })
  .input({
    order: exec.input("order"),
  })
  .execution({ mode: "wait", timeout: "5s" });

const parseOrder = Service("parse-order")
  .description("Parse and normalize order data")
  .version("1.0.0")
  .executor("set", {
    currency: "USD",
  })
  .input({
    validation_data: source.output(),
  })
  .execution({ mode: "wait", timeout: "5s" });

const enrichCustomer = Service("enrich-customer")
  .description("Enrich with customer data")
  .version("1.0.0")
  .executor("set", {
    customer_data: {
      tier: "gold",
      email: "customer@example.com",
      shipping_address: {
        state: "CA",
        city: "San Francisco",
        zip: "94102",
      },
    },
  })
  .input({
    customer_id: source.output("validation_data.order.customer_id"),
  })
  .execution({ mode: "wait", timeout: "5s" });

const calculateTotals = Service("calculate-totals")
  .description("Calculate totals")
  .version("1.0.0")
  .executor("set", {
    tax_rate: 0.0825,
    discount_rate: 0.10,
  })
  .input({
    items: exec.input("order.items"),
    customer_tier: source.output("customer_data.tier"),
    shipping_state: source.output("customer_data.shipping_address.state"),
    email: source.output("customer_data.email"),
  })
  .execution({ mode: "wait", timeout: "5s" });

const authorizePayment = Service("authorize-payment")
  .description("Authorize payment")
  .version("1.0.0")
  .executor("set", {
    payment_intent: {
      id: "pi_abc123",
      status: "requires_capture",
    },
  })
  .input({
    amount_cents: exec.input("order.total"),
    currency: exec.input("order.currency"),
    email: source.output("email"),
  })
  .execution({ mode: "wait", timeout: "5s" });

const capturePayment = Service("capture-payment")
  .description("Capture payment")
  .version("1.0.0")
  .executor("set", {
    capture_result: {
      status: "succeeded",
      amount_received: 7997,
    },
  })
  .input({
    payment_intent_id: source.output("payment_intent.id"),
  })
  .execution({ mode: "wait", timeout: "5s" });

const createShipment = Service("create-shipment")
  .description("Create shipment")
  .version("1.0.0")
  .executor("set", {
    shipment: {
      tracking_number: "1Z999AA10123456784",
      carrier: "UPS",
      label_url: "https://shipping.example.com/label/1Z999AA10123456784",
    },
  })
  .input({
    order: exec.input("order"),
    shipping_address: exec.input("order.shipping_address"),
    email: exec.input("order.customer_email"),
  })
  .execution({ mode: "wait", timeout: "5s" });

const sendConfirmation = Service("send-confirmation")
  .description("Send confirmation email")
  .version("1.0.0")
  .executor("set", {
    confirmation_sent: true,
    message_id: "msg_xyz789",
  })
  .input({
    tracking_number: source.output("shipment.tracking_number"),
    carrier: source.output("shipment.carrier"),
    email: source.output("email"),
    grand_total: exec.input("order.total"),
  })
  .execution({ mode: "wait", timeout: "5s" });

// --- Pipeline ---

const pipeline = validateOrder
  .then(parseOrder, {
    mappings: [
      map("validation_data", source.output()),
    ],
    validations: [
      validate(source.output("valid").eq(true), "Order validation failed"),
    ],
  })
  .then(enrichCustomer, {
    mappings: [
      map("customer_id", source.output("validation_data.order.customer_id")),
    ],
  })
  .then(calculateTotals, {
    mappings: [
      map("items", exec.input("order.items")),
      map("customer_tier", source.output("customer_data.tier")),
      map("shipping_state", source.output("customer_data.shipping_address.state")),
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
      map("tracking_number", source.output("shipment.tracking_number")),
      map("carrier", source.output("shipment.carrier")),
      map("email", source.output("email")),
      map("grand_total", exec.input("order.total")),
    ],
  });

// --- System ---

const manifest = System("order-processing")
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

console.log(JSON.stringify(manifest, null, 2));
