import type { ServiceDefinition } from "./service.js";
import type { SystemNodeManifest } from "./manifest.js";

export type SystemNode = ServiceDefinition<any, any, any> | SystemNodeManifest;

function toNode(value: SystemNode): SystemNodeManifest {
  if (value && typeof value === "object" && "node" in value && typeof (value as any).node === "function") {
    return (value as ServiceDefinition<any, any, any>).node();
  }
  return value as SystemNodeManifest;
}

/**
 * Compose multiple branches to run in parallel.
 *
 * All branches execute concurrently. The parallel node completes when
 * all branches complete.
 *
 * @example
 * ```ts
 * verify.then(
 *   Parallel(
 *     saveCustomer,
 *     sendEmail,
 *     updateAnalytics
 *   )
 * )
 * ```
 */
export function Parallel(
  ...branches: Array<ServiceDefinition<any, any, any> | SystemNodeManifest>
): SystemNodeManifest {
  return {
    kind: "parallel",
    branches: branches.map(toNode),
  };
}
