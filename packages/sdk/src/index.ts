export {
  Service,
  ServiceDefinition,
  CompositionNode,
  SequenceNode,
  ParallelCompositionHolder,
} from "./service.js";
export type {
  Connection,
  ExecutionConfig,
  Composition,
  InputBindings,
  InputValue,
  ServiceReference,
} from "./service.js";

export { defineConfig, type NeuronConfig } from "./cli/index.js";
export { System, SystemDefinition } from "./system.js";
export { Parallel } from "./composition.js";
export { connect } from "./connection.js";

export {
  string,
  number,
  boolean,
  list,
  record,
} from "./schema.js";
export type {
  Schema,
  SchemaField,
  Infer,
  InferSchema,
  SchemaObject,
  StringField,
  NumberField,
  BooleanField,
  ListField,
  RecordField,
  FieldRules,
} from "./schema.js";

export {
  createExpressionProxy,
} from "./expression.js";
export type {
  Expression,
  Expressionify,
  ExpressionArray,
  SourceContext,
  ExecutionContext,
} from "./expression.js";

export type {
  SystemManifest,
  ServiceManifest,
  ConnectorManifest,
  ConnectorMappingManifest,
  ConnectorValidationManifest,
  PortManifest,
  SystemNodeManifest,
  ServiceMappingManifest,
  ServiceValidationManifest,
} from "./manifest.js";
