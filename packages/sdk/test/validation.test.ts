import { describe, it, expect } from "vitest";
import {
  validate,
  required,
  string,
  number,
  boolean,
  array,
  object,
} from "../src/validation.js";
import { source } from "../src/expr.js";

describe("validate", () => {
  it("creates a connector validation", () => {
    const v = validate(source.output("valid").eq(true), "Order validation failed");
    expect(v).toEqual({
      expression: "source.output.valid == true",
      message: "Order validation failed",
    });
  });

  it("creates a validation with complex expression", () => {
    const v = validate(
      source.output("payment_intent.status").eq("requires_capture"),
      "Payment not authorized"
    );
    expect(v).toEqual({
      expression: "source.output.payment_intent.status == 'requires_capture'",
      message: "Payment not authorized",
    });
  });
});

describe("required", () => {
  it("returns required rule", () => {
    expect(required()).toEqual({ type: "required" });
  });
});

describe("string", () => {
  it("returns string type rule", () => {
    expect(string().toObject()).toEqual({ type: "string" });
  });

  it("builds min/max", () => {
    expect(string().min(1).max(100).toObject()).toEqual({
      type: "string",
      minLength: 1,
      maxLength: 100,
    });
  });

  it("builds pattern", () => {
    expect(string().pattern("^[a-z]+$").toObject()).toEqual({
      type: "string",
      pattern: "^[a-z]+$",
    });
  });

  it("builds email format", () => {
    expect(string().email().toObject()).toEqual({
      type: "string",
      format: "email",
    });
  });

  it("builds uuid format", () => {
    expect(string().uuid().toObject()).toEqual({
      type: "string",
      format: "uuid",
    });
  });

  it("chains multiple rules", () => {
    const rule = string().min(1).max(255).email();
    expect(rule.toObject()).toEqual({
      type: "string",
      minLength: 1,
      maxLength: 255,
      format: "email",
    });
  });
});

describe("number", () => {
  it("returns number type rule", () => {
    expect(number().toObject()).toEqual({ type: "number" });
  });

  it("builds min/max", () => {
    expect(number().min(0).max(100).toObject()).toEqual({
      type: "number",
      minimum: 0,
      maximum: 100,
    });
  });

  it("builds integer", () => {
    expect(number().integer().toObject()).toEqual({
      type: "number",
      multipleOf: 1,
    });
  });

  it("builds exclusive min/max", () => {
    expect(number().exclusiveMin(0).exclusiveMax(100).toObject()).toEqual({
      type: "number",
      exclusiveMinimum: 0,
      exclusiveMaximum: 100,
    });
  });
});

describe("boolean", () => {
  it("returns boolean type rule", () => {
    expect(boolean()).toEqual({ type: "boolean" });
  });
});

describe("array", () => {
  it("returns array type rule", () => {
    expect(array().toObject()).toEqual({ type: "array" });
  });

  it("builds minItems/maxItems", () => {
    expect(array().minItems(1).maxItems(10).toObject()).toEqual({
      type: "array",
      minItems: 1,
      maxItems: 10,
    });
  });

  it("builds uniqueItems", () => {
    expect(array().uniqueItems().toObject()).toEqual({
      type: "array",
      uniqueItems: true,
    });
  });
});

describe("object", () => {
  it("returns object type rule", () => {
    expect(object().toObject()).toEqual({ type: "object" });
  });

  it("builds required", () => {
    expect(object().required(["name", "email"]).toObject()).toEqual({
      type: "object",
      required: ["name", "email"],
    });
  });

  it("builds additionalProperties", () => {
    expect(object().additionalProperties(false).toObject()).toEqual({
      type: "object",
      additionalProperties: false,
    });
  });
});
