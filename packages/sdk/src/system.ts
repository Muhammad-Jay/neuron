import type {
  ConnectorManifest,
  SystemManifest,
} from "./manifest.js";
import { createExecutionContext, type Expressionify } from "./expression.js";
import type {
  Composition,
  SequenceComposition,
  ParallelComposition,
  ServiceComposition,
} from "./service.js";
import type { ServiceDefinition } from "./service.js";
import { schemaToManifest, type InferSchema, type SchemaObject } from "./schema.js";

type Runnable = Composition | { node(): Composition } | { _composition: Composition };

export class SystemDefinition<TInput extends object = object> {
  private _version?: string;
  private _description?: string;
  private _root?: Composition;
  private _services = new Map<string, ServiceDefinition<object, object>>();
  private _inputPorts: ReturnType<typeof schemaToManifest>["ports"] = [];

  constructor(readonly name: string) {}

  /** Set the system version. */
  version(version: string): this {
    this._version = version;
    return this;
  }

  /** Set the system description. */
  description(description: string): this {
    this._description = description;
    return this;
  }

  /** Set the root composition node. */
  inputSchema<T extends object>(): SystemDefinition<T>;
  inputSchema<S extends SchemaObject>(schema: S): SystemDefinition<InferSchema<S>>;
  inputSchema(schema?: SchemaObject): SystemDefinition<object> {
    const next = new SystemDefinition<object>(this.name);
    next._version = this._version;
    next._description = this._description;
    next._root = this._root;
    next._services = new Map(this._services);
    if (schema) {
      next._inputPorts = schemaToManifest(schema).ports;
    } else {
      next._inputPorts = [...this._inputPorts];
    }
    return next;
  }

  get input(): Expressionify<TInput> {
    return createExecutionContext<TInput>().input;
  }

  withParams(define: (input: Expressionify<TInput>) => Runnable): this {
    return this.run(define(this.input));
  }

  run(node: Runnable): this {
    if (node && typeof node === "object" && "_composition" in node) {
      this._root = node._composition;
    } else if (node && typeof node === "object" && "node" in node) {
      this._root = node.node();
    } else {
      this._root = node as Composition;
    }
    return this;
  }

  /**
   * Register a service definition with this system.
   */
  register(service: ServiceDefinition<object, object>): this {
    this._services.set(service.ref, service);
    return this;
  }

  /**
   * Register multiple service definitions.
   */
  registerAll(...services: Array<ServiceDefinition<object, object>>): this {
    for (const svc of services) {
      this._services.set(svc.ref, svc);
    }
    return this;
  }

  /**
   * Compile the system into a manifest.
   *
   * Traverses the composition tree, collects all referenced services,
   * extracts connectors from `.then()` chains, and produces a complete
   * `SystemManifest` (neuron/v1).
   */
  toManifest(): SystemManifest {
    if (!this._version) {
      throw new Error(`System "${this.name}" is missing a version`);
    }
    if (!this._root) {
      throw new Error(`System "${this.name}" has no definition`);
    }

    const refs = collectServiceRefs(this._root);
    const services = [];
    for (const ref of refs) {
      const def = this._services.get(ref);
      if (!def) {
        throw new Error(
          `Service "${ref}" is referenced but not registered. Call .register() or .registerAll() on the system.`
        );
      }
      services.push(def.toManifest());
    }

    const connectors = collectConnectors(this._root, this._services);

    return {
      apiVersion: "neuron/v1",
      kind: "System",
      metadata: {
        name: this.name,
        version: this._version,
        description: this._description,
      },
      inputs: this._inputPorts,
      services,
      connectors,
      definition: toManifestNode(this._root),
    } as SystemManifest;
  }
}

function toManifestNode(comp: Composition): SystemManifest["definition"] {
  switch (comp.kind) {
    case "service":
      return { kind: "service", service: comp.serviceRef };
    case "sequence":
      return { kind: "sequence", steps: comp.steps.map(toManifestNode) };
    case "parallel":
      return { kind: "parallel", branches: comp.branches.map(toManifestNode) };
  }
}

/**
 * Create a system definition.
 *
 * @example
 * ```ts
 * export default System("order-processing")
 *   .version("1.0.0")
 *   .run(validateOrder.then(parseOrder, {
 *     when: validateOrder.output.valid.equals(true),
 *     message: "Validation failed",
 *   }))
 *   .toManifest();
 * ```
 */
export function System(name: string): SystemDefinition {
  return new SystemDefinition(name);
}

// ---------------------------------------------------------------------------
// Tree traversal helpers
// ---------------------------------------------------------------------------

function collectServiceRefs(node: Composition): string[] {
  const refs: string[] = [];
  const seen = new Set<string>();
  walkServices(node, seen, refs);
  return refs;
}

