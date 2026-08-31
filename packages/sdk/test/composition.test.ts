import { describe, it, expect } from "vitest";
import { Parallel } from "../src/composition.js";
import { Service } from "../src/service.js";

describe("Parallel", () => {
  it("creates a parallel node from services", () => {
    const a = Service("a");
    const b = Service("b");

    const node = Parallel(a, b);
    expect(node).toEqual({
      kind: "parallel",
      branches: [
        { kind: "service", service: "a" },
        { kind: "service", service: "b" },
      ],
    });
  });

  it("creates a parallel node from manifest nodes", () => {
    const node = Parallel(
      { kind: "service", service: "a" },
      { kind: "service", service: "b" }
    );
    expect(node).toEqual({
      kind: "parallel",
      branches: [
        { kind: "service", service: "a" },
        { kind: "service", service: "b" },
      ],
    });
  });

  it("creates a parallel node from mixed types", () => {
    const a = Service("a");
    const b = { kind: "service" as const, service: "b" };

    const node = Parallel(a, b);
    expect(node.branches).toHaveLength(2);
  });

  it("works with nested parallel", () => {
    const a = Service("a");
    const b = Service("b");
    const c = Service("c");

    const node = Parallel(a, Parallel(b, c));
    expect(node).toEqual({
      kind: "parallel",
      branches: [
        { kind: "service", service: "a" },
        {
          kind: "parallel",
          branches: [
            { kind: "service", service: "b" },
            { kind: "service", service: "c" },
          ],
        },
      ],
    });
  });

  it("works in a .then() chain", () => {
    const verify = Service("verify");
    const save = Service("save");
    const notify = Service("notify");

    const seq = verify.then(Parallel(save, notify));
    expect(seq.kind).toBe("sequence");
    expect(seq.steps).toEqual([
      { kind: "service", service: "verify" },
      {
        kind: "parallel",
        branches: [
          { kind: "service", service: "save" },
          { kind: "service", service: "notify" },
        ],
      },
    ]);
  });
});
