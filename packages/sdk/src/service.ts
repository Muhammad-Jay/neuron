import type { PortManifest, ServiceManifest, SystemNodeManifest } from "./manifest.js";
import {
  createExpressionProxy,
  createSourceContext,
  expressionToString,
  type Expression,
  type Expressionify,
  type SourceContext,
} from "./expression.js";
import { schemaToManifest, type FieldRules, type InferSchema, type SchemaObject } from "./schema.js";

export interface ExecutionConfig {
  mode?: "wait" | "detach";
  timeout?: string;
  retries?: number;
  concurrency?: number;
  continueOnFail?: boolean;
}

export interface ServiceComposition {
  kind: "service";
  serviceRef: string;
  bindings: Record<string, string>;
  execution?: ExecutionConfig;
  incomingConditions: Array<{ expression: string; message?: string }>;
}

export interface SequenceComposition {
  kind: "sequence";
  steps: Composition[];
}

export interface ParallelComposition {
  kind: "parallel";
  branches: Composition[];
}

export type Composition = ServiceComposition | SequenceComposition | ParallelComposition;

export type Condition = {
  when: Expression<boolean>;
  message?: string;
};

export type InputValue<T> = T | Expression<T> | Expressionify<Exclude<T, undefined>>;

type RequiredKeys<T extends object> = {
  [K in keyof T]-?: undefined extends T[K] ? never : K;
}[keyof T];

type OptionalKeys<T extends object> = Exclude<keyof T, RequiredKeys<T>>;

export type InputBindings<T extends object> = {
  [K in RequiredKeys<T>]: InputValue<T[K]>;
} & {
  [K in OptionalKeys<T>]?: InputValue<Exclude<T[K], undefined>>;
};

export interface Connection<TSource extends object = object, TTarget extends object = object> {
  readonly __kind: "connection";
  readonly __serviceRef?: string;
  readonly __bindings: Array<{ target: string; expression: string }>;
  readonly __conditions: Array<{ expression: string; message: string }>;
  when(condition: Expression<boolean>, message: string): Connection<TSource, TTarget>;
}

export interface ServiceReference<TInput extends object, TOutput extends object> {
  readonly ref: string;
  readonly output: Expressionify<TOutput>;
  withInput(bindings: InputBindings<TInput>): CompositionNode<TInput, TOutput>;
  withInput<TSource extends object>(connection: Connection<TSource, TInput>): CompositionNode<TInput, TOutput>;
  withExecution(execution: ExecutionConfig): CompositionNode<TInput, TOutput>;
  connect<TSource extends object>(
    define: (source: SourceContext<TSource>) => InputBindings<TInput>
  ): Connection<TSource, TInput>;
  when(condition: Expression<boolean>, message: string): CompositionNode<TInput, TOutput>;
}

type ServiceState = {
  executorType: string;
  executorVersion?: string;
  executorSource?: string;
  executorConfig?: Record<string, unknown>;
  version?: string;
  description?: string;
  inputPorts: PortManifest[];
  outputPorts: PortManifest[];
  inputRules: Record<string, FieldRules>;
};

export class CompositionNode<TInput extends object = object, TOutput extends object = object> {
  readonly __inputMarker?: TInput;
  readonly __outputMarker?: TOutput;
  readonly _composition: ServiceComposition;

  constructor(serviceRef: string, compositionOverride?: Partial<ServiceComposition>) {
    this._composition = {
      kind: "service",
      serviceRef,
      bindings: compositionOverride?.bindings ?? {},
      execution: compositionOverride?.execution,
      incomingConditions: compositionOverride?.incomingConditions ?? [],
    };
  }

  get serviceRef(): string {
    return this._composition.serviceRef;
  }

  get bindings(): Record<string, string> {
    return this._composition.bindings;
  }

  get executionConfig(): ExecutionConfig | undefined {
    return this._composition.execution;
  }

  get output(): Expressionify<TOutput> {
    return createExpressionProxy<TOutput>("source.output");
  }

