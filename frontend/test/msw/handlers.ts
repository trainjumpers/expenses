import { HttpResponse, http } from "msw";

export const testUser = {
  id: 1,
  name: "Test User",
  email: "test1@example.com",
};

export const testAccount = {
  id: 1,
  name: "HDFC Savings",
  bank_type: "hdfc",
  currency: "inr",
  balance: 1000,
  created_by: 1,
} as const;

export const testCategory = {
  id: 1,
  name: "Food",
  created_by: 1,
} as const;

export const loginRequests: Array<{ email: string; password: string }> = [];

export const signupRequests: Array<{
  name: string;
  email: string;
  password: string;
}> = [];

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

  http.post("*/api/v1/signup", async ({ request }) => {
    const body = (await request.json()) as {
      name: string;
      email: string;
      password: string;
    };
    signupRequests.push(body);

    if (body.email === "existing@example.com") {
      return HttpResponse.json(
        { error: "user already exists" },
        { status: 409 }
      );
    }

    return HttpResponse.json(
      {
        message: "User signed up successfully",
        data: { user: { ...testUser, name: body.name, email: body.email } },
      },
      { status: 201 }
    );
  }),

  http.get("*/api/v1/user", () =>
    HttpResponse.json({
      message: "User retrieved successfully",
      data: testUser,
    })
  ),

  http.get("*/api/v1/account", () =>
    HttpResponse.json({
      message: "Accounts retrieved successfully",
      data: [testAccount],
    })
  ),

  http.get("*/api/v1/category", () =>
    HttpResponse.json({
      message: "Categories retrieved successfully",
      data: [testCategory],
    })
  ),

  http.post("*/api/v1/refresh", () =>
    HttpResponse.json({
      message: "Token refreshed successfully",
      data: { user: testUser },
    })
  ),

  http.post("*/api/v1/logout", () =>
    HttpResponse.json({ message: "Logged out" })
  ),

  // Components that mount lists (profile dropdown, analytics panels) fetch
  // these on render; default them to empty so tests never hit the network.
  http.get("*/api/v1/rule", () =>
    HttpResponse.json({
      message: "Rules retrieved successfully",
      data: { rules: [], total: 0, page: 1, page_size: 5 },
    })
  ),

  http.get("*/api/v1/statement", () =>
    HttpResponse.json({
      message: "Statements retrieved successfully",
      data: { statements: [], total: 0, page: 1, page_size: 5 },
    })
  ),

  http.get("*/api/v1/transaction", () =>
    HttpResponse.json({
      message: "Transactions retrieved successfully",
      data: { transactions: [], total: 0, page: 1, page_size: 15 },
    })
  ),
];
