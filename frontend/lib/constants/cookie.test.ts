import { describe, expect, it } from "vitest";

import {
  ACCESS_TOKEN_EXPIRY,
  ACCESS_TOKEN_NAME,
  REFRESH_TOKEN_EXPIRY,
  REFRESH_TOKEN_NAME,
} from "./cookie";

describe("cookie constants", () => {
  it("defaults the token lifetimes", () => {
    expect(ACCESS_TOKEN_EXPIRY).toBe(60 * 60);
    expect(REFRESH_TOKEN_EXPIRY).toBe(90 * 24 * 60 * 60);
  });

  it("exposes the cookie names", () => {
    expect(ACCESS_TOKEN_NAME).toBe("access_token");
    expect(REFRESH_TOKEN_NAME).toBe("refresh_token");
  });
});
