export interface SystemManifest {
  apiVersion: "neuron/v1";
  kind: "System";
  metadata: {
    name: string;
    version: string;
    description?: string;
  };
  config?: ProjectConfig;
  services: ServiceManifest[];
  inputs?: PortManifest[];
  connectors: ConnectorManifest[];
  definition: SystemNodeManifest;
}

export interface ServiceManifest {
  name: string;
  version?: string;
  description?: string;

  executor: {
    name: string;
    version: string;
    registry: string;
  };

  inputs: PortManifest[];
  outputs: PortManifest[];

  config?: Record<string, unknown>;

  execution?: {
    mode?: string;
    timeout?: string;
    retries?: number;
    concurrency?: number;
    continueOnFail?: boolean;
  };
}

// ProjectConfig carries project-level configuration that N.O.R.E. needs
// for runtime assembly. It is the TS-side analog of the YAML neuron.yaml
// project fields, merged into the manifest so .neuron/manifest.json is the
// single source of truth.
export interface ProjectConfig {
  executorRegistries?: Array<{
    name?: string;
    url: string;
  }>;
  runtime?: {
    execution?: {
      mode?: string;
      timeout?: string;
    };
    workers?: {
      min?: number;
      max?: number;
    };
  };
  storage?: {
    provider?: string;
    directory?: string;
  };
  inspector?: {
    enabled?: boolean;
    address?: string;
  };
  variables?: Record<string, unknown>;
}

export interface PortManifest {
  name: string;
  type: "any" | "string" | "number" | "boolean" | "object" | "array";
  required: boolean;
  rules?: Record<string, unknown>;
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
