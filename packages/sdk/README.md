# @neuron/sdk

A TypeScript SDK for **declaratively defining composable, executable systems**. You describe services, their inputs and outputs, how they connect, and the overall flow of execution — the SDK compiles that description into a portable, JSON-serializable **manifest** that an execution engine later runs.

The SDK is a **definition layer**, not a runtime. It produces a manifest that describes *what* your system does; the actual execution is handled by whatever runtime consumes the manifest.

## Installation

```bash
pnpm add @neuron/sdk
```

## Quick Start

The smallest complete system:

```ts
import { Service, System } from "@neuron/sdk";

const hello = Service({ name: "hello" })
  .executor({ name: "set" });

const manifest = System({
  name: "my-system",
  version: "1.0.0",
})
  .run(hello)
  .toManifest();

console.log(JSON.stringify(manifest, null, 2));
```

`toManifest()` throws if the system is missing a version, has no definition, or references a service that was never defined.

## Core Concepts

### Service

A `Service` is a unit of work. It has an **identity** (name, version, optional description), an **executor** (what will actually run it), and — optionally — typed **input/output contracts**.

```ts
import { Service } from "@neuron/sdk";

const validateOrder = Service({
  name: "validate-order",
  version: "1.0.0",
  description: "Validate incoming order request",
}).executor({ name: "set" });
```

#### Identity

Declare the service's name, version, and description once, up front:

```ts
Service({
  name: "http.get",
  version: "1.0.0",
  description: "Fetch a resource over HTTP",
})
```

#### Executor

An executor is what actually runs the service. By default, the executor name equals the service name, with version `latest` and a local registry:

```ts
// Default: executor is { name: "http.get", version: "latest", registry: "local" }
Service({ name: "http.get" })
```

Override the executor explicitly when you need a specific one:

```ts
Service({ name: "github.read" }).executor({
  name: "http.get",
  version: "^1.2.0",
  registry: "official",
})
```

#### Input and output contracts

You can declare typed contracts, which give you full IDE type-safety (autocomplete, type-checked binding, instant errors for wrong field names or types):

```ts
interface GitHubReadInput {
  owner: string;
  repository: string;
  path: string;
}

interface GitHubReadOutput {
  content: string;
  path: string;
  sha: string;
}

const githubRead = Service({
  name: "github.read",
  version: "1.0.0",
})
  .inputSchema<GitHubReadInput>()
  .outputSchema<GitHubReadOutput>();
```

You can also describe contracts with runtime validation rules, which also populate the manifest:

```ts
import { string, number, boolean } from "@neuron/sdk";

const createUser = Service({ name: "user.create" }).inputSchema({
  email: string().email().required(),
  age: number().min(18).max(120),
}).outputSchema({
  id: string().required(),
  active: boolean(),
});
```

### Expressions

The SDK exposes a typed, proxy-backed expression system for referring to a service's output or input fields, and for building conditions.

#### Referencing output fields

```ts
githubRead.output.content   // Expression<string>
githubRead.output.sha       // Expression<string>
githubRead.output.metadata.size  // Expression<number>
```

Autocomplete lists exactly the declared output fields — nothing else. Referencing an unknown field is an IDE error.

```ts
// @ts-expect-error — fileData is not a declared output
githubRead.output.fileData
```

#### Conditions

Expressions carry comparison operators that build guard conditions:

```ts
import { type Expression } from "@neuron/sdk";

const ready: Expression<boolean> =
  authorizePayment.output.paymentIntent.status.equals("requires_capture");

const valid: Expression<boolean> =
  githubRead.output.content.notEquals("");
```

Available operators: `equals`, `notEquals`, `greaterThan`, `greaterThanOrEqualTo`, `lessThan`, `lessThanOrEqualTo`, `and`, `or`.

### Binding input

**`withInput()`** binds a service's input fields to values or expressions. It is the primary way to feed data into a service.

#### Automatic style — pass expressions of a previous service's output

```ts
analyzeContent.withInput({
  content: githubRead.output.content,
  path: githubRead.output.path,
});
```

TypeScript checks every binding against the target's input schema. Wrong field names and wrong types are IDE errors:

```ts
// @ts-expect-error — sha is string but path expects a different type
analyzeContent.withInput({ content: githubRead.output.sha });
```

#### Literal values

```ts
githubRead.withInput({
  owner: "Muhammad-Jay",
  repository: "neuron",
  path: "README.md",
});
```

### Chaining services: `.next()`

**`.next()`** wires a service to the next one, producing a sequence. The original `ServiceDefinition` is used to start a chain, and the chain can continue with `.next()` on each step.

```ts
const pipeline = githubRead
  .next(
    analyzeContent.withInput({
      content: githubRead.output.content,
      path: githubRead.output.path,
    })
  )
  .next(saveResult);
```

#### Guard conditions

The second argument to `.next()` declares a condition that must hold before the next step runs:

```ts
authorizePayment
  .next(
    capturePayment.withInput({
      paymentIntentId: authorizePayment.output.paymentIntent.id,
    }),
    {
      when: authorizePayment.output.paymentIntent.status.equals("requires_capture"),
      message: "Payment not authorized",
    }
  )
```

