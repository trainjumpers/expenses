import { toast } from "sonner";
import { describe, expect, it } from "vitest";

import { handleApiError } from "./toast";

describe("handleApiError", () => {
  it("asks the user to check the input on 400", () => {
    handleApiError(400, "transaction");
    expect(toast.error).toHaveBeenCalledWith(
      "Please check your input and try again",
      { id: "bad-user-input" }
    );
  });

  it("offers a login action on 401", () => {
    handleApiError(401, "transaction");
    expect(toast.error).toHaveBeenCalledWith(
      "Please login again",
      expect.objectContaining({
        id: "unauthorized",
        action: expect.objectContaining({ label: "Login" }),
      })
    );
  });

  it("names the missing resource on 404", () => {
    handleApiError(404, "account");
    expect(toast.warning).toHaveBeenCalledWith("Account does not exist");
  });

  it("names the conflicting resource on 409", () => {
    handleApiError(409, "category");
    expect(toast.error).toHaveBeenCalledWith("Category already exists");
  });

  it("reports oversized files on 413", () => {
    handleApiError(413, "statement");
    expect(toast.error).toHaveBeenCalledWith("File is too large", {
      id: "file-too-large",
    });
  });

  it("reports rate limiting on 429", () => {
    handleApiError(429, "statement");
    expect(toast.error).toHaveBeenCalledWith(
      "Too many attempts. Please try again later",
      { id: "rate-limited" }
    );
  });

  it("falls back to a generic message", () => {
    handleApiError(500, "transaction");
    expect(toast.error).toHaveBeenCalledWith(
      "Something went wrong. Contact support if the problem persists",
      { id: "generic-error" }
    );
  });
});
