import { describe, it, expect } from "vitest";
import { source, exec } from "../src/expr.js";

describe("source", () => {
  it("produces source.output", () => {
    expect(String(source.output())).toBe("source.output");
  });

  it("produces source.output with path", () => {
    expect(String(source.output("valid"))).toBe("source.output.valid");
  });

  it("produces source.output with deep path", () => {
    expect(String(source.output("validation_data.order.customer_id"))).toBe(
      "source.output.validation_data.order.customer_id"
    );
  });

  it("produces source.input", () => {
    expect(String(source.input())).toBe("source.input");
  });

  it("produces source.input with path", () => {
    expect(String(source.input("order"))).toBe("source.input.order");
  });

  it("produces source.id", () => {
    expect(String(source.id)).toBe("source.id");
  });

  it("produces source.name", () => {
    expect(String(source.name)).toBe("source.name");
  });

  it("produces source.type", () => {
    expect(String(source.type)).toBe("source.type");
  });

  it("produces source.metadata.name", () => {
    expect(String(source.metadata.name)).toBe("source.metadata.name");
  });

  it("produces source.metadata.version", () => {
    expect(String(source.metadata.version)).toBe("source.metadata.version");
  });

  it("produces source.metadata.description", () => {
    expect(String(source.metadata.description)).toBe(
      "source.metadata.description"
    );
  });
});

describe("exec", () => {
  it("produces execution.input", () => {
    expect(String(exec.input())).toBe("execution.input");
  });

  it("produces execution.input with path", () => {
    expect(String(exec.input("order.items"))).toBe("execution.input.order.items");
  });

  it("produces execution.id", () => {
    expect(String(exec.id)).toBe("execution.id");
  });

  it("produces execution.correlationId", () => {
    expect(String(exec.correlationId)).toBe("execution.correlation_id");
  });

  it("produces execution.blueprint.id", () => {
    expect(String(exec.blueprint.id)).toBe("execution.blueprint.id");
  });

  it("produces execution.blueprint.name", () => {
    expect(String(exec.blueprint.name)).toBe("execution.blueprint.name");
  });

  it("produces execution.blueprint.version", () => {
    expect(String(exec.blueprint.version)).toBe("execution.blueprint.version");
  });
});

describe("expression comparison methods", () => {
  it("eq with boolean", () => {
    expect(String(source.output("valid").eq(true))).toBe(
      "source.output.valid == true"
    );
  });

  it("eq with string", () => {
    expect(String(source.output("status").eq("active"))).toBe(
      "source.output.status == 'active'"
    );
  });

  it("eq with number", () => {
    expect(String(source.output("count").eq(42))).toBe(
      "source.output.count == 42"
    );
  });

  it("neq", () => {
    expect(String(source.output("status").neq("failed"))).toBe(
      "source.output.status != 'failed'"
    );
  });

  it("gt", () => {
    expect(String(source.output("amount").gt(100))).toBe(
      "source.output.amount > 100"
    );
  });

  it("gte", () => {
    expect(String(source.output("amount").gte(100))).toBe(
      "source.output.amount >= 100"
    );
  });

  it("lt", () => {
    expect(String(source.output("amount").lt(100))).toBe(
      "source.output.amount < 100"
    );
  });

  it("lte", () => {
    expect(String(source.output("amount").lte(100))).toBe(
      "source.output.amount <= 100"
    );
  });

  it("and", () => {
    const left = source.output("valid").eq(true);
    const right = source.output("status").eq("ok");
    expect(String(left.and(right))).toBe(
      "(source.output.valid == true) && (source.output.status == 'ok')"
    );
  });

  it("or", () => {
    const left = source.output("valid").eq(true);
    const right = source.output("fallback").eq(true);
    expect(String(left.or(right))).toBe(
      "(source.output.valid == true) || (source.output.fallback == true)"
    );
  });

  it("chained comparisons", () => {
    const expr = source.output("amount").gte(0).and(source.output("amount").lte(1000));
    expect(String(expr)).toBe(
      "(source.output.amount >= 0) && (source.output.amount <= 1000)"
    );
  });
});

describe("expression is usable as string", () => {
  it("coerces to string via String()", () => {
    const e = source.output("valid");
    const str: string = String(e);
    expect(str).toBe("source.output.valid");
  });

  it("coerces to string in template literal", () => {
    const e = source.output("valid");
    expect(`${e}`).toBe("source.output.valid");
  });
});