  then<TNextInput extends object, TNextOutput extends object>(
    next:
      | ServiceDefinition<TNextInput, TNextOutput>
      | ServiceReference<TNextInput, TNextOutput>
      | CompositionNode<TNextInput, TNextOutput>
      | SequenceNode<TNextOutput>
      | ParallelCompositionHolder
      | Connection<TOutput, TNextInput>,
    condition?: Condition
  ): SequenceNode<TNextOutput> {
    return new SequenceNode(this._composition, toComposition(next, condition)) as SequenceNode<TNextOutput>;
  }

  when(condition: Expression<boolean>, message: string): CompositionNode<TInput, TOutput> {
    return new CompositionNode<TInput, TOutput>(this.serviceRef, {
      bindings: { ...this.bindings },
      execution: this.executionConfig,
      incomingConditions: [
        ...this._composition.incomingConditions,
        { expression: String(condition), message },
      ],
    });
  }

  node(): SystemNodeManifest {
    return toManifestNode(this._composition);
  }
}

export class SequenceNode<TOutput extends object = object> {
  readonly __outputMarker?: TOutput;
  readonly _composition: SequenceComposition;

  constructor(first: Composition, second?: Composition) {
    if (second) {
      this._composition = { kind: "sequence", steps: [first, ...splitSequence(second)] };
    } else if (first.kind === "sequence") {
      this._composition = first;
    } else {
      this._composition = { kind: "sequence", steps: [first] };
    }
  }

  get output(): Expressionify<TOutput> {
    return createExpressionProxy<TOutput>("source.output");
  }

  then<TNextInput extends object, TNextOutput extends object>(
    next:
      | ServiceDefinition<TNextInput, TNextOutput>
      | ServiceReference<TNextInput, TNextOutput>
      | CompositionNode<TNextInput, TNextOutput>
      | SequenceNode<TNextOutput>
      | ParallelCompositionHolder
      | Connection<TOutput, TNextInput>,
    condition?: Condition
  ): SequenceNode<TNextOutput> {
    return new SequenceNode<TNextOutput>({
      kind: "sequence",
      steps: [...this._composition.steps, ...splitSequence(toComposition(next, condition))],
    });
  }

  node(): SystemNodeManifest {
    return toManifestNode(this._composition);
  }
}

export class ParallelCompositionHolder {
  readonly _composition: ParallelComposition;

  constructor(branches: Composition[]) {
    this._composition = { kind: "parallel", branches };
  }

  node(): SystemNodeManifest {
    return toManifestNode(this._composition);
  }
}

