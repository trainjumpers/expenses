export const PORT = process.env.E2E_PORT ?? "3100";

export const API_BASE_URL =
  process.env.E2E_API_BASE_URL ?? "http://localhost:8080";

export const BASE_URL = process.env.E2E_BASE_URL ?? `http://localhost:${PORT}`;

export const E2E_USER = {
  name: "E2E Test User",
  email: "e2e@neurospend.test",
  password: "password123",
};

export const AUTH_STATE = "e2e/.auth/user.json";
