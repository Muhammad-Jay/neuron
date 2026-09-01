import { describe, it, expect } from "vitest";
import { Service } from "../src/service.js";
import { string, number, boolean, list, record } from "../src/schema.js";
import { createExecutionContext, createSourceContext } from "../src/expression.js";

describe("Service", () => {
  it("creates a service with default executor type matching ref", () => {
    const svc = Service("validate-order");
    expect(svc.ref).toBe("validate-order");
    expect(svc.toManifest()).toEqual({
      name: "validate-order",
      version: undefined,
      description: undefined,
      executor: { type: "validate-order" },
      inputs: [],
      outputs: [],
      mappings: [],
      validations: [],
      config: {},
      execution: {},
    });
  });

  it("sets executor type, config, version, source", () => {
    const svc = Service("http-call")
      .executor("http", { timeout: 30 }, "2.0.0")
      .executorSource("official");

    expect(svc.toManifest().executor).toEqual({
      type: "http",
      version: "2.0.0",
      source: "official",
      config: { timeout: 30 },
    });
  });

  it("sets version and description", () => {
    const svc = Service("my-service")
      .version("2.0.0")
      .description("A test service");

    expect(svc.toManifest().name).toBe("my-service");
    expect(svc.toManifest().version).toBe("2.0.0");
    expect(svc.toManifest().description).toBe("A test service");
  });

  it("produces a node reference", () => {
    const svc = Service("my-service");
    expect(svc.node()).toEqual({ kind: "service", service: "my-service" });
  });
});

describe("Service schemas", () => {
  it("declares input schema with runtime validation rules", () => {
    const svc = Service("my-service").inputSchema({
      email: string().email().required(),
      age: number().min(18).max(120),
    });

    const manifest = svc.toManifest();
    expect(manifest.inputs).toEqual([
      {
        name: "email",
        type: "string",
        required: true,
        rules: { type: "string", format: "email", required: true },
      },
      {
        name: "age",
        type: "number",
        required: false,
        rules: { type: "number", minimum: 18, maximum: 120 },
      },
    ]);
  });

  it("declares output schema with runtime validation rules", () => {
    const svc = Service("my-service").outputSchema({
      verified: boolean(),
      tier: string(),
      items: list(),
      meta: record(),
    });

    const manifest = svc.toManifest();
    expect(manifest.outputs.map((o) => o.type)).toEqual([
      "boolean",
      "string",
      "array",
      "object",
    ]);
  });

  it("supports the type-only schema overload (no runtime rules)", () => {
    const svc = Service("my-service").inputSchema().outputSchema();
    const manifest = svc.toManifest();
    expect(manifest.inputs).toEqual([]);
    expect(manifest.outputs).toEqual([]);
  });
});

describe("Service.withInput()", () => {
  it("returns an immutable composition node with bindings", () => {
    const svc = Service("github.read")
      .inputSchema({
        owner: string().required(),
        repository: string().required(),
      })
      .outputSchema({
        content: string(),
        sha: string(),
      });

    const node = svc.withInput({
      owner: "Muhammad-Jay",
      repository: "neuron",
    });

    expect(node).toBeInstanceOf(Object);
    expect(node.serviceRef).toBe("github.read");
    expect(node.bindings).toEqual({
      owner: "'Muhammad-Jay'",
      repository: "'neuron'",
    });
  });

  it("converts expressions to expression strings", () => {
    const svc = Service("a").inputSchema({
      customerId: string().required(),
      email: string().required(),
    });
    const exec = createExecutionContext<{ id: string }>();
    const source = createSourceContext<{ email: string }>();

    const node = svc.withInput({
      customerId: exec.input.id,
      email: source.output.email,
    });

    expect(node.bindings.customerId).toBe("execution.input.id");
    expect(node.bindings.email).toBe("source.output.email");
  });

  it("does not mutate the original service definition", () => {
    const svc = Service("a").outputSchema({ out: string() });
    const before = JSON.stringify(svc.toManifest());
    svc.withInput({});
    expect(JSON.stringify(svc.toManifest())).toBe(before);
  });

  it("produces typed output expressions", () => {
    const svc = Service("a").outputSchema({ content: string(), sha: string() });
    expect(String(svc.output.content)).toBe("source.output.content");
    expect(String(svc.output.sha)).toBe("source.output.sha");
  });
});

describe("Service.withExecution()", () => {
  it("returns a composition node with execution config", () => {
    const svc = Service("a");
    const node = svc.withExecution({ timeout: "10s", retries: 2 });
    expect(node.executionConfig).toEqual({ timeout: "10s", retries: 2 });
  });
});

describe("Service.then()", () => {
  it("chains two services into a flat sequence", () => {
    const a = Service("a");
    const b = Service("b");

    const seq = a.withInput({}).then(b.withInput({}));
    expect(seq._composition.kind).toBe("sequence");
    expect(seq._composition.steps).toHaveLength(2);
  });

  it("chains three services via .then().then()", () => {
    const a = Service("a");
    const b = Service("b");
    const c = Service("c");

    const seq = a.withInput({}).then(b.withInput({})).then(c.withInput({}));
    expect(seq._composition.kind).toBe("sequence");
    expect(seq._composition.steps).toHaveLength(3);
  });

  it("attaches conditions via the second argument", () => {
    const a = Service("a").outputSchema({ valid: boolean() });
    const b = Service("b");

    const seq = a.withInput({}).then(b.withInput({}), {
      when: a.output.valid.equals(true),
      message: "failed",
    });

    const target = seq._composition.steps[1] as any;
    expect(target.incomingConditions).toEqual([
      { expression: "source.output.valid == true", message: "failed" },
    ]);
  });
});

describe("installable service package pattern", () => {
  it("exports a reusable typed service definition", () => {
    const githubRead = Service("github.read")
      .version("1.0.0")
      .inputSchema({
        owner: string().required(),
        repository: string().required(),
        path: string().required(),
      })
      .outputSchema({
        content: string(),
        sha: string(),
      })
      .executor("github.read");

    expect(githubRead.ref).toBe("github.read");
    expect(githubRead.toManifest().inputs).toHaveLength(3);
    expect(githubRead.toManifest().outputs).toHaveLength(2);
    expect(githubRead.toManifest().executor.type).toBe("github.read");
  });
});
