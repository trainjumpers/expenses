import type { CreateTransaction } from "@/lib/models/transaction";
import { server } from "@/test/msw/server";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";

import {
  createTransaction,
  deleteTransaction,
  getAllTransactions,
  updateTransaction,
} from "./transaction";

const paginated = {
  transactions: [],
  total: 0,
  page: 1,
  page_size: 15,
};

describe("getAllTransactions", () => {
  it("skips empty params and keeps zero values", async () => {
    let search = "";
    server.use(
      http.get("*/api/v1/transaction", ({ request }) => {
        search = new URL(request.url).search;
        return HttpResponse.json({ message: "ok", data: paginated });
      })
    );

    await getAllTransactions({
      page: 1,
      search: "coffee",
      min_amount: 100,
      max_amount: 0,
      account_id: undefined,
      category_id: null as unknown as undefined,
      date_from: "",
    });

    expect(search).toContain("page=1");
    expect(search).toContain("search=coffee");
    expect(search).toContain("min_amount=100");
    expect(search).toContain("max_amount=0");
    expect(search).not.toContain("account_id");
    expect(search).not.toContain("category_id");
    expect(search).not.toContain("date_from");
  });

  it("omits the query string when there are no params", async () => {
    let url = "";
    server.use(
      http.get("*/api/v1/transaction", ({ request }) => {
        url = request.url;
        return HttpResponse.json({ message: "ok", data: paginated });
      })
    );

    await getAllTransactions();

    expect(new URL(url).search).toBe("");
  });
});

describe("createTransaction", () => {
  it("converts a date without a time part to ISO format", async () => {
    let body: CreateTransaction | undefined;
    server.use(
      http.post("*/api/v1/transaction", async ({ request }) => {
        body = (await request.json()) as CreateTransaction;
        return HttpResponse.json(
          { message: "created", data: { id: 5, ...body } },
          { status: 201 }
        );
      })
    );

    const created = await createTransaction({
      name: "Lunch",
      amount: 12,
      date: "2026-09-01",
      category_ids: [],
      account_id: 1,
    });

    expect(body?.date).toBe(new Date("2026-09-01").toISOString());
    expect(created.id).toBe(5);
  });

  it("keeps a date that already carries a time part", async () => {
    let body: CreateTransaction | undefined;
    server.use(
      http.post("*/api/v1/transaction", async ({ request }) => {
        body = (await request.json()) as CreateTransaction;
        return HttpResponse.json(
          { message: "created", data: { id: 6, ...body } },
          { status: 201 }
        );
      })
    );

    await createTransaction({
      name: "Lunch",
      amount: 12,
      date: "2026-09-01T10:00:00.000Z",
      category_ids: [],
      account_id: 1,
    });

    expect(body?.date).toBe("2026-09-01T10:00:00.000Z");
  });
});

describe("updateTransaction", () => {
  it("patches the transaction with the partial payload", async () => {
    let method = "";
    let body: Partial<CreateTransaction> | undefined;
    server.use(
      http.patch("*/api/v1/transaction/7", async ({ request }) => {
        method = request.method;
        body = (await request.json()) as Partial<CreateTransaction>;
        return HttpResponse.json({
          message: "updated",
          data: {
            id: 7,
            date: "2026-08-01T00:00:00.000Z",
            name: "Renamed",
            description: null,
            amount: 10,
            category_ids: [],
            account_id: 1,
            ...body,
          },
        });
      })
    );

    const updated = await updateTransaction(7, { name: "Renamed" });

    expect(method).toBe("PATCH");
    expect(body).toEqual({ name: "Renamed" });
    expect(updated.name).toBe("Renamed");
  });
});

describe("deleteTransaction", () => {
  it("resolves without a body", async () => {
    let method = "";
    server.use(
      http.delete("*/api/v1/transaction/7", ({ request }) => {
        method = request.method;
        return new HttpResponse(null, { status: 204 });
      })
    );

    await expect(deleteTransaction(7)).resolves.toBeUndefined();
    expect(method).toBe("DELETE");
  });
});
