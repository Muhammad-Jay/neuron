import type {
  PortManifest,
  ServiceManifest,
  ServiceMappingManifest,
  ServiceValidationManifest,
  SystemNodeManifest,
} from "./manifest.js";
import type { Expression } from "./expr.js";
import type { ConnectorMappingManifest } from "./mapping.js";
import type { ConnectorValidationManifest } from "./validation.js";

export type InputBinding = Expression | string;

export class ServiceDefinition<
  TInput extends Record<string, unknown> = Record<string, unknown>,
  TOutput extends Record<string, unknown> = Record<string, unknown>,
  TConfig extends Record<string, unknown> = Record<string, unknown>,
> {
  private _executorType: string;
  private _executorVersion?: string;
  private _executorSource?: string;
  private _executorConfig?: Record<string, unknown>;
  private _config: Record<string, unknown>;
  private _inputs: PortManifest[];
  private _outputs: PortManifest[];
  private _inputMappings: ServiceMappingManifest[];
  private _inputValidations: ServiceValidationManifest[];
  private _execution: ServiceManifest["execution"];
  private _version?: string;
  private _description?: string;

  constructor(
    readonly ref: string,
    opts?: {
      executorType?: string;
      executorVersion?: string;
      executorSource?: string;
      executorConfig?: Record<string, unknown>;
      config?: Record<string, unknown>;
      inputs?: PortManifest[];
      outputs?: PortManifest[];
      inputMappings?: ServiceMappingManifest[];
      inputValidations?: ServiceValidationManifest[];
      execution?: ServiceManifest["execution"];
      version?: string;
      description?: string;
    }
  ) {
    this._executorType = opts?.executorType ?? ref;
    this._executorVersion = opts?.executorVersion;
    this._executorSource = opts?.executorSource;
    this._executorConfig = opts?.executorConfig;
    this._config = opts?.config ?? {};
    this._inputs = opts?.inputs ? [...opts.inputs] : [];
    this._outputs = opts?.outputs ? [...opts.outputs] : [];
    this._inputMappings = opts?.inputMappings ? [...opts.inputMappings] : [];
    this._inputValidations = opts?.inputValidations
      ? [...opts.inputValidations]
      : [];
    this._execution = opts?.execution ?? {};
    this._version = opts?.version;
    this._description = opts?.description;
  }

  /** Set the executor type and optional executor-specific config (e.g. timeout, retries). */
  executor(type: string, config?: Record<string, unknown>): this {
    this._executorType = type;
    this._executorConfig = config;
    return this;
  }

  /** Set the executor version. */
  executorVersion(version: string): this {
    this._executorVersion = version;
    return this;
  }

  /** Set the executor source/registry. */
  executorSource(source: string): this {
    this._executorSource = source;
    return this;
  }

  /** Set the service version. */
  version(version: string): this {
    this._version = version;
    return this;
  }

  /** Set the service description. */
  description(description: string): this {
    this._description = description;
    return this;
  }

  /** Set the service configuration. */
  config(config: TConfig): this {
    this._config = config as Record<string, unknown>;
    return this;
  }

  /** Declare input ports and bind expressions to them. */
  input(bindings: { [K in keyof TInput]?: InputBinding }): this {
    for (const [name, value] of Object.entries(bindings)) {
      if (value === undefined) continue;

      this._inputs.push({
        name,
        type: "any",
        required: true,
      });

      const sourceExpr = String(value);
      this._inputMappings.push({
        direction: "input",
        source: sourceExpr,
        target: name,
      });
    }
    return this;
  }

  /** Declare input ports explicitly (for installable packages). */
  inputs(...ports: PortManifest[]): this {
    this._inputs.push(...ports);
    return this;
  }

  /** Declare output ports explicitly (for installable packages). */
  outputs(...ports: PortManifest[]): this {
    this._outputs.push(...ports);
    return this;
  }

  /** Declare output port names. */
  output(...names: string[]): this {
    for (const name of names) {
      this._outputs.push({ name, type: "any", required: false });
    }
    return this;
  }

  /** Add a service-level input validation rule. */
  validateInput(field: string, rules: Record<string, unknown>): this {
    this._inputValidations.push({ field, rules });
    return this;
  }

  /** Set execution settings. */
  execution(settings: {
    mode?: string;
    timeout?: string;
    retries?: number;
    concurrency?: number;
    continueOnFail?: boolean;
  }): this {
    this._execution = settings;
    return this;
  }

  /**
   * Chain this service to the next service or composition node.
   *
   * Optionally specify connector mappings and validations for this transition.
   *
   * @example
   * ```ts
   * verify.then(save, {
   *   mappings: [map("data", source.output())],
   *   validations: [validate(source.output("valid").eq(true), "Failed")],
   * })
   * ```
   */
  then(
    next: ServiceDefinition<any, any, any> | SystemNodeManifest,
    options?: {
      mappings?: ConnectorMappingManifest[];
      validations?: ConnectorValidationManifest[];
    }
  ): SequenceNode {
    const nextNode =
      next instanceof ServiceDefinition ? next.node() : next;

    const steps: SystemNodeManifest[] = [
      this.node(),
      ...("steps" in nextNode ? nextNode.steps : [nextNode]),
    ];

    const result = new SequenceNodeImpl(steps);

    if (options && (options.mappings || options.validations)) {
      result._connectors.push({
        from: this.ref,
        mappings: options.mappings ?? [],
        validations: options.validations ?? [],
      });
    }

    return result;
  }

  /** Convert to a node reference (used in composition trees). */
  node(): SystemNodeManifest {
    return { kind: "service", service: this.ref };
  }

  /** Convert to the service manifest representation. */
  toManifest(): ServiceManifest {
    return {
      name: this.ref,
      version: this._version,
      description: this._description,
      executor: {
        type: this._executorType,
        version: this._executorVersion,
        source: this._executorSource,
        config: this._executorConfig,
      },
      inputs: this._inputs,
      outputs: this._outputs,
      mappings: this._inputMappings,
      validations: this._inputValidations,
      config: this._config,
      execution: this._execution,
    };
  }
}