export class ServiceDefinition<TInput extends object = object, TOutput extends object = object>
  implements ServiceReference<TInput, TOutput> {
  readonly ref: string;
  private readonly state: ServiceState;

  constructor(ref: string, state?: Partial<ServiceState>) {
    this.ref = ref;
    this.state = {
      executorType: state?.executorType ?? ref,
      executorVersion: state?.executorVersion,
      executorSource: state?.executorSource,
      executorConfig: state?.executorConfig,
      version: state?.version,
      description: state?.description,
      inputPorts: state?.inputPorts ?? [],
      outputPorts: state?.outputPorts ?? [],
      inputRules: state?.inputRules ?? {},
    };
  }

  executor(type: string | { type: string; version?: string; source?: string; config?: Record<string, unknown> }, config?: Record<string, unknown>, version?: string): this {
    if (typeof type === "string") {
      return this.clone({
        executorType: type,
        executorConfig: config,
        executorVersion: version ?? this.state.executorVersion,
      }) as this;
    }
    return this.clone({
      executorType: type.type,
      executorVersion: type.version,
      executorSource: type.source,
      executorConfig: type.config,
    }) as this;
  }

  executorVersion(version: string): this {
    return this.clone({ executorVersion: version }) as this;
  }

  executorSource(source: string): this {
    return this.clone({ executorSource: source }) as this;
  }

  version(version: string): this {
    return this.clone({ version }) as this;
  }

  description(description: string): this {
    return this.clone({ description }) as this;
  }

  inputSchema<T extends object>(): ServiceDefinition<T, TOutput>;
  inputSchema<S extends SchemaObject>(schema: S): ServiceDefinition<InferSchema<S>, TOutput>;
  inputSchema(schema?: SchemaObject): ServiceDefinition<object, TOutput> {
    if (!schema) return this.clone();
    const { ports, rules } = schemaToManifest(schema);
    return this.clone({ inputPorts: ports, inputRules: rules });
  }

  outputSchema<T extends object>(): ServiceDefinition<TInput, T>;
  outputSchema<S extends SchemaObject>(schema: S): ServiceDefinition<TInput, InferSchema<S>>;
  outputSchema(schema?: SchemaObject): ServiceDefinition<TInput, object> {
    if (!schema) return this.clone();
    const { ports } = schemaToManifest(schema);
    return this.clone({ outputPorts: ports });
  }

  get output(): Expressionify<TOutput> {
    return createExpressionProxy<TOutput>("source.output");
  }

  get input(): Expressionify<TInput> {
    return createExpressionProxy<TInput>("source.input");
  }

  withInput(bindings: InputBindings<TInput>): CompositionNode<TInput, TOutput>;
  withInput<TSource extends object>(connection: Connection<TSource, TInput>): CompositionNode<TInput, TOutput>;
  withInput(bindingsOrConnection: InputBindings<TInput> | Connection<object, TInput>): CompositionNode<TInput, TOutput> {
    if (isConnection(bindingsOrConnection)) {
      return new CompositionNode<TInput, TOutput>(this.ref, {
        bindings: bindingsFromConnection(bindingsOrConnection),
        incomingConditions: bindingsOrConnection.__conditions.map((condition) => ({
          expression: condition.expression,
          message: condition.message,
        })),
      });
    }
    return new CompositionNode<TInput, TOutput>(this.ref, {
      bindings: bindingsFromObject(bindingsOrConnection),
    });
  }

  withExecution(execution: ExecutionConfig): CompositionNode<TInput, TOutput> {
    return new CompositionNode<TInput, TOutput>(this.ref, { execution });
  }

  connect<TSource extends object>(
    define: (source: SourceContext<TSource>) => InputBindings<TInput>
  ): Connection<TSource, TInput> {
    return makeConnection(this.ref, define(createSourceContext<TSource>()));
  }

  when(condition: Expression<boolean>, message: string): CompositionNode<TInput, TOutput> {
    return this.withInput({} as InputBindings<TInput>).when(condition, message);
  }

  node(): SystemNodeManifest {
    return { kind: "service", service: this.ref };
  }

  toManifest(): ServiceManifest {
    return {
      name: this.ref,
      version: this.state.version,
      description: this.state.description,
      executor: {
        type: this.state.executorType,
        version: this.state.executorVersion,
        source: this.state.executorSource,
        config: this.state.executorConfig,
      },
      inputs: this.state.inputPorts,
      outputs: this.state.outputPorts,
      mappings: [],
      validations: Object.entries(this.state.inputRules).map(([field, rules]) => ({ field, rules })),
      config: {},
      execution: {},
    };
  }

  asReference(): ServiceReference<TInput, TOutput> {
    return this;
  }

  private clone<TNextInput extends object = TInput, TNextOutput extends object = TOutput>(
    patch: Partial<ServiceState> = {}
  ): ServiceDefinition<TNextInput, TNextOutput> {
    return new ServiceDefinition<TNextInput, TNextOutput>(this.ref, {
      ...this.state,
      inputPorts: [...this.state.inputPorts],
      outputPorts: [...this.state.outputPorts],
      inputRules: { ...this.state.inputRules },
      ...patch,
    });
  }
}

export function Service<TInput extends object = object, TOutput extends object = object>(
  ref: string
): ServiceDefinition<TInput, TOutput> {
  return new ServiceDefinition<TInput, TOutput>(ref);
}

export function toManifestNode(comp: Composition): SystemNodeManifest {
  switch (comp.kind) {
    case "service":
      return { kind: "service", service: comp.serviceRef };
    case "sequence":
      return { kind: "sequence", steps: comp.steps.map(toManifestNode) };
    case "parallel":
      return { kind: "parallel", branches: comp.branches.map(toManifestNode) };
  }
}

