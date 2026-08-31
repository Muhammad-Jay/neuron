export interface SystemManifest {
  apiVersion: "neuron/v1";
  kind: "System";

  metadata: {
    name: string;
    version: string;
    description?: string;
  };

  services: ServiceManifest[];

  connectors: ConnectorManifest[];

  definition: SystemNodeManifest;
}

export interface ServiceManifest {
  name: string;
  version?: string;
  description?: string;

  executor: {
    type: string;
    version?: string;
    source?: string;
    config?: Record<string, unknown>;
  };

  inputs: PortManifest[];

  outputs: PortManifest[];

  mappings: ServiceMappingManifest[];

  validations: ServiceValidationManifest[];

  config: Record<string, unknown>;

  execution: {
    mode?: string;
    timeout?: string;
    retries?: number;
    concurrency?: number;
    continueOnFail?: boolean;
  };
}

export interface PortManifest {
  name: string;
  type: "any" | "string" | "number" | "boolean" | "object" | "array";
  required: boolean;
}

export interface ConnectorManifest {
  from: string;
  to: string;
  mappings: ConnectorMappingManifest[];
  validations: ConnectorValidationManifest[];
}

export interface ConnectorMappingManifest {
  target: string;
  expression: string;
}

export interface ConnectorValidationManifest {
  expression: string;
  message: string;
}

export interface ServiceMappingManifest {
  direction: "input" | "output";
  source: string;
  target: string;
}

export interface ServiceValidationManifest {
  field: string;
  rules: Record<string, unknown>;
}

export type SystemNodeManifest =
  | {
      kind: "service";
      service: string;
    }
  | {
      kind: "sequence";
      steps: SystemNodeManifest[];
    }
  | {
      kind: "parallel";
      branches: SystemNodeManifest[];
    };