### Explicit mapping: `.connect()`

When automatic field-by-field binding isn't enough (different field names, transformations, or referencing a source that isn't a direct previous step), use **`.connect()`** to define an explicit mapping from a source output to the target's input.

`.connect()` is fully typed: the callback receives the source output, and the returned object is checked against the target's input schema.

```ts
const githubToAnalyzer = analyzeContent.connect<GitHubReadOutput>((source) => ({
  content: source.output.content,
  path: source.output.path,
}));
```

You can then pass the connection to `.next()` or into another service's `.withInput()`.

Referencing a field that doesn't exist on the source is an IDE error:

```ts
analyzeContent.connect<GitHubReadOutput>((source) => ({
  // @ts-expect-error — source output has no fileData
  content: source.output.fileData,
  path: source.output.path,
}));
```

### Execution settings

**`.executionConfig()`** applies runtime execution policy to a specific invocation:

```ts
githubRead
  .withInput({ owner: "openai", repository: "neuron", path: "README.md" })
  .executionConfig({
    timeout: "30s",
    retries: 2,
  })
```

Supported options: `mode` (`"wait" | "detach"`), `timeout`, `retries`, `concurrency`, `continueOnFail`.

> Execution config is the **same shape for every service** — it describes runtime policy (timeouts, retries, concurrency), not service-specific business config.

### Parallel composition

**`Parallel()`** runs multiple branches concurrently. The parallel node completes when all branches complete:

```ts
import { Parallel } from "@neuron/sdk";

verify
  .next(
    Parallel(
      saveCustomer.withInput({ ... }),
      sendEmail.withInput({ ... }),
      updateAnalytics.withInput({ ... })
    )
  )
  .next(finish);
```

### System

A `System` aggregates a root composition and compiles it into a manifest. Services referenced in the tree are **auto-discovered** — there is no manual registration.

```ts
import { System } from "@neuron/sdk";

const system = System({
  name: "order-processing",
  version: "1.0.0",
  description: "Order processing pipeline",
});

const manifest = system
  .inputSchema<SystemInput>()
  .run(pipeline)
  .toManifest();
```

`toManifest()` walks the composition tree, collects every service definition, derives the connections between services, and produces the final `SystemManifest`. If a service is referenced in the tree but never defined, it throws.

#### Typed system input with `.withParams()`

Bind the system's execution input to the very first service in the chain:

```ts
System({ name: "order-processing", version: "1.0.0" })
  .inputSchema<SystemInput>()
  .withParams((input) =>
    validateOrder.withInput({
      order: input.order,
    })
  )
  .toManifest();
```

## The Manifest

`toManifest()` produces this shape:

```ts
interface SystemManifest {
  apiVersion: "neuron/v1";
  kind: "System";
  metadata: { name: string; version: string; description?: string };
  services: ServiceManifest[];
  inputs?: PortManifest[];
  connectors: ConnectorManifest[];
  definition: CompositionManifest;
}
```

- **`services`** — every service definition in the system (identity, executor, contracts).
- **`connectors`** — the derived connections between services, including field mappings and guard conditions. You don't author this list directly; the SDK derives it from `.next()` / `.withInput()` / `.connect()`.
- **`definition`** — the composition tree (`service` / `sequence` / `parallel`).

## CLI

The SDK ships a small CLI (`neuron-sdk`) to compile a project's entry into a manifest.

```bash
neuron-sdk build
```

It looks for a `neuron.config.ts` (or `.js` / `.mjs`) and uses the `entry` (default `index.ts`) as the manifest source. The import must default-export the compiled manifest. The resulting JSON is written to `.neuron/manifest.json`.

```bash
neuron-sdk version   # print the SDK version
neuron-sdk help      # show usage
```

### Configuration

```ts
// neuron.config.ts
import { defineConfig } from "@neuron/sdk";

export default defineConfig({
  entry: "./system.ts",
});
```

## Package Exports

Public API re-exported from `@neuron/sdk`:

- **Factories:** `Service`, `System`, `Parallel`, `connect`, `defineConfig`
- **Schema builders:** `string`, `number`, `boolean`, `list`, `record`
- **Types:** `ServiceDefinition`, `SystemDefinition`, `Expression`, `Expressionify`, `ExpressionArray`, `SourceContext`, `ExecutionContext`, `Connection`, `ExecutionConfig`, `InputBindings`, `InputValue`, `Composition`
- **Schema types:** `Schema`, `SchemaField`, `Infer`, `InferSchema`, `SchemaObject`, `StringField`, `NumberField`, `BooleanField`, `ListField`, `RecordField`, `FieldRules`
- **Manifest types:** `SystemManifest`, `ServiceManifest`, `ConnectorManifest`, `ConnectorMappingManifest`, `ConnectorValidationManifest`, `PortManifest`, `SystemNodeManifest`
- **Config types:** `NeuronConfig`

## Development

```bash
pnpm install          # link workspace
pnpm build:sdk        # build to dist/ via tsup
pnpm test:sdk         # run vitest
pnpm typecheck:sdk    # tsc --noEmit
```

Tests live in `packages/sdk/test/` and cover services, expressions, composition, connections, schemas, and full system manifests.
