declare const __expressionBrand: unique symbol;

/**
 * A type-safe CEL expression string.
 *
 * `Expression` is a branded string type that prevents raw string literals
 * from being used where an expression is expected. Use the `source` and `exec`
 * helpers to build expressions that are compatible with N.O.R.E.'s CEL resolver.
 *
 * Expressions coerce to strings when used in string contexts (template literals,
 * concatenation, `String(expr)`), but also carry comparison methods for building
 * complex CEL expressions.
 *
 * @example
 * ```ts
 * source.output("valid")          // "source.output.valid"
 * exec.input("order.items")       // "execution.input.order.items"
 * source.output("valid").eq(true) // "source.output.valid == true"
 * ```
 */
export type Expression = string & { readonly [__expressionBrand]: never };

/**
 * An expression that carries comparison methods for building complex CEL
 * conditions. It remains usable as a plain string.
 */
export type ExpressionBuilder = Expression & {
  eq(value: unknown): Expression;
  neq(value: unknown): Expression;
  gt(value: unknown): Expression;
  gte(value: unknown): Expression;
  lt(value: unknown): Expression;
  lte(value: unknown): Expression;
  and(other: Expression): Expression;
  or(other: Expression): Expression;
  // Override string methods so the builder still concatenates as a string.
  toString(): string;
  valueOf(): string;
};

function comparison(base: string): ExpressionBuilder {
  const methods = {
    toString(): string {
      return base;
    },
    valueOf(): string {
      return base;
    },
    eq(value: unknown): Expression {
      return comparison(`${base} == ${serialize(value)}`);
    },
    neq(value: unknown): Expression {
      return comparison(`${base} != ${serialize(value)}`);
    },
    gt(value: unknown): Expression {
      return comparison(`${base} > ${serialize(value)}`);
    },
    gte(value: unknown): Expression {
      return comparison(`${base} >= ${serialize(value)}`);
    },
    lt(value: unknown): Expression {
      return comparison(`${base} < ${serialize(value)}`);
    },
    lte(value: unknown): Expression {
      return comparison(`${base} <= ${serialize(value)}`);
    },
    and(other: Expression): Expression {
      return comparison(`(${base}) && (${other})`);
    },
    or(other: Expression): Expression {
      return comparison(`(${base}) || (${other})`);
    },
  };

  return methods as unknown as ExpressionBuilder;
}

function serialize(value: unknown): string {
  if (typeof value === "string") return `'${value}'`;
  if (typeof value === "boolean" || typeof value === "number") return String(value);
  if (value === null || value === undefined) return "null";
  return JSON.stringify(value);
}

/**
 * Access the source service environment in CEL expressions.
 *
 * The `source` variable is available in connector mapping and validation
 * expressions. It represents the output of the service that a connector
 * originates from.
 *
 * @example
 * ```ts
 * source.output("valid")                              // "source.output.valid"
 * source.output("validation_data.order.customer_id")  // "source.output.validation_data.order.customer_id"
 * source.metadata.name                                // "source.metadata.name"
 * ```
 */
export const source = {
  get id(): ExpressionBuilder {
    return comparison("source.id");
  },

  get name(): ExpressionBuilder {
    return comparison("source.name");
  },

  get type(): ExpressionBuilder {
    return comparison("source.type");
  },

  input(path?: string): ExpressionBuilder {
    return comparison(path ? `source.input.${path}` : "source.input");
  },

  output(path?: string): ExpressionBuilder {
    return comparison(path ? `source.output.${path}` : "source.output");
  },

  metadata: {
    get id(): ExpressionBuilder {
      return comparison("source.metadata.id");
    },
    get name(): ExpressionBuilder {
      return comparison("source.metadata.name");
    },
    get description(): ExpressionBuilder {
      return comparison("source.metadata.description");
    },
    get version(): ExpressionBuilder {
      return comparison("source.metadata.version");
    },
  },
};

/**
 * Access the execution environment in CEL expressions.
 *
 * The `execution` variable is available in both connector expressions
 * and service configuration templates. It represents the current
 * execution context.
 *
 * @example
 * ```ts
 * exec.input("order.items")       // "execution.input.order.items"
 * exec.id                         // "execution.id"
 * exec.blueprint.name             // "execution.blueprint.name"
 * ```
 */
export const exec = {
  get id(): ExpressionBuilder {
    return comparison("execution.id");
  },

  get correlationId(): ExpressionBuilder {
    return comparison("execution.correlation_id");
  },

  input(path?: string): ExpressionBuilder {
    return comparison(path ? `execution.input.${path}` : "execution.input");
  },

  blueprint: {
    get id(): ExpressionBuilder {
      return comparison("execution.blueprint.id");
    },
    get name(): ExpressionBuilder {
      return comparison("execution.blueprint.name");
    },
    get version(): ExpressionBuilder {
      return comparison("execution.blueprint.version");
    },
  },
};