function walkServices(
  node: Composition,
  seen: Set<string>,
  out: string[]
): void {
  switch (node.kind) {
    case "service":
      if (!seen.has(node.serviceRef)) {
        seen.add(node.serviceRef);
        out.push(node.serviceRef);
      }
      break;
    case "sequence":
      for (const step of node.steps) {
        walkServices(step, seen, out);
      }
      break;
    case "parallel":
      for (const branch of node.branches) {
        walkServices(branch, seen, out);
      }
      break;
  }
}

function collectConnectors(
  node: Composition,
  services: Map<string, ServiceDefinition<object, object>>
): ConnectorManifest[] {
  const connectors: ConnectorManifest[] = [];
  walkConnectors(node, connectors, services);
  return connectors;
}

function walkConnectors(
  node: Composition,
  out: ConnectorManifest[],
  services: Map<string, ServiceDefinition<object, object>>
): void {
  switch (node.kind) {
    case "sequence": {
      const seq = node as SequenceComposition;
      for (let i = 0; i < seq.steps.length - 1; i++) {
        const current = seq.steps[i]!;
        const next = seq.steps[i + 1]!;

        const fromRefs = getServiceRefsFlat(current);
        const nextServiceTargets = getTargetServiceComps(next);

        // Each predecessor service connects to each target service of the next step.
        for (const from of fromRefs) {
          for (const target of nextServiceTargets) {
            const mappings = connectorMappings(from, target, services);
            const validations = target.incomingConditions.map((c) => ({
              expression: c.expression,
              message: c.message ?? "",
            }));
            out.push({
              from,
              to: target.serviceRef,
              mappings,
              validations,
            });
          }
        }
      }

      // Recurse into nested steps (parallel branches etc.)
      for (const step of seq.steps) {
        walkNested(step, out, services);
      }
      break;
    }
    case "parallel": {
      const par = node as ParallelComposition;
      // Connectors are defined by sequences; a standalone parallel has none.
      for (const branch of par.branches) {
        walkNested(branch, out, services);
      }
      break;
    }
    case "service":
      break;
  }
}

function walkNested(
  node: Composition,
  out: ConnectorManifest[],
  services: Map<string, ServiceDefinition<object, object>>
): void {
  switch (node.kind) {
    case "sequence":
      walkConnectors(node, out, services);
      break;
    case "parallel":
      walkConnectors(node, out, services);
      break;
    default:
      break;
  }
}

function connectorMappings(
  fromRef: string,
  target: ServiceTarget,
  services: Map<string, ServiceDefinition<object, object>>
): ConnectorManifest["mappings"] {
  const explicit = Object.entries(target.bindings).map(([name, expression]) => ({
    target: name,
    expression,
  }));
  if (explicit.length > 0) return explicit;

  const from = services.get(fromRef)?.toManifest();
  const to = services.get(target.serviceRef)?.toManifest();
  if (!from || !to) return [];

  const outputs = new Map(from.outputs.map((port) => [port.name, port]));
  return to.inputs.flatMap((input) => {
    const output = outputs.get(input.name);
    if (!output || !compatiblePortTypes(output.type, input.type)) return [];
    return [{ target: input.name, expression: `source.output.${input.name}` }];
  });
}

function compatiblePortTypes(from: string, to: string): boolean {
  return from === to || from === "any" || to === "any";
}

/**
 * The set of service compositions that are direct targets of the NEXT step.
 * For a service leaf → itself. For a sequence → its first step's targets.
 * For a parallel → all branch entry targets.
 */
function getTargetServiceComps(node: Composition): ServiceTarget[] {
  if (node.kind === "service") {
    return [node];
  }
  if (node.kind === "sequence") {
    const steps = node.steps;
    if (steps.length === 0) return [];
    return getTargetServiceComps(steps[0]!);
  }
  if (node.kind === "parallel") {
    const targets: ServiceTarget[] = [];
    for (const branch of node.branches) {
      targets.push(...getTargetServiceComps(branch));
    }
    return targets;
  }
  return [];
}

interface ServiceTarget {
  serviceRef: string;
  bindings: Record<string, string>;
  incomingConditions: Array<{ expression: string; message?: string }>;
}

function getServiceRefsFlat(node: Composition): string[] {
  if (node.kind === "service") return [node.serviceRef];
  if (node.kind === "parallel") {
    const refs: string[] = [];
    for (const branch of node.branches) {
      refs.push(...getServiceRefsFlat(branch));
    }
    return refs;
  }
  if (node.kind === "sequence") {
    if (node.steps.length === 0) return [];
    return getServiceRefsFlat(node.steps[0]!);
  }
  return [];
}