export interface ConnectorData {
  from: string;
  mappings: ConnectorMappingManifest[];
  validations: ConnectorValidationManifest[];
}

export class SequenceNodeImpl {
  readonly kind = "sequence" as const;
  _connectors: ConnectorData[] = [];

  toJSON() {
    return { kind: this.kind, steps: this.steps };
  }

  constructor(
    readonly steps: SystemNodeManifest[],
    existingConnectors?: ConnectorData[]
  ) {
    if (existingConnectors) {
      this._connectors = [...existingConnectors];
    }
  }

  /**
   * Continue the sequence by chaining another service or composition node.
   *
   * @example
   * ```ts
   * a.then(b).then(c)
   * a.then(Parallel(b, c)).then(d)
   * ```
   */
  then(
    next: ServiceDefinition<any, any, any> | SystemNodeManifest,
    options?: {
      mappings?: ConnectorMappingManifest[];
      validations?: ConnectorValidationManifest[];
    }
  ): SequenceNodeImpl {
    const nextNode =
      next instanceof ServiceDefinition ? next.node() : next;

    const newSteps: SystemNodeManifest[] = [
      ...this.steps,
      ...("steps" in nextNode ? nextNode.steps : [nextNode]),
    ];

    const result = new SequenceNodeImpl(newSteps, this._connectors);

    if (options && (options.mappings || options.validations)) {
      const lastService = findLastServiceRef(this.steps);
      if (lastService) {
        result._connectors.push({
          from: lastService,
          mappings: options.mappings ?? [],
          validations: options.validations ?? [],
        });
      }
    }

    return result;
  }
}

function findLastServiceRef(steps: SystemNodeManifest[]): string | null {
  for (let i = steps.length - 1; i >= 0; i--) {
    const step = steps[i]!;
    if (step.kind === "service") return step.service;
  }
  return null;
}

export type SequenceNode = SequenceNodeImpl;

/**
 * Create a typed service definition.
 *
 * @typeParam TInput  - The shape of data this service expects as input.
 * @typeParam TOutput - The shape of data this service produces as output.
 * @typeParam TConfig - The shape of the service configuration.
 *
 * @example
 * ```ts
 * interface VerifyInput {
 *   customer_id: string;
 *   email: string;
 * }
 *
 * interface VerifyOutput {
 *   verified: boolean;
 *   tier: string;
 * }
 *
 * const verifyCustomer = Service<VerifyInput, VerifyOutput>("customer.verify")
 *   .executor("http.post", { timeout: 30 })
 *   .input({
 *     customer_id: exec.input("order.customer_id"),
 *     email: exec.input("order.email"),
 *   })
 *   .output("verified", "tier", "email");
 * ```
 */
export function Service<
  TInput extends Record<string, unknown> = Record<string, unknown>,
  TOutput extends Record<string, unknown> = Record<string, unknown>,
  TConfig extends Record<string, unknown> = Record<string, unknown>,
>(ref: string): ServiceDefinition<TInput, TOutput, TConfig> {
  return new ServiceDefinition<TInput, TOutput, TConfig>(ref);
}

/** Create a required port declaration. */
export function Input(
  name: string,
  type: PortManifest["type"] = "any",
  required = true
): PortManifest {
  return { name, type, required };
}

/** Create an output port declaration. */
export function Output(
  name: string,
  type: PortManifest["type"] = "any"
): PortManifest {
  return { name, type, required: false };
}
