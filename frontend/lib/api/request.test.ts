import { server } from "@/test/msw/server";
import { HttpResponse, http } from "msw";
import { toast } from "sonner";
import { describe, expect, it, vi } from "vitest";

import { apiRequest } from "./request";

const endpoint = "http://localhost:8080/api/v1/thing";

describe("apiRequest", () => {
  it("unwraps the data field from a successful response", async () => {
    server.use(
      http.get(endpoint, () => HttpResponse.json({ data: { id: 7 } }))
    );

    await expect(
      apiRequest(endpoint, { method: "GET" }, "thing")
    ).resolves.toEqual({ id: 7 });
  });

  it("sends credentials with every request", async () => {
    let credentials: RequestCredentials | undefined;
    server.use(
      http.get(endpoint, ({ request }) => {
        credentials = request.credentials;
        return HttpResponse.json({ data: null });
      })
    );

    await apiRequest(endpoint, { method: "GET" }, "thing");

    expect(credentials).toBe("include");
  });

  it("lets a custom error handler take over", async () => {
    server.use(
      http.get(endpoint, () =>
        HttpResponse.json({ error: "invalid credentials" }, { status: 401 })
      )
    );
    const handler = vi.fn(() => true);

    await expect(
      apiRequest(endpoint, { method: "GET" }, "thing", [handler])
    ).rejects.toThrow("invalid credentials");

    expect(handler).toHaveBeenCalled();
    expect(toast.error).not.toHaveBeenCalled();
  });

  it("shows the default error toast when no handler claims the error", async () => {
    server.use(
      http.get(endpoint, () => HttpResponse.json({}, { status: 404 }))
    );

    await expect(
      apiRequest(endpoint, { method: "GET" }, "thing")
    ).rejects.toThrow("Request failed");

    expect(toast.warning).toHaveBeenCalledWith("Thing does not exist");
  });

  it("falls back to a generic message for non-JSON error bodies", async () => {
    server.use(
      http.get(endpoint, () => new HttpResponse("internal", { status: 500 }))
    );

    await expect(
      apiRequest(endpoint, { method: "GET" }, "thing")
    ).rejects.toThrow("Request failed");

    expect(toast.error).toHaveBeenCalledWith(
      "Something went wrong. Contact support if the problem persists",
      { id: "generic-error" }
    );
  });

  it("attaches the response status to the thrown error", async () => {
    server.use(
      http.get(endpoint, () =>
        HttpResponse.json({ error: "nope" }, { status: 409 })
      )
    );

    const error: unknown = await apiRequest(
      endpoint,
      { method: "GET" },
      "thing"
    ).catch((e: unknown) => e);

    expect(error).toMatchObject({
      message: "nope",
      status: 409,
      data: { error: "nope" },
    });
  });

  it("reports network failures with the provided message", async () => {
    server.use(http.get(endpoint, () => HttpResponse.error()));

    await expect(
      apiRequest(endpoint, { method: "GET" }, "thing", [], "Upload failed")
    ).rejects.toThrow();

    expect(toast.error).toHaveBeenCalledWith("Upload failed");
  });

  it("stays silent when the request is aborted", async () => {
    // Node's fetch and jsdom supply DOMException from different realms, so
    // simulate the same-realm rejection a browser produces.
    const abortError = new DOMException(
      "The operation was aborted.",
      "AbortError"
    );
    const fetchSpy = vi
      .spyOn(globalThis, "fetch")
      .mockRejectedValueOnce(abortError);

    await expect(apiRequest(endpoint, { method: "GET" }, "thing")).rejects.toBe(
      abortError
    );

    expect(toast.error).not.toHaveBeenCalled();
    fetchSpy.mockRestore();
  });
});
