import { describe, it, expect } from "vitest";
import { Service, Input, Output } from "../src/service.js";
import { exec, source } from "../src/expr.js";

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

  it("sets executor type and config", () => {
    const svc = Service("http-call")
      .executor("http", { timeout: 30 })
      .executorVersion("1.0.0")
      .executorSource("official");

    expect(svc.toManifest().executor).toEqual({
      type: "http",
      version: "1.0.0",
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

  it("sets config", () => {
    const svc = Service("my-service").config({
      tax_rate: 0.0825,
      discount_rate: 0.1,
    });

    expect(svc.toManifest().config).toEqual({
      tax_rate: 0.0825,
      discount_rate: 0.1,
    });
  });

  it("sets execution settings", () => {
    const svc = Service("my-service").execution({
      timeout: "5s",
      retries: 3,
      mode: "wait",
    });

    expect(svc.toManifest().execution).toEqual({
      timeout: "5s",
      retries: 3,
      mode: "wait",
    });
  });

  it("binds input mappings from expressions", () => {
    const svc = Service("parse-order").input({
      validation_data: source.output("validation_data"),
    });

    const manifest = svc.toManifest();
    expect(manifest.inputs).toEqual([
      { name: "validation_data", type: "any", required: true },
    ]);
    expect(manifest.mappings).toEqual([
      {
        direction: "input",
        source: "source.output.validation_data",
        target: "validation_data",
      },
    ]);
  });

  it("binds multiple inputs", () => {
    const svc = Service("calculate").input({
      items: exec.input("order.items"),
      email: source.output("customer_data.email"),
      tier: source.output("customer_data.tier"),
    });

    const manifest = svc.toManifest();
    expect(manifest.inputs).toHaveLength(3);
    expect(manifest.mappings).toHaveLength(3);
    expect(manifest.mappings[0]).toEqual({
      direction: "input",
      source: "execution.input.order.items",
      target: "items",
    });
  });

  it("declares input and output ports explicitly", () => {
    const svc = Service("my-service")
      .inputs(
        Input("name", "string", true),
        Input("age", "number", true)
      )
      .outputs(Output("id", "string"), Output("created", "boolean"));

    const manifest = svc.toManifest();
    expect(manifest.inputs).toEqual([
      { name: "name", type: "string", required: true },
      { name: "age", type: "number", required: true },
    ]);
    expect(manifest.outputs).toEqual([
      { name: "id", type: "string", required: false },
      { name: "created", type: "boolean", required: false },
    ]);
  });

  it("declares output names", () => {
    const svc = Service("my-service").output("verified", "tier", "email");

    expect(svc.toManifest().outputs).toEqual([
      { name: "verified", type: "any", required: false },
      { name: "tier", type: "any", required: false },
      { name: "email", type: "any", required: false },
    ]);
  });

  it("adds input validations", () => {
    const svc = Service("my-service").validateInput("email", {
      type: "string",
      format: "email",
    });

    expect(svc.toManifest().validations).toEqual([
      { field: "email", rules: { type: "string", format: "email" } },
    ]);
  });

  it("produces a node reference", () => {
    const svc = Service("my-service");
    expect(svc.node()).toEqual({ kind: "service", service: "my-service" });
  });
});

describe("Service.then()", () => {
  it("chains two services into a sequence", () => {
    const a = Service("a");
    const b = Service("b");

    const seq = a.then(b);
    expect(seq.kind).toBe("sequence");
    expect(seq.steps).toEqual([
      { kind: "service", service: "a" },
      { kind: "service", service: "b" },
    ]);
  });

  it("chains a service with a parallel node", () => {
    const a = Service("a");
    const b = Service("b");
    const c = Service("c");

    const seq = a.then({ kind: "parallel", branches: [b.node(), c.node()] });
    expect(seq.steps).toEqual([
      { kind: "service", service: "a" },
      { kind: "parallel", branches: [{ kind: "service", service: "b" }, { kind: "service", service: "c" }] },
    ]);
  });

  it("chains three services via .then().then()", () => {
    const a = Service("a");
    const b = Service("b");
    const c = Service("c");

    const seq = a.then(b).then(c);
    expect(seq.kind).toBe("sequence");
    expect(seq.steps).toEqual([
      { kind: "service", service: "a" },
      { kind: "service", service: "b" },
      { kind: "service", service: "c" },
    ]);
  });

  it("attaches connector data from .then() options", () => {
    const a = Service("a");
    const b = Service("b");

    const seq = a.then(b, {
      mappings: [{ target: "data", expression: "source.output" }],
      validations: [{ expression: "source.output.valid == true", message: "Failed" }],
    });

    expect((seq as any)._connectors).toEqual([{
      from: "a",
      mappings: [{ target: "data", expression: "source.output" }],
      validations: [{ expression: "source.output.valid == true", message: "Failed" }],
    }]);
  });

  it("flattens nested sequences", () => {
    const a = Service("a");
    const b = Service("b");
    const c = Service("c");

    const seq = a.then(b).then(c);
    expect(seq.steps).toHaveLength(3);
  });
});

describe("typed service", () => {
  interface VerifyInput {
    customer_id: string;
    email: string;
  }

  interface VerifyOutput {
    verified: boolean;
    tier: string;
  }

  it("works with generic type parameters", () => {
    const svc = Service<VerifyInput, VerifyOutput>("customer.verify")
      .executor("http.post", { timeout: 30 })
      .input({
        customer_id: exec.input("order.customer_id"),
        email: exec.input("order.email"),
      })
      .output("verified", "tier")
      .config({ api_url: "https://api.example.com" });

    const manifest = svc.toManifest();
    expect(manifest.name).toBe("customer.verify");
    expect(manifest.executor.type).toBe("http.post");
    expect(manifest.inputs).toHaveLength(2);
    expect(manifest.outputs).toHaveLength(2);
  });
});

describe("installable service package pattern", () => {
  it("exports a reusable typed service definition", () => {
    // Simulating an installable package
    interface GitHubReadInput {
      owner: string;
      repo: string;
      path: string;
    }

    interface GitHubReadOutput {
      content: string;
      sha: string;
    }

    const GitHubRead = Service<GitHubReadInput, GitHubReadOutput>(
      "github.read"
    )
      .executor("http", { timeout: 10 })
      .inputs(Input("owner", "string", true), Input("repo", "string", true), Input("path", "string", true))
      .outputs(Output("content", "string"), Output("sha", "string"));

    expect(GitHubRead.ref).toBe("github.read");
    expect(GitHubRead.toManifest().inputs).toHaveLength(3);
    expect(GitHubRead.toManifest().outputs).toHaveLength(2);
  });
});