export function makeConnection<TSource extends object, TTarget extends object>(
  serviceRef: string | undefined,
  bindings: InputBindings<TTarget>,
  conditions: Array<{ expression: string; message: string }> = []
): Connection<TSource, TTarget> {
  const entries = Object.entries(bindings).flatMap(([target, value]) =>
    value === undefined ? [] : [{ target, expression: expressionToString(value) }]
  );
  return {
    __kind: "connection",
    __serviceRef: serviceRef,
    __bindings: entries,
    __conditions: conditions,
    when(condition: Expression<boolean>, message: string) {
      return makeConnection<TSource, TTarget>(serviceRef, bindings, [
        ...conditions,
        { expression: String(condition), message },
      ]);
    },
  };
}

function toComposition(
  value:
    | ServiceDefinition<object, object>
    | ServiceReference<object, object>
    | CompositionNode<object, object>
    | SequenceNode<object>
    | ParallelCompositionHolder
    | Connection<object, object>,
  condition?: Condition
): Composition {
  if (value instanceof CompositionNode) {
    return withCondition(value._composition, condition);
  }
  if (value instanceof SequenceNode) {
    const steps = value._composition.steps.map(cloneComposition);
    if (condition) {
      const target = findFirstServiceComposition(steps[0]);
      if (target) {
        target.incomingConditions = [
          ...target.incomingConditions,
          { expression: String(condition.when), message: condition.message },
        ];
      }
    }
    return { kind: "sequence", steps };
  }
  if (value instanceof ParallelCompositionHolder) {
    return value._composition;
  }
  if (isConnection(value)) {
    return {
      kind: "service",
      serviceRef: value.__serviceRef ?? "",
      bindings: bindingsFromConnection(value),
      incomingConditions: value.__conditions.map((condition) => ({
        expression: condition.expression,
        message: condition.message,
      })),
    };
  }
  if (isServiceReference(value)) {
    return withCondition({
      kind: "service",
      serviceRef: value.ref,
      bindings: {},
      incomingConditions: [],
    }, condition);
  }
  throw new Error("Invalid node provided to a `.then()` chain");
}

function withCondition(comp: ServiceComposition, condition?: Condition): ServiceComposition {
  return {
    kind: "service",
    serviceRef: comp.serviceRef,
    bindings: { ...comp.bindings },
    execution: comp.execution,
    incomingConditions: [
      ...comp.incomingConditions,
      ...(condition ? [{ expression: String(condition.when), message: condition.message }] : []),
    ],
  };
}

function bindingsFromObject<T extends object>(bindings: InputBindings<T>): Record<string, string> {
  return Object.fromEntries(
    Object.entries(bindings).flatMap(([name, value]) =>
      value === undefined ? [] : [[name, expressionToString(value)]]
    )
  );
}

function bindingsFromConnection(connection: Connection<object, object>): Record<string, string> {
  return Object.fromEntries(connection.__bindings.map((binding) => [binding.target, binding.expression]));
}

function splitSequence(comp: Composition): Composition[] {
  return comp.kind === "sequence" ? comp.steps : [comp];
}

function cloneComposition(comp: Composition): Composition {
  if (comp.kind === "service") {
    return {
      kind: "service",
      serviceRef: comp.serviceRef,
      bindings: { ...comp.bindings },
      execution: comp.execution,
      incomingConditions: comp.incomingConditions.map((condition) => ({ ...condition })),
    };
  }
  if (comp.kind === "sequence") {
    return { kind: "sequence", steps: comp.steps.map(cloneComposition) };
  }
  return { kind: "parallel", branches: comp.branches.map(cloneComposition) };
}

function findFirstServiceComposition(comp: Composition | undefined): ServiceComposition | null {
  if (!comp) return null;
  if (comp.kind === "service") return comp;
  if (comp.kind === "sequence") return findFirstServiceComposition(comp.steps[0]);
  for (const branch of comp.branches) {
    const found = findFirstServiceComposition(branch);
    if (found) return found;
  }
  return null;
}

function isConnection(value: unknown): value is Connection<object, object> {
  return Boolean(value && typeof value === "object" && "__kind" in value && value.__kind === "connection");
}

function isServiceReference(value: unknown): value is ServiceReference<object, object> {
  return Boolean(value && typeof value === "object" && "ref" in value && "withInput" in value);
}
