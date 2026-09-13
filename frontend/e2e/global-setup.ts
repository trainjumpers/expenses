import { request } from "@playwright/test";

import { API_BASE_URL, E2E_USER } from "./config";

export default async function globalSetup() {
  const context = await request.newContext({ baseURL: API_BASE_URL });
  try {
    const response = await context.post("/api/v1/signup", { data: E2E_USER });
    if (![201, 409].includes(response.status())) {
      throw new Error(
        `Could not ensure the E2E user exists (${response.status()}): ${await response.text()}. Is the backend running at ${API_BASE_URL}?`
      );
    }
  } finally {
    await context.dispose();
  }
}
