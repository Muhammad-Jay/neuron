import type {
  ConnectorManifest,
  ConnectorMappingManifest,
  ConnectorValidationManifest,
  SystemManifest,
  SystemNodeManifest,
} from "./manifest.js";
import type { ServiceDefinition, SequenceNodeImpl, ConnectorData } from "./service.js";

export class SystemDefinition {
  private _version?: string;
  private _description?: string;
  private _root?: SystemNodeManifest;
  private _services = new Map<string, ServiceDefinition<any, any, any>>();
  private _config: Record<string, unknown> = {};

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

  /** Set system-level configuration. */
  config(config: Record<string, unknown>): this {
    this._config = config;
    return this;
  }

  /** Set the root composition node. */
  run(node: SystemNodeManifest): this {
    this._root = node;
    return this;
  }

  /**
   * Register a service definition with this system.
   *
   * Services can be registered explicitly or auto-discovered from the
   * composition tree during `toManifest()`.
   */
  register(service: ServiceDefinition<any, any, any>): this {
    this._services.set(service.ref, service);
    return this;
  }

  /**
   * Register multiple service definitions.
   */
  registerAll(
    ...services: Array<ServiceDefinition<any, any, any>>
  ): this {
    for (const svc of services) {
      this._services.set(svc.ref, svc);
    }
    return this;
  }

  /**
   * Compile the system into a manifest.
   *
   * This traverses the composition tree, collects all referenced services,
   * extracts implicit connectors from `.then()` chains, and produces
   * a complete `SystemManifest`.
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

    const connectors = collectConnectors(this._root);

    return {
      apiVersion: "neuron/v1",
      kind: "System",
      metadata: {
        name: this.name,
        version: this._version,
        description: this._description,
      },
      services,
      connectors,
      definition: this._root,
    };
  }
}

/**
 * Create a system definition.
 *
 * @example
 * ```ts
 * export default System("order-processing")
 *   .version("1.0.0")
 *   .register(validateOrder)
 *   .register(parseOrder)
 *   .run(
 *     validateOrder.then(parseOrder, {
 *       mappings: [map("validation_data", source.output())],
 *       validations: [validate(source.output("valid").eq(true), "Failed")],
 *     })
 *   );
 * ```
 */
export function System(name: string): SystemDefinition {
  return new SystemDefinition(name);
}

// ---------------------------------------------------------------------------
// Tree traversal helpers
// ---------------------------------------------------------------------------

function collectServiceRefs(node: SystemNodeManifest): string[] {
  const refs: string[] = [];
  walkServices(node, new Set(), refs);
  return refs;
}

function walkServices(
  node: SystemNodeManifest,
  seen: Set<string>,
  out: string[]
): void {
  switch (node.kind) {
    case "service":
      if (!seen.has(node.service)) {
        seen.add(node.service);
        out.push(node.service);
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

function collectConnectors(node: SystemNodeManifest): ConnectorManifest[] {
  const connectors: ConnectorManifest[] = [];
  walkConnectors(node, connectors);
  return connectors;
}

function walkConnectors(
  node: SystemNodeManifest,
  out: ConnectorManifest[]
): void {
  if (node.kind === "sequence") {
    const seqNode = node as SequenceNodeImpl;

    for (let i = 0; i < node.steps.length - 1; i++) {
      const current = node.steps[i]!;
      const next = node.steps[i + 1]!;

      const fromRef = getServiceRef(current);
      const toRefs = getServiceRefsFlat(next);

      if (!fromRef) continue;

      const connectorData = seqNode._connectors.find(
        (c) => c.from === fromRef
      );

      for (const toRef of toRefs) {
        out.push({
          from: fromRef,
          to: toRef,
          mappings: connectorData?.mappings ?? [],
          validations: connectorData?.validations ?? [],
        });
      }
    }

    for (const step of node.steps) {
      walkConnectors(step, out);
    }
  } else if (node.kind === "parallel") {
    for (const branch of node.branches) {
      walkConnectors(branch, out);
    }
  }
}

function getServiceRef(node: SystemNodeManifest): string | null {
  if (node.kind === "service") return node.service;
  return null;
}

function getServiceRefsFlat(node: SystemNodeManifest): string[] {
  if (node.kind === "service") return [node.service];
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
