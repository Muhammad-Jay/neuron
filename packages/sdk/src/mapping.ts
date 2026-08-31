import type { Expression } from "./expr.js";

export interface ConnectorMappingManifest {
  target: string;
  expression: string;
}

export interface ServiceMappingManifest {
  direction: "input" | "output";
  source: string;
  target: string;
}

/**
 * Creates a connector mapping for use in `.then()` options.
 *
 * The expression should be built using `source` or `exec` helpers from `@neuron/sdk/expr`.
 *
 * @example
 * ```ts
 * map("validation_data", source.output())
 * // → { target: "validation_data", expression: "source.output" }
 *
 * map("customer_id", source.output("validation_data.order.customer_id"))
 * // → { target: "customer_id", expression: "source.output.validation_data.order.customer_id" }
 * ```
 */
export function map(target: string, expression: Expression): ConnectorMappingManifest {
  return { target, expression: expression.toString() };
}

/**
 * Creates a service-level input mapping.
 */
export function inputMapping(source: string, target: string): ServiceMappingManifest {
  return { direction: "input", source, target };
}

/**
 * Creates a service-level output mapping.
 */
export function outputMapping(source: string, target: string): ServiceMappingManifest {
  return { direction: "output", source, target };
}