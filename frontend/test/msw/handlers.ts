import { HttpResponse, http } from "msw";

export const testUser = {
  id: 1,
  name: "Test User",
  email: "test1@example.com",
};

export const loginRequests: Array<{ email: string; password: string }> = [];

export const handlers = [
  http.post("*/api/v1/login", async ({ request }) => {
    const body = (await request.json()) as { email: string; password: string };
    loginRequests.push(body);

    if (body.email === testUser.email && body.password === "password123") {
      return HttpResponse.json({
        message: "User logged in successfully",
        data: { user: testUser },
      });
    }

    return HttpResponse.json({ error: "invalid credentials" }, { status: 401 });
  }),
];
