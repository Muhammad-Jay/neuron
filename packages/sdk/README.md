# @neuron/sdk

A TypeScript SDK for defining **Neuron systems** declaratively. It composes services, connectors, mappings, and validations into a typed, JSON serializable `SystemManifest` that the Neuron runtime (Go parser/compiler) consumes.

The SDK is a **system definition language**, not a runtime — it produces a manifest that describes *what* your system does; the N.O.R.E. runtime is responsible for *executing* it.

## Architecture

```
┌─────────────────────┐      ┌──────────────────────────┐      ┌──────────────────┐
│  TypeScript SDK      │      │  SystemManifest (JSON)   │      │  Go parser/nore   │
│  (this package)      │ ───► │  services + connectors   │ ───► │  core.System       │
│  declarative API     │      │  + definition tree       │      │  N.O.R.E. runtime  │
└─────────────────────┘      └──────────────────────────┘      └──────────────────┘
```

The SDK compiles your composition tree and produces a manifest that contains:

- **`services`** — a flattened list of every service definition
- **`connectors`** — a flattened list of every connection between services, including mappings and validations
- **`definition`** — the composition tree (sequence / parallel / service nodes)

## Installation

```bash
pnpm add @neuron/sdk
```

This SDK is part of a pnpm workspace. To build and run its tests:

```bash
pnpm build:sdk       # builds packages/sdk → dist/
pnpm test:sdk        # runs vitest
pnpm typecheck:sdk   # runs tsc --noEmit
```

## Quick Start

The smallest complete system:

```ts
import { System, Service } from "@neuron/sdk";

const hello = Service("hello")
  .executor("set", { message: "Hello, world!" });

const manifest = System("my-system")
  .version("1.0.0")
  .register(hello)
  .run(hello)
  .toManifest();

console.log(JSON.stringify(manifest, null, 2));
```

`toManifest()` throws if the system is missing a version, has no definition, or references a service that wasn't registered.

## Core Concepts

### Service

A `Service` represents a unit of work with an **executor**, **configuration**, **input/output ports**, and **mappings**. Create one with the `Service(ref)` factory:

```ts
import { Service, exec, source } from "@neuron/sdk";

const validateOrder = Service("validate-order")
  .description("Validate incoming order request")
  .version("1.0.0")
  .executor("set", { status: "validated", valid: true })
  .input({ order: exec.input("order") })
  .output("valid", "status")
  .execution({ mode: "wait", timeout: "5s" });
```

#### Typed services

`Service` is generic over input, output, and config shapes for type safety:

```ts
interface VerifyInput { customer_id: string; email: string; }
interface VerifyOutput { verified: boolean; tier: string; }

const verify = Service<VerifyInput, VerifyOutput>("customer.verify")
  .executor("http.post", { timeout: 30 })
  .input({
    customer_id: exec.input("order.customer_id"),
    email: exec.input("order.email"),
  })
  .output("verified", "tier");
```

#### Service builder methods

