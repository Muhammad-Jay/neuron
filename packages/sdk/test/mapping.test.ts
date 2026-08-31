import { describe, it, expect } from "vitest";
import { map, inputMapping, outputMapping } from "../src/mapping.js";
import { source, exec } from "../src/expr.js";

describe("map", () => {
  it("creates a connector mapping", () => {
    const m = map("validation_data", source.output());
    expect(m).toEqual({
      target: "validation_data",
      expression: "source.output",
    });
  });

  it("creates a mapping with deep expression", () => {
    const m = map("customer_id", source.output("validation_data.order.customer_id"));
    expect(m).toEqual({
      target: "customer_id",
      expression: "source.output.validation_data.order.customer_id",
    });
  });

  it("creates a mapping from exec input", () => {
    const m = map("items", exec.input("order.items"));
    expect(m).toEqual({
      target: "items",
      expression: "execution.input.order.items",
    });
  });
});

describe("inputMapping", () => {
  it("creates a service input mapping", () => {
    const m = inputMapping("execution.input.order", "order");
    expect(m).toEqual({
      direction: "input",
      source: "execution.input.order",
      target: "order",
    });
  });
});

describe("outputMapping", () => {
  it("creates a service output mapping", () => {
    const m = outputMapping("result", "output_data");
    expect(m).toEqual({
      direction: "output",
      source: "result",
      target: "output_data",
    });
  });
});
