import { describe, expect, it } from "vitest";

import {
  type HttpError,
  getErrorMessage,
  getErrorStatus,
  isAuthError,
  isConflictError,
  isForbiddenError,
  isHttpError,
  isNetworkError,
  isNotFoundError,
  isServerError,
  isStatementPasswordRequiredError,
  isValidationError,
} from "./errors";

function httpError(status: number, data?: unknown): HttpError {
  return Object.assign(new Error("boom"), {
    status,
    statusText: "Boom",
    data,
  });
}

describe("error type guards", () => {
  it("detects HTTP errors by status", () => {
    expect(isHttpError(httpError(400))).toBe(true);
    expect(isHttpError(new Error("plain"))).toBe(false);
    expect(isHttpError("nope")).toBe(false);
  });

  it("maps statuses to their specific error types", () => {
    expect(isValidationError(httpError(400))).toBe(true);
    expect(isAuthError(httpError(401))).toBe(true);
    expect(isForbiddenError(httpError(403))).toBe(true);
    expect(isNotFoundError(httpError(404))).toBe(true);
    expect(isConflictError(httpError(409))).toBe(true);
    expect(isServerError(httpError(500))).toBe(true);
    expect(isServerError(httpError(599))).toBe(true);

    expect(isAuthError(httpError(400))).toBe(false);
    expect(isServerError(httpError(400))).toBe(false);
  });

  it("detects network errors by code and missing status", () => {
    expect(
      isNetworkError(
        Object.assign(new Error("refused"), { code: "ECONNREFUSED" })
      )
    ).toBe(true);
    expect(
      isNetworkError(
        Object.assign(new Error("refused"), { code: "ECONNREFUSED", status: 0 })
      )
    ).toBe(false);
    expect(isNetworkError(new Error("plain"))).toBe(false);
  });
});

describe("getErrorMessage", () => {
  it("prefers the API message", () => {
    expect(getErrorMessage(httpError(400, { message: "Bad input" }))).toBe(
      "Bad input"
    );
  });

  it("falls back to the error message", () => {
    expect(getErrorMessage(httpError(500, {}))).toBe("boom");
  });

  it("falls back to the HTTP status line", () => {
    const error = httpError(500);
    error.message = "";
    expect(getErrorMessage(error)).toBe("HTTP 500: Boom");
  });

  it("handles plain errors, strings and unknown values", () => {
    expect(getErrorMessage(new Error("plain"))).toBe("plain");
    expect(getErrorMessage("raw")).toBe("raw");
    expect(getErrorMessage(42)).toBe("An unknown error occurred");
  });
});

describe("getErrorStatus", () => {
  it("returns the status for HTTP errors only", () => {
    expect(getErrorStatus(httpError(404))).toBe(404);
    expect(getErrorStatus(new Error("plain"))).toBeUndefined();
  });
});

describe("isStatementPasswordRequiredError", () => {
  it("matches the flag and the message", () => {
    expect(isStatementPasswordRequiredError({ isPasswordRequired: true })).toBe(
      true
    );
    expect(
      isStatementPasswordRequiredError({
        message: "statement password required",
      })
    ).toBe(true);
  });

  it("rejects other values", () => {
    expect(isStatementPasswordRequiredError({ message: "other" })).toBe(false);
    expect(isStatementPasswordRequiredError(null)).toBe(false);
    expect(isStatementPasswordRequiredError(undefined)).toBe(false);
  });
});