| Method | Description |
| --- | --- |
| `.executor(type, config?)` | Set the executor type and optional executor config (timeout, retries, etc.). Set in the manifest's `executor.config`. |
| `.executorVersion(v)` | Set the executor version. |
| `.executorSource(s)` | Set the executor source/registry. |
| `.version(v)` | Set the service version. |
| `.description(d)` | Set the service description. |
| `.config(cfg)` | Set arbitrary service config. Typed via the `TConfig` generic. |
| `.input(bindings)` | Declare input ports and bind each to an expression. Each entry adds an input port and an input mapping. |
| `.inputs(...ports)` | Declare input ports explicitly as `PortManifest`s (for installable packages). |
| `.outputs(...ports)` | Declare output ports explicitly as `PortManifest`s. |
| `.output(...names)` | Declare output port names (each of type `any`). |
| `.validateInput(field, rules)` | Add a service-level input validation rule. |
| `.execution(settings)` | Set runtime execution settings: `mode`, `timeout`, `retries`, `concurrency`, `continueOnFail`. |
| `.then(next, options?)` | Chain this service to the next service/node. See [Composition](#composition). |
| `.node()` | Return a `SystemNodeManifest` referencing this service. |
| `.toManifest()` | Return the `ServiceManifest` for this service. |

Note: `.executor(type, config)` sets the executor config, while `.config()` sets the service configuration — they are separate fields in the manifest.

### Expressions

Expressions are **CEL (Common Expression Language) strings** compatible with N.O.R.E.'s resolver. Two environment builders are provided: `source` and `exec`.

Expressions are type-safe (branded strings) but coerce to plain strings automatically in template literals / `String()`. They also carry comparison methods.

#### `source.*` — the upstream service's environment

Available inside connector mappings and validations. Represents the output of the service a connector originates from.

```ts
source.output("valid")                              // "source.output.valid"
source.output("validation_data.order.customer_id")  // "source.output.validation_data.order.customer_id"
source.input("order")                               // "source.input.order"
source.id                                           // "source.id"
source.name                                         // "source.name"
source.type                                         // "source.type"
source.metadata.name                                // "source.metadata.name"
source.metadata.version                             // "source.metadata.version"
```

#### `exec.*` — the current execution environment

Available in connector expressions and service configuration templates.

```ts
exec.input("order.items")    // "execution.input.order.items"
exec.input("order")          // "execution.input.order"
exec.id                      // "execution.id"
exec.correlationId           // "execution.correlation_id"
exec.blueprint.name          // "execution.blueprint.name"
```

#### Comparison methods

Every `ExpressionBuilder` (returned by `source.*` / `exec.*`) exposes comparison helpers that build complex CEL conditions:

```ts
source.output("valid").eq(true)                       // "source.output.valid == true"
source.output("status").neq("failed")                 // "source.output.status != 'failed'"
source.output("amount").gt(100)                       // "source.output.amount > 100"
source.output("amount").gte(100)                      // "source.output.amount >= 100"
source.output("amount").lt(100)                       // "source.output.amount < 100"
source.output("amount").lte(100)                      // "source.output.amount <= 100"

// boolean composition
const expr = source.output("valid").eq(true)
  .and(source.output("status").eq("ok"));             // "(source.output.valid == true) && (source.output.status == 'ok')"
```

Values are serialized: strings become `'quoted'`, numbers/booleans are literal, `null`/`undefined` become `null`.

### Mapping

Mappings wire data between services or from the execution context into a service's inputs.

#### Connector mappings (in `.then()`)

```ts
import { map, source, exec } from "@neuron/sdk";

map("customer_id", source.output("validation_data.order.customer_id"))
// → { target: "customer_id", expression: "source.output.validation_data.order.customer_id" }
```

#### Service-level mappings

```ts
import { inputMapping, outputMapping } from "@neuron/sdk";

inputMapping("execution.input.order", "order")   // → { direction: "input", source, target }
outputMapping("result", "output")                 // → { direction: "output", source, target }
```

### Validation

Two kinds of validation exist.

#### Connector validations (in `.then()`)

Guard a transition with a CEL condition. If the condition is false, the connector fails with the given message:

```ts
import { validate, source } from "@neuron/sdk";

validate(source.output("valid").eq(true), "Order validation failed")
// → { expression: "source.output.valid == true", message: "Order validation failed" }
```

#### Service-level input validation (`.validateInput`)

JSON Schema–like rule builders:

```ts
import { required, string, number, boolean, array, object } from "@neuron/sdk";

// Connector validation (CEL-based)
validate(source.output("valid").eq(true), "Order validation failed")

// Service input validation (schema-based)
service.validateInput("email", {
  ...string().email().min(5).toObject(),
  ...required(),
});
```

The schema builders:

| Builder | Methods |
| --- | --- |
| `required()` | No extra methods; `{ type: "required" }`. |
| `string()` | `.min(n)` → `minLength`, `.max(n)` → `maxLength`, `.pattern(re)` → `pattern`, `.email()`, `.uuid()`, `.toObject()`. |
| `number()` | `.min(n)` → `minimum`, `.max(n)` → `maximum`, `.exclusiveMin(n)`, `.exclusiveMax(n)`, `.integer()`, `.toObject()`. |
| `boolean()` | `{ type: "boolean" }`. |
| `array()` | `.minItems(n)`, `.maxItems(n)`, `.uniqueItems()`, `.toObject()`. |
| `object()` | `.required(props[])`, `.additionalProperties(bool)`, `.toObject()`. |

Each validator has `.toObject()` which returns a plain `ValidationRule` (`{ type, ... }`) you can spread into a rules object.

### Composition

Compose services into a **definition tree**. The two composition primitives are `.then()` (sequence) and `Parallel()`.

#### Sequence (`.then`)

Chain services linearly. Each `.then()` can carry connector `mappings` and `validations` for that specific transition.

```ts
const pipeline = validateOrder
  .then(parseOrder, {
    mappings: [map("validation_data", source.output())],
    validations: [validate(source.output("valid").eq(true), "Order validation failed")],
  })
  .then(enrichCustomer, {
    mappings: [map("customer_id", source.output("validation_data.order.customer_id"))],
  });
```

You can chain as many steps as you like — `.then().then().then()`:

```ts
const seq = a.then(b).then(c).then(d);
```

#### Parallel (`Parallel`)

Run multiple branches concurrently. The parallel node completes when all branches complete:

```ts
import { Parallel } from "@neuron/sdk";

verify.then(
  Parallel(
    saveCustomer,
    sendEmail,
    updateAnalytics
  )
).then(finish);
```

`Parallel` accepts service definitions or composition nodes as branches.

### System

A `System` aggregates registered services and a root composition node into a manifest.

```ts
import { System } from "@neuron/sdk";

const sys = System("order-processing")
  .version("1.0.0")
  .description("Order processing pipeline")
  .config({ environment: "development" })
  .register(validateOrder)
  .register(parseOrder)
  // or register them all at once:
  .registerAll(validateOrder, parseOrder, enrichCustomer)
  .run(pipeline)
  .toManifest();
```

#### System builder methods

| Method | Description |
| --- | --- |
| `.version(v)` | Set system version (required before `toManifest()`). |
| `.description(d)` | Set system description. |
| `.config(cfg)` | Set system-level configuration. |
| `.register(service)` | Register a service definition. |
| `.registerAll(...services)` | Register multiple services at once. |
| `.run(node)` | Set the root composition node (the `definition`). |
| `.toManifest()` | Compile to a `SystemManifest`. |

#### Auto-collection

Services are **auto-discovered** from the composition tree during `toManifest()`. Explicit registration via `.register()`/`.registerAll()` is not strictly required for services present in the tree — but it is a best practice (and required if you want the system to guarantee they exist). If a service is referenced in the tree but not registered, `toManifest()` throws.

Connectors are **auto-extracted** from `.then()` chains, so you don't have to declare the `connectors` list manually — the SDK walks the tree and derives `from`/`to` pairs, attaching the mappings/validations you passed to each `.then()`.

## The `SystemManifest`

`toManifest()` produces this shape:

```ts
interface SystemManifest {
  apiVersion: "neuron/v1";
  kind: "System";
  metadata: { name: string; version: string; description?: string };
  services: ServiceManifest[];    // flattened service definitions
  connectors: ConnectorManifest[]; // flattened connectors (from → to)
  definition: SystemNodeManifest;  // the composition tree
}
```

`SystemNodeManifest` is a discriminated union:

```ts
type SystemNodeManifest =
  | { kind: "service";  service: string }
  | { kind: "sequence"; steps: SystemNodeManifest[] }
  | { kind: "parallel"; branches: SystemNodeManifest[] };
```

## Complete Example

A full order-processing pipeline (see `examples/ecommerce_order_ts/`):

```ts
import {
  System, Service, Parallel, map, validate, source, exec,
} from "@neuron/sdk";

const validateOrder = Service("validate-order")
  .description("Validate incoming order request")
  .executor("set", { status: "validated", valid: true })
  .input({ order: exec.input("order") })
  .execution({ mode: "wait", timeout: "5s" });

const parseOrder = Service("parse-order")
  .executor("set", { currency: "USD" })
  .input({ validation_data: source.output() })
  .execution({ mode: "wait", timeout: "5s" });

const pipeline = validateOrder
  .then(parseOrder, {
    mappings: [map("validation_data", source.output())],
    validations: [validate(source.output("valid").eq(true), "Order validation failed")],
  });

const manifest = System("order-processing")
  .version("1.0.0")
  .description("Order processing pipeline")
  .registerAll(validateOrder, parseOrder)
  .run(pipeline)
  .toManifest();
```

## Package & Exports

All public API is re-exported from the package root (`@neuron/sdk`):

- `System`, `SystemDefinition`
- `Service`, `ServiceDefinition`, `Input`, `Output`, `SequenceNode`, `InputBinding`
- `Parallel`
- `source`, `exec`, `Expression`, `ExpressionBuilder`
- `map`, `inputMapping`, `outputMapping`
- `validate`, `required`, `string`, `number`, `boolean`, `array`, `object`
- Validation types: `ValidationRule`, `StringValidator`, `NumberValidator`, `ArrayValidator`, `ObjectValidator`
- Manifest types: `SystemManifest`, `ServiceManifest`, `ConnectorManifest`, `ConnectorMappingManifest`, `ConnectorValidationManifest`, `PortManifest`, `SystemNodeManifest`, `ServiceMappingManifest`, `ServiceValidationManifest`

## Development

```bash
pnpm install          # link workspace
pnpm build:sdk        # build to dist/ via tsup
pnpm test:sdk         # run vitest (90 tests)
pnpm typecheck:sdk    # tsc --noEmit
```

Tests live in `packages/sdk/test/` and cover expressions, services, composition, mapping, validation, and full system manifests.
