import { describe, expect, it, vi } from "vitest";

import { createResource } from "./suspense";

function flush() {
  return new Promise((resolve) => setTimeout(resolve, 0));
}

describe("createResource", () => {
  it("suspends until the value resolves", async () => {
    const resource = createResource(() => Promise.resolve("value"));

    expect(() => resource.read()).toThrow();
    await flush();
    expect(resource.read()).toBe("value");
  });

  it("rethrows the failure after settling", async () => {
    const resource = createResource(() => Promise.reject(new Error("boom")));

    expect(() => resource.read()).toThrow();
    await flush();
    expect(() => resource.read()).toThrow("boom");
  });

  it("forwards the abort signal", () => {
    const loader = vi.fn().mockResolvedValue(1);
    const controller = new AbortController();

    createResource(loader, controller.signal);

    expect(loader).toHaveBeenCalledWith(controller.signal);
  });
});
