import type { Expression } from "./expr.js";

export interface ConnectorValidationManifest {
  expression: string;
  message: string;
}

/**
 * Creates a connector validation for use in `.then()` options.
 *
 * The expression should be built using `source` or `exec` helpers.
 *
 * @example
 * ```ts
 * validate(source.output("valid").eq(true), "Order validation failed")
 * // → { expression: "source.output.valid == true", message: "Order validation failed" }
 * ```
 */
export function validate(expression: Expression, message: string): ConnectorValidationManifest {
  return { expression: expression.toString(), message };
}

export interface ValidationRule {
  type: string;
  [key: string]: unknown;
}

/** Required field validation. */
export function required(): ValidationRule {
  return { type: "required" };
}

/** String field validation builder. */
export function string(): StringValidator {
  return new StringValidator();
}

/** Number field validation builder. */
export function number(): NumberValidator {
  return new NumberValidator();
}

/** Boolean field validation. */
export function boolean(): ValidationRule {
  return { type: "boolean" };
}

/** Array field validation builder. */
export function array(): ArrayValidator {
  return new ArrayValidator();
}

/** Object field validation builder. */
export function object(): ObjectValidator {
  return new ObjectValidator();
}

export class StringValidator {
  private rules: Record<string, unknown> = { type: "string" };

  /** Minimum string length. */
  min(length: number): this {
    this.rules.minLength = length;
    return this;
  }

  /** Maximum string length. */
  max(length: number): this {
    this.rules.maxLength = length;
    return this;
  }

  /** Regex pattern. */
  pattern(regex: string): this {
    this.rules.pattern = regex;
    return this;
  }

  /** Email format. */
  email(): this {
    this.rules.format = "email";
    return this;
  }

  /** UUID format. */
  uuid(): this {
    this.rules.format = "uuid";
    return this;
  }

  /** Convert to plain object. */
  toObject(): ValidationRule {
    return this.rules as ValidationRule;
  }
}

export class NumberValidator {
  private rules: Record<string, unknown> = { type: "number" };

  /** Minimum value. */
  min(value: number): this {
    this.rules.minimum = value;
    return this;
  }

  /** Maximum value. */
  max(value: number): this {
    this.rules.maximum = value;
    return this;
  }

  /** Exclusive minimum. */
  exclusiveMin(value: number): this {
    this.rules.exclusiveMinimum = value;
    return this;
  }

  /** Exclusive maximum. */
  exclusiveMax(value: number): this {
    this.rules.exclusiveMaximum = value;
    return this;
  }

  /** Must be integer. */
  integer(): this {
    this.rules.multipleOf = 1;
    return this;
  }

  /** Convert to plain object. */
  toObject(): ValidationRule {
    return this.rules as ValidationRule;
  }
}

export class ArrayValidator {
  private rules: Record<string, unknown> = { type: "array" };

  /** Minimum array length. */
  minItems(count: number): this {
    this.rules.minItems = count;
    return this;
  }

  /** Maximum array length. */
  maxItems(count: number): this {
    this.rules.maxItems = count;
    return this;
  }

  /** Unique items. */
  uniqueItems(): this {
    this.rules.uniqueItems = true;
    return this;
  }

  /** Convert to plain object. */
  toObject(): ValidationRule {
    return this.rules as ValidationRule;
  }
}

export class ObjectValidator {
  private rules: Record<string, unknown> = { type: "object" };

  /** Required properties. */
  required(props: string[]): this {
    this.rules.required = props;
    return this;
  }

  /** Additional properties. */
  additionalProperties(allowed: boolean): this {
    this.rules.additionalProperties = allowed;
    return this;
  }

  /** Convert to plain object. */
  toObject(): ValidationRule {
    return this.rules as ValidationRule;
  }
}